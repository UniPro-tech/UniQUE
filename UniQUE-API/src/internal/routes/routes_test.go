package routes_test

import (
	"bytes"
	"encoding/json"

	"github.com/UniPro-tech/UniQUE-API/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// テストヘルパー: DBをコンテキストに追加
func setDBInContext(c *gin.Context, db *gorm.DB) {
	c.Set("db", db)
}

// テストヘルパー: ユーザーをコンテキストに追加
func setUserInContext(c *gin.Context, user *model.User) {
	c.Set("user", user)
}

// テストヘルパー: HTTPリクエストボディを作成
func createRequestBody(v interface{}) *bytes.Reader {
	data, _ := json.Marshal(v)
	return bytes.NewReader(data)
}
