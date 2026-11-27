package db

import (
	"os"

	"github.com/UniPro-tech/UniQUE-API/internal/logging"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB() (*gorm.DB, error) {
	dsn := os.Getenv("DB_DSN")
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logging.NewGormLogger(logger.Warn),
	})
}
