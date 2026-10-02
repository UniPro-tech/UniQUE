package settings

import (
	"os"

	"gorm.io/gorm"
)

type settingRow struct {
	Key   string `gorm:"column:key"`
	Value string `gorm:"column:value"`
}

func (settingRow) TableName() string { return "settings" }

func LoadValues(db *gorm.DB) (map[string]string, error) {
	values := make(map[string]string)
	var rows []settingRow
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	return values, nil
}

func Resolve(values map[string]string, environment, key, fallback string) string {
	if value, exists := os.LookupEnv(environment); exists {
		return value
	}
	if value, exists := values[key]; exists {
		return value
	}
	return fallback
}
