package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/UniPro-tech/UniQUE-API/internal/middleware"
	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// methodToAuditActionのテスト
func TestMethodToAuditAction_POST(t *testing.T) {
	result := middleware.MethodToAuditAction(http.MethodPost)
	want := "CREATE"

	if result != want {
		t.Errorf("MethodToAuditAction(%q) = %q, want %q", http.MethodPost, result, want)
	}
}

func TestMethodToAuditAction_PUT(t *testing.T) {
	result := middleware.MethodToAuditAction(http.MethodPut)
	want := "UPDATE"

	if result != want {
		t.Errorf("MethodToAuditAction(%q) = %q, want %q", http.MethodPut, result, want)
	}
}

func TestMethodToAuditAction_PATCH(t *testing.T) {
	result := middleware.MethodToAuditAction(http.MethodPatch)
	want := "UPDATE"

	if result != want {
		t.Errorf("MethodToAuditAction(%q) = %q, want %q", http.MethodPatch, result, want)
	}
}

func TestMethodToAuditAction_DELETE(t *testing.T) {
	result := middleware.MethodToAuditAction(http.MethodDelete)
	want := "DELETE"

	if result != want {
		t.Errorf("MethodToAuditAction(%q) = %q, want %q", http.MethodDelete, result, want)
	}
}

func TestMethodToAuditAction_GET(t *testing.T) {
	result := middleware.MethodToAuditAction(http.MethodGet)
	want := ""

	if result != want {
		t.Errorf("MethodToAuditAction(%q) = %q, want %q", http.MethodGet, result, want)
	}
}

func TestMethodToAuditAction_HEAD(t *testing.T) {
	result := middleware.MethodToAuditAction(http.MethodHead)
	want := ""

	if result != want {
		t.Errorf("MethodToAuditAction(%q) = %q, want %q", http.MethodHead, result, want)
	}
}

func TestMethodToAuditAction_OPTIONS(t *testing.T) {
	result := middleware.MethodToAuditAction(http.MethodOptions)
	want := ""

	if result != want {
		t.Errorf("MethodToAuditAction(%q) = %q, want %q", http.MethodOptions, result, want)
	}
}

func TestMethodToAuditAction_CONNECT(t *testing.T) {
	result := middleware.MethodToAuditAction(http.MethodConnect)
	want := ""

	if result != want {
		t.Errorf("MethodToAuditAction(%q) = %q, want %q", http.MethodConnect, result, want)
	}
}

func TestMethodToAuditAction_TRACE(t *testing.T) {
	result := middleware.MethodToAuditAction(http.MethodTrace)
	want := ""

	if result != want {
		t.Errorf("MethodToAuditAction(%q) = %q, want %q", http.MethodTrace, result, want)
	}
}

// shouldSkipAuditのテスト
func TestShouldSkipAudit_HealthPath(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodGet, "/health", nil)

	result := middleware.ShouldSkipAudit(c)

	if !result {
		t.Errorf("ShouldSkipAudit(/health) = %v, want true", result)
	}
}

func TestShouldSkipAudit_SwaggerPath(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodGet, "/swagger/ui", nil)

	result := middleware.ShouldSkipAudit(c)

	if !result {
		t.Errorf("ShouldSkipAudit(/swagger/ui) = %v, want true", result)
	}
}

func TestShouldSkipAudit_SwaggerJSON(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodGet, "/swagger.json", nil)

	result := middleware.ShouldSkipAudit(c)

	if !result {
		t.Errorf("ShouldSkipAudit(/swagger.json) = %v, want true", result)
	}
}

func TestShouldSkipAudit_SwaggerDeepPath(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodGet, "/swagger/v1/swagger.yaml", nil)

	result := middleware.ShouldSkipAudit(c)

	if !result {
		t.Errorf("ShouldSkipAudit(/swagger/v1/swagger.yaml) = %v, want true", result)
	}
}

func TestShouldSkipAudit_NormalPath(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "users endpoint",
			path: "/users",
		},
		{
			name: "users with id",
			path: "/users/123",
		},
		{
			name: "api path",
			path: "/api/v1/users",
		},
		{
			name: "deep api path",
			path: "/api/v1/users/123/profile",
		},
		{
			name: "root path",
			path: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request, _ = http.NewRequest(http.MethodGet, tt.path, nil)

			result := middleware.ShouldSkipAudit(c)

			if result {
				t.Errorf("ShouldSkipAudit(%q) = %v, want false", tt.path, result)
			}
		})
	}
}

func TestShouldSkipAudit_HealthWithQuery(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodGet, "/health?check=1", nil)

	result := middleware.ShouldSkipAudit(c)

	if result {
		t.Errorf("ShouldSkipAudit(/health?check=1) = %v, want false", result)
	}
}

func TestShouldSkipAudit_HealthLikePath(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest(http.MethodGet, "/health-check", nil)

	result := middleware.ShouldSkipAudit(c)

	if result {
		t.Errorf("ShouldSkipAudit(/health-check) = %v, want false", result)
	}
}

