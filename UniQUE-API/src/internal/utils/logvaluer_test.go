package utils

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func logValue(t *testing.T, value slog.LogValuer) string {
	t.Helper()
	var buffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buffer, nil))
	logger.Info("test", slog.Any("value", value))
	return buffer.String()
}

func TestSecretIsRedacted(t *testing.T) {
	output := logValue(t, Secret("super-secret-token"))
	if strings.Contains(output, "super-secret-token") {
		t.Fatalf("secret leaked into log: %s", output)
	}
	if !strings.Contains(output, redactedValue) {
		t.Fatalf("expected redaction marker in %s", output)
	}
}

func TestMaskedBodyRedactsCredentialsButKeepsErrorCode(t *testing.T) {
	body := MaskedBody([]byte(`{"message":"Bad Request","code":50035,"access_token":"abc","nested":{"refresh_token":"def"}}`))
	output := logValue(t, body)

	for _, leaked := range []string{"abc", "def"} {
		if strings.Contains(output, `"`+leaked+`"`) {
			t.Fatalf("credential %q leaked into log: %s", leaked, output)
		}
	}
	if !strings.Contains(output, "50035") {
		t.Fatalf("Discord error code should be preserved: %s", output)
	}
	if !strings.Contains(output, redactedValue) {
		t.Fatalf("expected redaction marker in %s", output)
	}
}

func TestMaskedBodyRedactsAuthorizationHeader(t *testing.T) {
	output := logValue(t, MaskedBody([]byte("Authorization: Bot abc.def.ghi")))
	if strings.Contains(output, "abc.def.ghi") {
		t.Fatalf("bot token leaked into log: %s", output)
	}
}

func TestRedactedURLMasksQuerySecrets(t *testing.T) {
	output := logValue(t, RedactedURL("https://auth.example/internal/token_verify?jti=session-id&code=oauth-code&safe=1"))
	for _, leaked := range []string{"session-id", "oauth-code"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("value %q leaked into log: %s", leaked, output)
		}
	}
	if !strings.Contains(output, "safe=1") {
		t.Fatalf("non-sensitive query value should be preserved: %s", output)
	}
}

func TestMaskedBodyTruncates(t *testing.T) {
	body := MaskedBody(bytes.Repeat([]byte("a"), maxLoggedBodyBytes*2))
	output := logValue(t, body)
	if !strings.Contains(output, "...(truncated)") {
		t.Fatalf("expected truncation marker in %s", output)
	}
}
