package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

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
