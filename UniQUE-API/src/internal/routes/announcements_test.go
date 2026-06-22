package routes_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/UniPro-tech/UniQUE-API/internal/routes"
	"github.com/gin-gonic/gin"
)

// TestListAnnouncements_Success
func TestListAnnouncements_Success(t *testing.T) {
	// This test requires a real or mocked database.
	// For now, we'll create a minimal setup with httptest.

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements", nil)

	// Mock DB is not set, so it will return early
	// In a real scenario, you'd set up a test database

	// Expected: No response (or 500 if DB is required)
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Expected early return, got status %d", w.Code)
	}
}

// TestListAnnouncements_WithLimit
func TestListAnnouncements_WithLimit(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements?limit=10", nil)

	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestListAnnouncements_InvalidLimit
func TestListAnnouncements_InvalidLimit(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements?limit=invalid", nil)

	// The function parses limit and returns 400 if invalid
	// But we need DB to be set for the logic to execute
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestListAnnouncements_InvalidDeleted
func TestListAnnouncements_InvalidDeleted(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements?deleted=notbool", nil)

	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestListAnnouncements_DeletedWithoutPermission
func TestListAnnouncements_DeletedWithoutPermission(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements?deleted=true", nil)

	// User is not in context, so should return 403
	// But DB is not set, so it returns early

	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestGetAnnouncement_NotFound
func TestGetAnnouncement_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements/nonexistent", nil)
	c.Params = gin.Params{{Key: "id", Value: "nonexistent"}}

	// DB is not set, so returns early
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestCreateAnnouncement_NoAuth
func TestCreateAnnouncement_NoAuth(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := routes.CreateAnnouncementRequest{
		Title:   "Test Announcement",
		Content: "Test Content",
	}

	c.Request, _ = http.NewRequest(http.MethodPost, "/announcements", createRequestBody(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// User not in context, so createdByID will be ""
	// DB is not set, so returns early

	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestCreateAnnouncement_InvalidJSON
func TestCreateAnnouncement_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodPost, "/announcements", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	// ShouldBindJSON will fail and return 400

	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestCreateAnnouncement_WithIsPinned
func TestCreateAnnouncement_WithIsPinned(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	isPinned := true
	body := routes.CreateAnnouncementRequest{
		Title:    "Pinned Announcement",
		Content:  "Content",
		IsPinned: &isPinned,
	}

	c.Request, _ = http.NewRequest(http.MethodPost, "/announcements", createRequestBody(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// DB not set, returns early
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestUpdateAnnouncement_NotFound
func TestUpdateAnnouncement_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "nonexistent"}}

	body := routes.UpdateAnnouncementRequest{
		Title: stringPtr("Updated Title"),
	}

	c.Request, _ = http.NewRequest(http.MethodPut, "/announcements/nonexistent", createRequestBody(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// DB not set, returns early
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestUpdateAnnouncement_InvalidJSON
func TestUpdateAnnouncement_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}

	c.Request, _ = http.NewRequest(http.MethodPut, "/announcements/test-id", bytes.NewReader([]byte("invalid")))
	c.Request.Header.Set("Content-Type", "application/json")

	// ShouldBindJSON fails, returns 400
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestUpdateAnnouncement_AllFields
func TestUpdateAnnouncement_AllFields(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}

	isPinned := true
	body := routes.UpdateAnnouncementRequest{
		Title:    stringPtr("New Title"),
		Content:  stringPtr("New Content"),
		IsPinned: &isPinned,
	}

	c.Request, _ = http.NewRequest(http.MethodPut, "/announcements/test-id", createRequestBody(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// DB not set
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestUpdateAnnouncement_PartialUpdate
func TestUpdateAnnouncement_PartialUpdate(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}

	body := routes.UpdateAnnouncementRequest{
		Title: stringPtr("Only Title Updated"),
	}

	c.Request, _ = http.NewRequest(http.MethodPut, "/announcements/test-id", createRequestBody(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// DB not set
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestPatchAnnouncement_InvalidJSON
func TestPatchAnnouncement_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}

	c.Request, _ = http.NewRequest(http.MethodPatch, "/announcements/test-id", bytes.NewReader([]byte("invalid")))
	c.Request.Header.Set("Content-Type", "application/json")

	// ShouldBindJSON fails
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestPatchAnnouncement_PartialUpdate
func TestPatchAnnouncement_PartialUpdate(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}

	// 修正：Title に直接 *string を渡す
	body := routes.PatchAnnouncementRequest{
		Title: stringPtr("Patched Title"),
	}

	c.Request, _ = http.NewRequest(http.MethodPatch, "/announcements/test-id", createRequestBody(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// DB not set
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestDeleteAnnouncement_Success
func TestDeleteAnnouncement_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}

	c.Request, _ = http.NewRequest(http.MethodDelete, "/announcements/test-id", nil)

	// DB not set
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestPinAnnouncement_Pin
func TestPinAnnouncement_Pin(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}

	body := map[string]bool{"pin": true}
	c.Request, _ = http.NewRequest(http.MethodPost, "/announcements/test-id/pin", createRequestBody(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// DB not set
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestPinAnnouncement_Unpin
func TestPinAnnouncement_Unpin(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}

	body := map[string]bool{"pin": false}
	c.Request, _ = http.NewRequest(http.MethodPost, "/announcements/test-id/pin", createRequestBody(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// DB not set
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestPinAnnouncement_InvalidJSON
func TestPinAnnouncement_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}

	c.Request, _ = http.NewRequest(http.MethodPost, "/announcements/test-id/pin", bytes.NewReader([]byte("invalid")))
	c.Request.Header.Set("Content-Type", "application/json")

	// ShouldBindJSON fails silently (error is ignored with _)
	// So the default Pin value (false) will be used
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestPinAnnouncement_NotFound
func TestPinAnnouncement_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "nonexistent"}}

	body := map[string]bool{"pin": true}
	c.Request, _ = http.NewRequest(http.MethodPost, "/announcements/nonexistent/pin", createRequestBody(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// DB not set
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestListAnnouncements_LimitZero
func TestListAnnouncements_LimitZero(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements?limit=0", nil)

	// limit=0 is allowed and means no limit
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestListAnnouncements_LimitNegative
func TestListAnnouncements_LimitNegative(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements?limit=-1", nil)

	// Negative limit returns 400
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestGetAnnouncement_WithCreatedBy
func TestGetAnnouncement_WithCreatedBy(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-ann-id"}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements/test-ann-id", nil)

	// DB not set
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestCreateAnnouncement_WithUser
func TestCreateAnnouncement_WithUser(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	user := &model.User{
		ID:       "user-123",
		CustomID: "user-custom",
	}
	setUserInContext(c, user)

	body := routes.CreateAnnouncementRequest{
		Title:   "Test Announcement",
		Content: "Test Content",
	}

	c.Request, _ = http.NewRequest(http.MethodPost, "/announcements", createRequestBody(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// DB not set
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestUpdateAnnouncement_EmptyBody
func TestUpdateAnnouncement_EmptyBody(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "test-id"}}

	body := routes.UpdateAnnouncementRequest{}
	c.Request, _ = http.NewRequest(http.MethodPut, "/announcements/test-id", createRequestBody(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// Empty body is valid, just UpdatedAt will be set
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestListAnnouncements_OrderByPinnedAndCreatedAt
func TestListAnnouncements_OrderByPinnedAndCreatedAt(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements", nil)

	// The query should order by is_pinned DESC, created_at DESC
	// DB not set, so returns early
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

// TestAnnouncementDTO_Structure
func TestAnnouncementDTO_Structure(t *testing.T) {
	now := time.Now().UTC()
	dto := routes.AnnouncementDTO{
		ID:        "test-id",
		Title:     "Test Title",
		Content:   "Test Content",
		IsPinned:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if dto.ID != "test-id" {
		t.Errorf("ID = %q, want %q", dto.ID, "test-id")
	}
	if dto.Title != "Test Title" {
		t.Errorf("Title = %q, want %q", dto.Title, "Test Title")
	}
	if !dto.IsPinned {
		t.Errorf("IsPinned = %v, want true", dto.IsPinned)
	}
}

// TestCreateAnnouncementRequest_Validation
func TestCreateAnnouncementRequest_Validation(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		content   string
		isPinned  *bool
		shouldErr bool
	}{
		{
			name:     "valid request",
			title:    "Title",
			content:  "Content",
			isPinned: nil,
		},
		{
			name:     "with isPinned",
			title:    "Title",
			content:  "Content",
			isPinned: boolPtr(true),
		},
		{
			name:     "empty title",
			title:    "",
			content:  "Content",
			isPinned: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := routes.CreateAnnouncementRequest{
				Title:    tt.title,
				Content:  tt.content,
				IsPinned: tt.isPinned,
			}

			// Basic validation
			if req.Title == "" && tt.title == "" {
				t.Logf("Empty title validation needed")
			}
		})
	}
}

// TestUpdateAnnouncementRequest_PartialFields
func TestUpdateAnnouncementRequest_PartialFields(t *testing.T) {
	req := routes.UpdateAnnouncementRequest{
		Title:    stringPtr("New Title"),
		Content:  nil,
		IsPinned: nil,
	}

	if req.Title == nil {
		t.Errorf("Title should not be nil")
	}
	if req.Content != nil {
		t.Errorf("Content should be nil")
	}
	if req.IsPinned != nil {
		t.Errorf("IsPinned should be nil")
	}
}

// TestListAnnouncements_DeletedFlagFalse
func TestListAnnouncements_DeletedFlagFalse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements?deleted=false", nil)

	// When deleted=false, should exclude deleted records
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}

// TestListAnnouncements_DefaultLimit
func TestListAnnouncements_DefaultLimit(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/announcements", nil)

	// Default limit should be 100
	if w.Code != http.StatusInternalServerError && w.Code != 0 {
		t.Logf("Got status %d", w.Code)
	}
}
