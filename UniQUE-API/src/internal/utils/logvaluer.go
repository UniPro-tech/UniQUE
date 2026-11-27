package utils

import (
	"encoding/json"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
)

// maxLoggedBodyBytes bounds how much of an external response body is written
// to the logs. Bodies are truncated after masking so that a single log line
// cannot grow without limit.
const maxLoggedBodyBytes = 2048

const redactedValue = "[REDACTED]"

// jsonSensitiveKeys are JSON object keys whose values must never be logged.
// "code" is intentionally excluded because providers (e.g. Discord) use it
// for machine-readable error codes, which are useful when diagnosing failures.
var jsonSensitiveKeys = map[string]struct{}{
	"access_token":  {},
	"refresh_token": {},
	"id_token":      {},
	"token":         {},
	"secret":        {},
	"client_secret": {},
	"bot_token":     {},
	"password":      {},
	"authorization": {},
	"api_key":       {},
	"jti":           {},
	"sid":           {},
	"session_id":    {},
}

// querySensitiveKeys additionally covers values that are only secret in a URL
// query string (OAuth authorization codes and state nonces).
var querySensitiveKeys = map[string]struct{}{
	"access_token":  {},
	"refresh_token": {},
	"id_token":      {},
	"token":         {},
	"secret":        {},
	"client_secret": {},
	"bot_token":     {},
	"password":      {},
	"authorization": {},
	"api_key":       {},
	"jti":           {},
	"sid":           {},
	"session_id":    {},
	"code":          {},
	"state":         {},
}

// Secret wraps a credential so that slog prints a fixed placeholder instead of
// the value. Use it for tokens, secrets and session identifiers.
type Secret string

// LogValue implements slog.LogValuer.
func (s Secret) LogValue() slog.Value {
	if s == "" {
		return slog.StringValue("[unset]")
	}
	return slog.StringValue(redactedValue)
}

// MaskedBody wraps an external HTTP response body so that sensitive JSON fields
// are redacted and the payload is truncated before it reaches the log.
type MaskedBody []byte

// LogValue implements slog.LogValuer.
func (b MaskedBody) LogValue() slog.Value {
	return slog.StringValue(maskBody(b))
}

// MaskBody returns the masked, truncated body as a plain string. Callers that
// need to sanitize line breaks before logging can use this instead of the
// MaskedBody LogValuer.
func MaskBody(body []byte) string {
	return maskBody(body)
}

// RedactedURL masks credentials embedded in an absolute URL. Use
// RedactedQuery for a raw query string.
type RedactedURL string

// LogValue implements slog.LogValuer.
func (u RedactedURL) LogValue() slog.Value {
	return slog.StringValue(redactURL(string(u)))
}

// RedactedQuery masks credentials in a raw query string. Use this for
// url.URL.RawQuery, which may legitimately contain "?" inside values.
type RedactedQuery string

// LogValue implements slog.LogValuer.
func (q RedactedQuery) LogValue() slog.Value {
	return slog.StringValue(redactQuery(string(q)))
}

func maskBody(body []byte) string {
	if len(body) == 0 {
		return "(empty body)"
	}

	var masked string
	var decoded any
	if err := json.Unmarshal(body, &decoded); err == nil {
		if out, err := json.Marshal(maskJSONValue(decoded)); err == nil {
			masked = string(out)
		} else {
			masked = redactText(string(body))
		}
	} else {
		masked = redactText(string(body))
	}

	if len(masked) > maxLoggedBodyBytes {
		masked = masked[:maxLoggedBodyBytes] + "...(truncated)"
	}
	return masked
}

func maskJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			if _, sensitive := jsonSensitiveKeys[strings.ToLower(key)]; sensitive {
				typed[key] = redactedValue
				continue
			}
			typed[key] = maskJSONValue(nested)
		}
		return typed
	case []any:
		for i := range typed {
			typed[i] = maskJSONValue(typed[i])
		}
		return typed
	case string:
		return redactText(typed)
	default:
		return value
	}
}

var (
	textSensitiveKV = regexp.MustCompile(`(?i)(["']?(?:access_token|refresh_token|id_token|token|secret|client_secret|bot_token|password|authorization|api_key|jti|sid|session_id)["']?\s*[=:]\s*)("(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|"(?:\\.|[^"\\])*\\?$|'(?:\\.|[^'\\])*\\?$|[^"',}\s&]+)`)
	authorizationKV = regexp.MustCompile(`(?i)\b(Bearer|Bot|Basic)\s+[A-Za-z0-9._~+/=-]+`)
)

func redactText(text string) string {
	// Mask scheme-prefixed credentials first so that a value such as
	// "Authorization: Bot <token>" is not reduced to "Authorization: Bot".
	text = authorizationKV.ReplaceAllString(text, `$1 `+redactedValue)
	text = textSensitiveKV.ReplaceAllString(text, `$1"`+redactedValue+`"`)
	return text
}

func redactURL(raw string) string {
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "[REDACTED_URL]"
	}
	parsed.User = nil
	parsed.RawQuery = redactQuery(parsed.RawQuery)
	return parsed.String()
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
		if _, sensitive := querySensitiveKeys[strings.ToLower(key)]; sensitive {
			values.Set(key, redactedValue)
		}
	}
	return values.Encode()
}
