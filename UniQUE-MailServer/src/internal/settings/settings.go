package settings

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func LoadValues() (map[string]string, error) {
	values := make(map[string]string)
	dsn := os.Getenv("SETTINGS_DB_DSN")
	if dsn == "" {
		dsn = os.Getenv("DB_DSN")
	}
	if dsn == "" {
		return values, nil
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("SELECT `key`, `value` FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, rows.Err()
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
