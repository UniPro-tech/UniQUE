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

func TestMaskedBodyRedactsBasicAuthorization(t *testing.T) {
	output := logValue(t, MaskedBody([]byte("Authorization: Basic dXNlcjpwYXNz")))
	if strings.Contains(output, "dXNlcjpwYXNz") {
		t.Fatalf("basic credentials leaked into log: %s", output)
	}
}

func TestMaskedBodyRedactsSingleQuotedKeys(t *testing.T) {
	cases := []string{
		`{'access_token': 'secret-value'}`,
		`{'password': 'secret"tail'}`,
		`{"password": "secret'tail"}`,
		`{'password': 'secret\'tail'}`,
		`{'password': 'sample-secret`,
		`{"password": "sample-secret`,
	}
	for _, input := range cases {
		output := logValue(t, MaskedBody([]byte(input)))
		if strings.Contains(output, "secret") || strings.Contains(output, "tail") {
			t.Fatalf("credential with quoted key/value leaked into log: %s", output)
		}
	}
}

func TestMaskedBodyRedactsJSONStringValues(t *testing.T) {
	output := logValue(t, MaskedBody([]byte(`{"message":"Authorization: Bearer abc.def.ghi"}`)))
	if strings.Contains(output, "abc.def.ghi") {
		t.Fatalf("credential in a JSON string leaked into log: %s", output)
	}
}

func TestRedactedQueryHandlesQuestionMarkInValues(t *testing.T) {
	output := logValue(t, RedactedQuery("token=secret?tail&safe=1"))
	if strings.Contains(output, "secret") {
		t.Fatalf("credential in raw query leaked into log: %s", output)
	}
	if !strings.Contains(output, "safe") {
		t.Fatalf("non-sensitive query value should be preserved: %s", output)
	}
}

func TestRedactedQueryParseFailureIsSafe(t *testing.T) {
	output := logValue(t, RedactedQuery("%74oken=secret&bad=%ZZ"))
	if strings.Contains(output, "secret") {
		t.Fatalf("credential leaked when query parsing failed: %s", output)
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

func TestRedactedURLRejectsNonURL(t *testing.T) {
	output := logValue(t, RedactedURL("token=secret/path://host"))
	if strings.Contains(output, "secret") {
		t.Fatalf("non-URL value leaked into log: %s", output)
	}
	if !strings.Contains(output, "[REDACTED_URL]") {
		t.Fatalf("non-URL value should be replaced with a fixed marker: %s", output)
	}
}

func TestMaskedBodyTruncates(t *testing.T) {
	body := MaskedBody(bytes.Repeat([]byte("a"), maxLoggedBodyBytes*2))
	output := logValue(t, body)
	if !strings.Contains(output, "...(truncated)") {
		t.Fatalf("expected truncation marker in %s", output)
	}
}
