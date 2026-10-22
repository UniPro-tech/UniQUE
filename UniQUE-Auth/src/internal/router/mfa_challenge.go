package router

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/UniPro-tech/UniQUE-Auth/internal/config"
	"github.com/UniPro-tech/UniQUE-Auth/internal/model"
	"github.com/UniPro-tech/UniQUE-Auth/internal/query"
	"github.com/UniPro-tech/UniQUE-Auth/internal/util"
	"github.com/oklog/ulid/v2"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const mfaMaxAttempts = 5

func challengeHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func issueMFAChallenge(q *query.Query, userID string, remember bool) (string, error) {
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(random[:])
	now := time.Now().UTC()
	err := q.MfaChallenge.Create(&model.MfaChallenge{
		TokenHash: challengeHash(token), UserID: userID, IsRemember: remember,
		CreatedAt: now, ExpiresAt: now.Add(5 * time.Minute),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

// A nil session with no error means authentication was rejected. Rejections
// commit the attempt counter; operational failures roll back the entire flow.
func completeMFA(db *gorm.DB, req AuthenticationRequest, cfg config.Config) (session *model.Session, user *model.User, jwt string, err error) {
	if len(req.MFAChallenge) != 43 {
		return
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)
		challenge, err := q.MfaChallenge.Clauses(clause.Locking{Strength: "UPDATE"}).Where(q.MfaChallenge.TokenHash.Eq(challengeHash(req.MFAChallenge))).First()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if challenge.ConsumedAt != nil || !now.Before(challenge.ExpiresAt) || challenge.Attempts >= mfaMaxAttempts {
			return nil
		}
		if _, err := q.MfaChallenge.Where(q.MfaChallenge.TokenHash.Eq(challenge.TokenHash)).Update(q.MfaChallenge.Attempts, challenge.Attempts+1); err != nil {
			return err
		}
		candidate, err := q.User.Where(q.User.ID.Eq(challenge.UserID), q.User.Status.Eq("active")).First()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		// The challenge, never the submitted username, selects the account.
		if (req.Username != "" && req.Username != candidate.CustomID) || !candidate.IsTotpEnabled || candidate.TotpSecret == nil || !totp.Validate(req.Code, *candidate.TotpSecret) {
			return nil
		}
		s := &model.Session{
			ID: ulid.Make().String(), UserID: candidate.ID,
			IPAddress: req.IPAddress, UserAgent: req.UserAgent,
			IsRemember: challenge.IsRemember, ExpiresAt: CalculateSessionExpiry(challenge.IsRemember),
			LastLoginAt: now, CreatedAt: now, UpdatedAt: now,
		}
		if _, err := q.MfaChallenge.Where(q.MfaChallenge.TokenHash.Eq(challenge.TokenHash)).Update(q.MfaChallenge.ConsumedAt, now); err != nil {
			return err
		}
		if err := q.Session.Create(s); err != nil {
			return err
		}
		token, err := util.GenerateSessionJWT(s.ID, candidate.ID, s.ExpiresAt, cfg)
		if err != nil {
			return err
		}
		session, user, jwt = s, candidate, token
		return nil
	})
	if err != nil {
		session, user, jwt = nil, nil, ""
	}
	return
}
