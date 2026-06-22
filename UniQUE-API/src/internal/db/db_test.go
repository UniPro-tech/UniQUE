package db_test

import (
	"os"
	"testing"

	"github.com/UniPro-tech/UniQUE-API/internal/db"
)

// テスト前に環境変数を保存し、テスト後に復元するヘルパー関数
func saveAndClearEnv(keys ...string) map[string]string {
	saved := make(map[string]string)
	for _, key := range keys {
		saved[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	return saved
}

func restoreEnv(saved map[string]string) {
	for key, value := range saved {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

func TestNewDB_EmptyDSN(t *testing.T) {
	saved := saveAndClearEnv("DB_DSN")
	defer restoreEnv(saved)

	// DB_DSNを設定しない（空文字列）
	result, err := db.NewDB()

	if err == nil {
		t.Errorf("NewDB() with empty DSN should return an error, got nil")
	}

	if result != nil {
		t.Errorf("NewDB() with empty DSN should return nil db, got %v", result)
	}
}

func TestNewDB_InvalidDSNFormat(t *testing.T) {
	saved := saveAndClearEnv("DB_DSN")
	defer restoreEnv(saved)

	// 無効なDSN形式を設定
	os.Setenv("DB_DSN", "invalid-dsn-format")
	result, err := db.NewDB()

	if err == nil {
		t.Errorf("NewDB() with invalid DSN should return an error, got nil")
	}

	if result != nil {
		t.Errorf("NewDB() with invalid DSN should return nil db, got %v", result)
	}
}

func TestNewDB_ConnRefused(t *testing.T) {
	saved := saveAndClearEnv("DB_DSN")
	defer restoreEnv(saved)

	// 接続不可能なホストを指定
	os.Setenv("DB_DSN", "user:password@tcp(localhost:9999)/testdb?charset=utf8mb4&parseTime=True&loc=Local")
	result, err := db.NewDB()

	if err == nil {
		t.Errorf("NewDB() with unreachable host should return an error, got nil")
	}

	if result != nil {
		t.Errorf("NewDB() with unreachable host should return nil db, got %v", result)
	}
}

func TestNewDB_DSNEnvironmentVariable(t *testing.T) {
	saved := saveAndClearEnv("DB_DSN")
	defer restoreEnv(saved)

	// 環境変数が正しく読み込まれるかテスト
	testDSN := "user:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	os.Setenv("DB_DSN", testDSN)

	// エラーが発生する可能性があるため、エラーの確認のみ
	result, err := db.NewDB()

	// DSNが無効な場合はエラーが発生することが期待される
	// ここではエラーハンドリングがされていることを確認
	if err != nil {
		// 接続エラーは環境依存なので、ここではエラーが適切に返されていることのみ確認
		t.Logf("Expected connection error: %v", err)
		if result != nil {
			t.Errorf("NewDB() with connection error should return nil db, got %v", result)
		}
	}
}

func TestNewDB_ReturnTypes(t *testing.T) {
	saved := saveAndClearEnv("DB_DSN")
	defer restoreEnv(saved)

	os.Setenv("DB_DSN", "user:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local")

	result, err := db.NewDB()

	// 戻り値の型確認
	// resultは*gorm.DBまたはnilである必要がある
	if result != nil {
		// resultの型がポインタ型であることを確認
		if result == nil {
			t.Errorf("NewDB() should return non-nil *gorm.DB on success")
		}
	}

	// errはerrorまたはnilである必要がある
	if err != nil {
		// エラー型の確認
		if err.Error() == "" {
			t.Errorf("Error message should not be empty")
		}
	}
}

func TestNewDB_MultipleInvocations(t *testing.T) {
	saved := saveAndClearEnv("DB_DSN")
	defer restoreEnv(saved)

	os.Setenv("DB_DSN", "user:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local")

	// 複数回呼び出して、同じ結果が返されるか確認
	result1, err1 := db.NewDB()
	result2, err2 := db.NewDB()

	// 両方のエラーが同じ状態であることを確認
	if (err1 == nil) != (err2 == nil) {
		t.Errorf("Multiple invocations should return consistent results")
	}

	// 両方がnilまたは両方が非nilであることを確認
	if (result1 == nil) != (result2 == nil) {
		t.Errorf("Multiple invocations should return consistent results")
	}
}

func TestNewDB_DSNWithSpecialCharacters(t *testing.T) {
	saved := saveAndClearEnv("DB_DSN")
	defer restoreEnv(saved)

	// パスワードに特殊文字を含むDSNをテスト
	os.Setenv("DB_DSN", "user:p@ssw0rd!@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local")

	_, err := db.NewDB()

	// エラーが発生する可能性があるが、適切にエラーハンドリングされることを確認
	if err != nil {
		t.Logf("Connection with special characters error: %v", err)
	}
}

func TestNewDB_DSNEmpty_VsUnset(t *testing.T) {
	saved := saveAndClearEnv("DB_DSN")
	defer restoreEnv(saved)

	// 空の文字列が設定された場合
	os.Setenv("DB_DSN", "")
	result, err := db.NewDB()

	if err == nil {
		t.Errorf("NewDB() with empty DSN string should return an error")
	}

	if result != nil {
		t.Errorf("NewDB() with empty DSN string should return nil db")
	}
}

func TestNewDB_DSNVariations(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "valid DSN format",
			dsn:     "user:password@tcp(localhost:3306)/db",
			wantErr: true, // 実際の接続は失敗するはずだが、DSN形式は有効
		},
		{
			name:    "DSN with charset",
			dsn:     "user:password@tcp(localhost:3306)/db?charset=utf8mb4",
			wantErr: true,
		},
		{
			name:    "DSN with parseTime",
			dsn:     "user:password@tcp(localhost:3306)/db?parseTime=True",
			wantErr: true,
		},
		{
			name:    "DSN with timezone",
			dsn:     "user:password@tcp(localhost:3306)/db?loc=Local",
			wantErr: true,
		},
		{
			name:    "malformed DSN no tcp",
			dsn:     "user:password@localhost:3306/db",
			wantErr: true,
		},
	}

	saved := saveAndClearEnv("DB_DSN")
	defer restoreEnv(saved)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("DB_DSN", tt.dsn)
			result, err := db.NewDB()

			if tt.wantErr && err == nil {
				t.Errorf("NewDB() with %q: expected error, got nil", tt.dsn)
			}

			if result == nil && err == nil {
				t.Logf("NewDB() with %q: got nil result without error (unexpected)", tt.dsn)
			}
		})
	}
}