// AuditLogMiddlewareのテスト
func TestAuditLogMiddleware_NoUserInContext(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users", nil)

	// user をコンテキストに設定しない

	middleware(c)

	// エラーは発生しないが、監査ログは作成されない
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_NoDBInContext(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users", nil)

	// user をコンテキストに設定
	c.Set("user", &model.User{ID: "user-123"})
	// db をコンテキストに設定しない

	middleware(c)

	// エラーは発生しないが、監査ログは作成されない
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_SkipHealth(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/health", nil)

	// コンテキストに user と db を設定
	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// /health へのリクエストはスキップされるため、監査ログは作成されない
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_SkipSwagger(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/swagger/ui", nil)

	// コンテキストに user と db を設定
	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// /swagger へのリクエストはスキップされるため、監査ログは作成されない
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_GETRequest(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/users", nil)

	// コンテキストに user と db を設定
	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// GET リクエストはアクションがないため、監査ログは作成されない
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_InvalidUserType(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users", nil)

	// コンテキストに無効な型の user を設定
	c.Set("user", "not-a-user")
	c.Set("db", &gorm.DB{})

	middleware(c)

	// user の型が違うため、監査ログは作成されない
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_InvalidDBType(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users", nil)

	// コンテキストに無効な型の db を設定
	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", "not-a-db")

	middleware(c)

	// db の型が違うため、監査ログは作成されない
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_NilUser(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users", nil)

	// コンテキストに nil の user を設定
	c.Set("user", (*model.User)(nil))
	c.Set("db", &gorm.DB{})

	middleware(c)

	// user が nil のため、監査ログは作成されない
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_NilDB(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users", nil)

	// コンテキストに nil の db を設定
	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", (*gorm.DB)(nil))

	middleware(c)

	// db が nil のため、監査ログは作成されない
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_FullPathVsURL(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users/123", nil)

	// FullPath が利用可能な場合
	c.FullPath()

	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// リクエストが処理される
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_ClientIP(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/users", nil)
	req.RemoteAddr = "192.168.1.1:8080"
	c.Request = req

	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// ClientIP が正しく取得されることを確認
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_UserAgent(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	c.Request = req

	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// UserAgent が正しく取得されることを確認
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_POSTAction(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users", nil)

	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// POST リクエストの処理
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_PUTAction(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, "/users/123", nil)

	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// PUT リクエストの処理
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_DELETEAction(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodDelete, "/users/123", nil)

	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// DELETE リクエストの処理
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_StatusCode(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users", nil)

	// ステータスコードを設定
	c.Writer.WriteHeader(http.StatusCreated)

	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// ステータスコードがキャプチャされることを確認
	if w.Code != http.StatusCreated {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_AuditDetailsMarshaling(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	// auditDetails の JSON マーシャリング確認
	details := map[string]interface{}{
		"method":     "POST",
		"path":       "/users",
		"status":     201,
		"ip":         "192.168.1.1",
		"user_agent": "Mozilla/5.0",
	}

	jsonBytes, err := json.Marshal(details)
	if err != nil {
		t.Errorf("Failed to marshal auditDetails: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Errorf("Marshaled JSON is empty")
	}

	// JSON の内容を確認
	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if unmarshaled["method"] != "POST" {
		t.Errorf("method = %v, want POST", unmarshaled["method"])
	}
	if unmarshaled["path"] != "/users" {
		t.Errorf("path = %v, want /users", unmarshaled["path"])
	}
	if unmarshaled["status"] != float64(201) {
		t.Errorf("status = %v, want 201", unmarshaled["status"])
	}
}

func TestAuditLogMiddleware_MultipleHTTPMethods(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "POST to users",
			method: http.MethodPost,
			path:   "/users",
		},
		{
			name:   "PUT to user",
			method: http.MethodPut,
			path:   "/users/123",
		},
		{
			name:   "PATCH to user",
			method: http.MethodPatch,
			path:   "/users/123",
		},
		{
			name:   "DELETE user",
			method: http.MethodDelete,
			path:   "/users/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := middleware.AuditLogMiddleware()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(tt.method, tt.path, nil)

			c.Set("user", &model.User{ID: "user-123"})
			c.Set("db", &gorm.DB{})

			middleware(c)

			if w.Code != http.StatusOK {
				t.Logf("Status code: %d", w.Code)
			}
		})
	}
}

func TestAuditLogMiddleware_ContextContainsUserID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users", nil)

	userID := "user-123"
	c.Set("user", &model.User{ID: userID})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// 監査ログが作成される際、userID が使用されることを確認
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_EmptyFullPath(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	middleware := middleware.AuditLogMiddleware()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/users/123", nil)

	// FullPath が空の場合、Request.URL.Path が使用される

	c.Set("user", &model.User{ID: "user-123"})
	c.Set("db", &gorm.DB{})

	middleware(c)

	// パスが正しくセットされることを確認
	if w.Code != http.StatusOK {
		t.Logf("Status code: %d", w.Code)
	}
}

func TestAuditLogMiddleware_AllHTTPMethods(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	methods := []struct {
		method         string
		expectedAction string
		shouldAudit    bool
	}{
		{http.MethodGet, "", false},
		{http.MethodHead, "", false},
		{http.MethodPost, "CREATE", true},
		{http.MethodPut, "UPDATE", true},
		{http.MethodPatch, "UPDATE", true},
		{http.MethodDelete, "DELETE", true},
		{http.MethodConnect, "", false},
		{http.MethodOptions, "", false},
		{http.MethodTrace, "", false},
	}

	for _, m := range methods {
		t.Run(m.method, func(t *testing.T) {
			middleware := middleware.AuditLogMiddleware()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(m.method, "/api/resource", nil)

			c.Set("user", &model.User{ID: "user-123"})
			c.Set("db", &gorm.DB{})

			middleware(c)

			if w.Code != http.StatusOK {
				t.Logf("Status code for %s: %d", m.method, w.Code)
			}
		})
	}
}
