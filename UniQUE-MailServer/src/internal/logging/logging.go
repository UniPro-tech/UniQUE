package logging

import (
	"io"
	"log/slog"
	"net/url"
	"os"
	"strings"
)

const redactedValue = "[REDACTED]"

// Init configures the process-wide JSON logger and attaches the service name.
func Init(component string) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler).With(slog.String("component", component)))
}

type slogWriter struct{}

// Write routes third-party output (e.g. Gin) through slog.
func (slogWriter) Write(p []byte) (int, error) {
	message := strings.TrimSpace(string(p))
	if message != "" {
		slog.Warn("gin_output", slog.String("message", message))
	}
	return len(p), nil
}

// Writer returns an io.Writer that emits each line as a structured log entry.
func Writer() io.Writer {
	return slogWriter{}
}

var sensitiveQueryKeys = map[string]struct{}{
	"access_token":  {},
	"refresh_token": {},
	"id_token":      {},
	"token":         {},
	"secret":        {},
	"client_secret": {},
	"password":      {},
	"authorization": {},
	"jti":           {},
	"sid":           {},
	"session_id":    {},
	"code":          {},
	"state":         {},
}

// Secret wraps a credential so slog prints a fixed placeholder.
type Secret string

// LogValue implements slog.LogValuer.
func (s Secret) LogValue() slog.Value {
	if s == "" {
		return slog.StringValue("[unset]")
	}
	return slog.StringValue(redactedValue)
}

// RedactedURL masks credentials embedded in a URL or raw query string.
type RedactedURL string

// LogValue implements slog.LogValuer.
func (u RedactedURL) LogValue() slog.Value {
	return slog.StringValue(redactURL(string(u)))
}

func redactURL(raw string) string {
	if raw == "" {
		return ""
	}
	if strings.Contains(raw, "?") || strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err == nil {
			parsed.User = nil
			parsed.RawQuery = redactQuery(parsed.RawQuery)
			return parsed.String()
		}
	}
	return redactQuery(raw)
}

func redactQuery(raw string) string {
	if raw == "" {
		return ""
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return "[REDACTED_QUERY]"
	}
	for key := range values {
		if _, sensitive := sensitiveQueryKeys[strings.ToLower(key)]; sensitive {
			values.Set(key, redactedValue)
		}
	}
	return values.Encode()
}