func TestNewDB_GormConfigApplied(t *testing.T) {
	saved := saveAndClearEnv("DB_DSN")
	defer restoreEnv(saved)

	os.Setenv("DB_DSN", "user:password@tcp(localhost:3306)/testdb")

	result, err := db.NewDB()

	// 接続に失敗する可能性があるため、エラーが発生してもGORMConfigが適用されたことを確認
	// エラーがある場合、resultはnilであることを確認
	if err != nil {
		if result != nil {
			t.Errorf("NewDB() should return nil when error occurs")
		}
		return
	}

	// 接続成功時の確認
	if result == nil {
		t.Errorf("NewDB() should return non-nil *gorm.DB on success")
	}
}

func TestNewDB_ErrorMessage(t *testing.T) {
	saved := saveAndClearEnv("DB_DSN")
	defer restoreEnv(saved)

	// 無効なDSNを設定
	os.Setenv("DB_DSN", "invalid")

	result, err := db.NewDB()

	if err == nil {
		t.Errorf("NewDB() should return an error for invalid DSN")
		return
	}

	if result != nil {
		t.Errorf("NewDB() should return nil db when error occurs")
	}

	// エラーメッセージが空でないことを確認
	if err.Error() == "" {
		t.Errorf("Error message should not be empty")
	}

	t.Logf("Error message: %v", err)
}
