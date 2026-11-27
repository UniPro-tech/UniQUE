package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
)

const (
	// HeaderRequestID carries a caller-supplied request identifier.
	HeaderRequestID = "X-Request-ID"
	// HeaderTraceID carries a caller-supplied trace identifier.
	HeaderTraceID = "X-Trace-ID"
	// HeaderTraceparent is the W3C Trace Context header.
	HeaderTraceparent = "traceparent"
)

// RequestContext holds the correlation identifiers for one request.
type RequestContext struct {
	RequestID string
	TraceID   string
}

type requestContextKey struct{}

// NewRequestContext reads correlation identifiers from the request, generating
// the missing ones. A W3C traceparent trace-id takes precedence when X-Trace-ID
// is absent.
func NewRequestContext(r *http.Request) RequestContext {
	rc := RequestContext{
		RequestID: safeRequestID(r.Header.Get(HeaderRequestID)),
		TraceID:   safeTraceID(r.Header.Get(HeaderTraceID)),
	}
	if rc.TraceID == "" {
		rc.TraceID = traceIDFromTraceparent(r.Header.Get(HeaderTraceparent))
	}
	if rc.RequestID == "" {
		rc.RequestID = newID(16)
	}
	if rc.TraceID == "" {
		rc.TraceID = newID(16)
	}
	return rc
}

// LogAttrs returns the correlation attributes for slog.
func (rc RequestContext) LogAttrs() []any {
	return []any{
		slog.String("request_id", rc.RequestID),
		slog.String("trace_id", rc.TraceID),
	}
}

// WithContext stores the correlation identifiers in ctx.
func (rc RequestContext) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestContextKey{}, rc)
}

// FromContext retrieves the correlation identifiers from ctx.
func FromContext(ctx context.Context) (RequestContext, bool) {
	rc, ok := ctx.Value(requestContextKey{}).(RequestContext)
	return rc, ok
}

// Inject copies the correlation identifiers from ctx onto outbound headers.
func Inject(ctx context.Context, header http.Header) {
	rc, ok := FromContext(ctx)
	if !ok {
		return
	}
	header.Set(HeaderRequestID, rc.RequestID)
	header.Set(HeaderTraceID, rc.TraceID)
	header.Set(HeaderTraceparent, Traceparent(rc.TraceID))
}

// Traceparent builds a W3C traceparent value for an outbound request.
func Traceparent(traceID string) string {
	return "00-" + traceID + "-" + newID(8) + "-01"
}

func safeRequestID(value string) string {
	value = strings.TrimSpace(value)
	if len(value) == 0 || len(value) > 128 {
		return ""
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return ""
	}
	return value
}

func safeTraceID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) != 32 || !isHex(value) || value == strings.Repeat("0", 32) {
		return ""
	}
	return value
}

func traceIDFromTraceparent(value string) string {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) < 4 {
		return ""
	}
	traceID := strings.ToLower(parts[1])
	if len(traceID) != 32 || !isHex(traceID) || traceID == strings.Repeat("0", 32) {
		return ""
	}
	return traceID
}

func isHex(value string) bool {
	_, err := hex.DecodeString(value)
	return err == nil
}

func newID(size int) string {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		// crypto/rand failure is not expected; fall back to a fixed-length value.
		return strings.Repeat("0", size*2)
	}
	return hex.EncodeToString(buffer)
}
