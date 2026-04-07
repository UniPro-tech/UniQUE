package routes

import "github.com/gin-gonic/gin"

// IsOAuth returns true if the request was authenticated using an OAuth access token
func IsOAuth(c *gin.Context) bool {
	if v, ok := c.Get("isOAuth"); ok {
		if b, ok2 := v.(bool); ok2 {
			return b
		}
	}
	return false
}

// GetScope returns the OAuth scope string from the context (may be empty)
func GetScope(c *gin.Context) string {
	if v, ok := c.Get("scope"); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}
