package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestSafeTraceIDRejectsAllZeroAndNormalizesCase(t *testing.T) {
	if got := safeTraceID("00000000000000000000000000000000"); got != "" {
		t.Fatalf("all-zero trace id should be rejected, got %q", got)
	}
	got := safeTraceID("ABCDEF0123456789ABCDEF0123456789")
	if got != "abcdef0123456789abcdef0123456789" {
		t.Fatalf("trace id should be lowercased, got %q", got)
	}
	if got := traceIDFromTraceparent("00-00000000000000000000000000000000-1111111111111111-01"); got != "" {
		t.Fatalf("all-zero traceparent trace id should be rejected, got %q", got)
	}
	if got := traceIDFromTraceparent("00-ABCDEF0123456789ABCDEF0123456789-1111111111111111-01"); got != "abcdef0123456789abcdef0123456789" {
		t.Fatalf("traceparent trace id should be lowercased, got %q", got)
	}
}

func TestPanicAttrDoesNotLeakValue(t *testing.T) {
	var buffer bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buffer, nil))
	logger.Error("panic_recovered", PanicAttr("token=secret"))

	output := buffer.String()
	if strings.Contains(output, "secret") {
		t.Fatalf("panic value leaked into log: %s", output)
	}
	if !strings.Contains(output, "string") {
		t.Fatalf("panic type should be recorded: %s", output)
	}
}
