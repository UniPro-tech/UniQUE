package util

import (
	"time"

	"gorm.io/gorm"
)

func DeletedAtPtr(t gorm.DeletedAt) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}
