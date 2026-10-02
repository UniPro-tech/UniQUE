package settings_test

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/UniPro-tech/UniQUE-MailServer/internal/settings"
)

func TestResolvePrecedence(t *testing.T) {
	const environment = "SMTP_HOST"
	previous, existed := os.LookupEnv(environment)
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(environment, previous)
		} else {
			_ = os.Unsetenv(environment)
		}
	})
	_ = os.Unsetenv(environment)
	values := map[string]string{"smtp.host": "database"}
	if got := settings.Resolve(values, environment, "smtp.host", "fallback"); got != "database" {
		t.Fatalf("database value = %q", got)
	}
	_ = os.Setenv(environment, "environment")
	if got := settings.Resolve(values, environment, "smtp.host", "fallback"); got != "environment" {
		t.Fatalf("environment value = %q", got)
	}
}

// A stalled connection must not block service startup indefinitely.
func TestLoadValuesConnectionTimeout(t *testing.T) {
	const network = "settings-timeout-test"
	mysql.RegisterDialContext(network, func(ctx context.Context, _ string) (net.Conn, error) {
		if _, ok := ctx.Deadline(); !ok {
			return nil, errors.New("settings connection has no deadline")
		}
		<-ctx.Done()
		return nil, ctx.Err()
	})
	t.Cleanup(func() { mysql.DeregisterDialContext(network) })
	t.Setenv("SETTINGS_DB_DSN", "test:test@"+network+"(unused)/test")
	start := time.Now()
	_, err := settings.LoadValues()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("LoadValues error = %v, want deadline exceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("settings timeout took %s, want at most 10s", elapsed)
	}
}
