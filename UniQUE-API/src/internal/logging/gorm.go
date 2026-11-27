package logging

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormLogger adapts GORM's logger to the structured log schema.
type GormLogger struct {
	level logger.LogLevel
}

// NewGormLogger returns a slog-backed GORM logger.
func NewGormLogger(level logger.LogLevel) logger.Interface {
	return &GormLogger{level: level}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &GormLogger{level: level}
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Info {
		slog.Info("gorm_info", slog.String("message", msg), slog.Any("data", data))
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Warn {
		slog.Warn("gorm_warning", slog.String("message", msg), slog.Any("data", data))
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Error {
		slog.Error("gorm_error", slog.String("message", msg), slog.Any("data", data))
	}
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= logger.Silent {
		return
	}

	sql, rows := fc()
	attrs := []any{
		slog.Int64("latency_ms", time.Since(begin).Milliseconds()),
		slog.Int64("rows", rows),
		slog.String("sql", sql),
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		attrs = append(attrs, slog.Any("error", err))
		slog.Error("gorm_query_failed", attrs...)
		return
	}

	if l.level >= logger.Info {
		slog.Debug("gorm_query", attrs...)
	}
}
