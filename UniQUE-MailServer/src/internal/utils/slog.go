package utils

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Secret string

func (Secret) LogValue() slog.Value {
	return slog.StringValue("**SECRET**")
}

type LoggerOption struct {
	Funcname    string
	PackageName string
	FilePath    string
	Layer       string
}

func GetLogger(c *gin.Context, option LoggerOption) *slog.Logger {
	logger := slog.Default()
	requestID, _ := c.Get("request_id")
	traceID, _ := c.Get("trace_id")
	logger.With(
		slog.String("trace_id", traceID.(string)),
		slog.String("request_id", requestID.(string)),
		slog.String("package", option.PackageName),
		slog.String("file", option.FilePath),
		slog.String("function", option.Funcname),
		slog.String("layer", option.Layer),
	)
	return logger
}
