package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Init configures the process-wide JSON logger and attaches the service name.
func Init(component string) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler).With(slog.String("component", component)))
}

type slogWriter struct{}

// Write routes third-party output (e.g. Gin) through slog so that no plain-text
// lines reach stdout.
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

// PanicAttr describes a panic without exposing its value, which may contain
// credentials or other sensitive data.
func PanicAttr(recovered any) slog.Attr {
	return slog.String("panic_type", fmt.Sprintf("%T", recovered))
}
