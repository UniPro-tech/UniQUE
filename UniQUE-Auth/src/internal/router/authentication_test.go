package router

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/UniPro-tech/UniQUE-Auth/internal/model"
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/UniPro-tech/UniQUE-Auth/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"github.com/pquerna/otp/totp"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Use an isolated, disposable MySQL database: this exercises real InnoDB locks.
func TestAuthenticationMFA(t *testing.T) {
	dsn := os.Getenv("MFA_TEST_DSN")
	if dsn == "" {
		t.Skip("set MFA_TEST_DSN to an isolated MySQL test database")
	}
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn, DisableDatetimePrecision: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Session{}, &model.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasTable(&model.MfaChallenge{}) {
		sql, err := os.ReadFile("../../../../UniQUE-DB/migrations/22_mfa_challenges.up.sql")
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(string(sql)).Error; err != nil {
			t.Fatal(err)
		}
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{KeyPairs: []config.KeyPairConfig{{PublicKey: key.PublicKey, PrivateKey: *key}}}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("db", db); c.Set("config", &cfg) })
	r.POST("/internal/authentication", AuthenticationPost)
	r.GET("/internal/session_verify", SessionVerifyGet)
	r.GET("/.well-known/jwks.json", WellKnownJWKS)
	post := func(req AuthenticationRequest) (int, AuthenticationResponse) {
		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("POST", "/internal/authentication", bytes.NewReader(body)))
		var response AuthenticationResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		return w.Code, response
	}
	hash, err := util.HashPassword("local-test-password")
	if err != nil {
		t.Fatal(err)
	}
	newUser := func(enabled bool) model.User {
		secret, err := totp.Generate(totp.GenerateOpts{Issuer: "local-test", AccountName: "local-test"})
		if err != nil {
			t.Fatal(err)
		}
		s := secret.Secret()
		u := model.User{ID: ulid.Make().String(), CustomID: ulid.Make().String(), PasswordHash: hash, Status: "active", IsTotpEnabled: enabled, TotpSecret: &s}
		if err := db.Create(&u).Error; err != nil {
			t.Fatal(err)
		}
		return u
	}
	u := newUser(true)
	code := func() string {
		c, err := totp.GenerateCode(*u.TotpSecret, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	count := func() int64 {
		var n int64
		if err := db.Model(&model.Session{}).Count(&n).Error; err != nil {
			t.Fatal(err)
		}
		return n
	}
	challenge := func(remember bool) string {
		before := count()
		status, res := post(AuthenticationRequest{Type: "password", Username: u.CustomID, Password: "local-test-password", IsRemember: remember})
		if status != 200 || !res.RequireMFA || len(res.MFAChallenge) != 43 || res.SessionJWT != "" || count() != before {
			t.Fatal("password step did not produce challenge only")
		}
		return res.MFAChallenge
	}
	reject := func(req AuthenticationRequest) {
		t.Helper()
		before := count()
		status, res := post(req)
		if status != 401 || res.SessionJWT != "" || res.MFAChallenge != "" || count() != before {
			t.Fatalf("expected rejection without session; status=%d", status)
		}
	}
	t.Run("T02_T06_missing_forged_challenge", func(t *testing.T) {
		for _, token := range []string{"", strings.Repeat("x", 43)} {
			reject(AuthenticationRequest{Type: "totp", Username: u.CustomID, Code: code(), MFAChallenge: token})
		}
	})
	t.Run("T04_T16_bad_password_unknown_user", func(t *testing.T) {
		reject(AuthenticationRequest{Type: "password", Username: u.CustomID, Password: "wrong"})
		reject(AuthenticationRequest{Type: "password", Username: "missing", Password: "wrong"})
		reject(AuthenticationRequest{Type: "totp", Username: "missing", Code: code()})
	})
	t.Run("T03_T05_T07_invalid_bound_expired", func(t *testing.T) {
		token := challenge(false)
		reject(AuthenticationRequest{Type: "totp", MFAChallenge: token, Code: "invalid"})
		reject(AuthenticationRequest{Type: "totp", Username: newUser(true).CustomID, MFAChallenge: token, Code: code()})
		q := query.Use(db)
		if _, err := q.MfaChallenge.Where(q.MfaChallenge.TokenHash.Eq(challengeHash(token))).Update(q.MfaChallenge.ExpiresAt, time.Now().Add(-time.Minute)); err != nil {
			t.Fatal(err)
		}
		reject(AuthenticationRequest{Type: "totp", MFAChallenge: token, Code: code()})
	})
	t.Run("T10_attempt_limit", func(t *testing.T) {
		token := challenge(false)
		for i := 0; i < mfaMaxAttempts; i++ {
			reject(AuthenticationRequest{Type: "totp", MFAChallenge: token, Code: "invalid"})
		}
		reject(AuthenticationRequest{Type: "totp", MFAChallenge: token, Code: code()})
	})
	t.Run("T01_T08_success_replay_remember", func(t *testing.T) {
		for _, remember := range []bool{false, true} {
			token := challenge(remember)
			before := count()
			req := AuthenticationRequest{Type: "totp", MFAChallenge: token, Code: code(), IsRemember: !remember}
			status, res := post(req)
			if status != 200 || res.SessionJWT == "" || count() != before+1 {
				t.Fatal("MFA success did not create exactly one session")
			}
			var s model.Session
			if err := db.Where("user_id = ?", u.ID).Order("id DESC").First(&s).Error; err != nil {
				t.Fatal(err)
			}
			if s.IsRemember != remember {
				t.Fatal("remember was not bound to password step")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("GET", "/internal/session_verify?jti="+s.ID, nil))
			var verified SessionVerifyResponse
			_ = json.Unmarshal(w.Body.Bytes(), &verified)
			if !verified.Valid || verified.UserID != u.ID {
				t.Fatal("session verification failed")
			}
			reject(req)
		}
	})
	t.Run("T09_concurrent_consumption", func(t *testing.T) {
		token := challenge(false)
		before := count()
		req := AuthenticationRequest{Type: "totp", MFAChallenge: token, Code: code()}
		var wg sync.WaitGroup
		results := make(chan int, 8)
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); status, _ := post(req); results <- status }()
		}
		wg.Wait()
		close(results)
		success := 0
		for status := range results {
			if status == 200 {
				success++
			} else if status != 401 {
				t.Fatalf("unexpected concurrent status %d", status)
			}
		}
		if success != 1 || count() != before+1 {
			t.Fatal("concurrent challenge created multiple sessions")
		}
	})
	t.Run("signing_failure_rolls_back_consumption_and_session", func(t *testing.T) {
		token := challenge(false)
		before := count()
		req := AuthenticationRequest{Type: "totp", MFAChallenge: token, Code: code()}
		s, _, jwt, err := completeMFA(db, req, config.Config{})
		if err == nil || s != nil || jwt != "" || count() != before {
			t.Fatal("signing failure did not roll back session creation")
		}
		q := query.Use(db)
		stored, err := q.MfaChallenge.Where(q.MfaChallenge.TokenHash.Eq(challengeHash(token))).First()
		if err != nil {
			t.Fatal(err)
		}
		if stored.ConsumedAt != nil || stored.Attempts != 0 {
			t.Fatal("signing failure consumed the challenge")
		}
		status, res := post(req)
		if status != 200 || res.SessionJWT == "" {
			t.Fatal("rolled-back challenge could not be retried")
		}
	})
	t.Run("user_state_is_rechecked", func(t *testing.T) {
		for _, state := range []string{"inactive", "deleted", "secret_removed"} {
			t.Run(state, func(t *testing.T) {
				candidate := newUser(true)
				status, response := post(AuthenticationRequest{Type: "password", Username: candidate.CustomID, Password: "local-test-password"})
				if status != 200 || response.MFAChallenge == "" {
					t.Fatal("missing challenge")
				}
				validCode, err := totp.GenerateCode(*candidate.TotpSecret, time.Now())
				if err != nil {
					t.Fatal(err)
				}
				var result *gorm.DB
				switch state {
				case "inactive":
					result = db.Model(&candidate).Update("status", "inactive")
				case "deleted":
					result = db.Delete(&candidate)
				case "secret_removed":
					result = db.Model(&candidate).Update("totp_secret", nil)
				}
				if result.Error != nil {
					t.Fatal(result.Error)
				}
				reject(AuthenticationRequest{Type: "totp", MFAChallenge: response.MFAChallenge, Code: validCode})
			})
		}
	})
	t.Run("T11_T12_no_totp", func(t *testing.T) {
		plain := newUser(false)
		status, res := post(AuthenticationRequest{Type: "password", Username: plain.CustomID, Password: "local-test-password"})
		if status != 200 || res.SessionJWT == "" || res.RequireMFA {
			t.Fatal("password-only login failed")
		}
		reject(AuthenticationRequest{Type: "totp", Username: plain.CustomID, Code: code()})
		token := challenge(false)
		if err := db.Model(&u).Update("is_totp_enabled", false).Error; err != nil {
			t.Fatal(err)
		}
		reject(AuthenticationRequest{Type: "totp", MFAChallenge: token, Code: code()})
	})
	// Optional local-only bridge for the Front Action integration test. This
	// fixture exists only in the test binary; never expose its port publicly.
	if os.Getenv("MFA_TEST_HTTP") == "1" {
		fixture := newUser(true)
		r.GET("/test/fixture", func(c *gin.Context) {
			current, err := totp.GenerateCode(*fixture.TotpSecret, time.Now())
			if err != nil {
				c.Status(500)
				return
			}
			c.JSON(200, gin.H{"username": fixture.CustomID, "password": "local-test-password", "code": current})
		})
		r.GET("/test/sessions", func(c *gin.Context) {
			var n int64
			db.Model(&model.Session{}).Where("user_id = ?", fixture.ID).Count(&n)
			c.JSON(200, gin.H{"count": n})
		})
		server := &http.Server{Addr: ":8080", Handler: r, ReadHeaderTimeout: 5 * time.Second}
		time.AfterFunc(3*time.Minute, func() { _ = server.Close() })
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Fatal(err)
		}
	}
}
