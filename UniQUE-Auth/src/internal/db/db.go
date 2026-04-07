package db

import (
	"os"

	//"github.com/UniPro-tech/UniQUE-Auth/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB() (*gorm.DB, error) {
	dsn := os.Getenv("DB_DSN")
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}
