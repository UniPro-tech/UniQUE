package discord

import (
	"io"
	"log/slog"
	"strings"

	"github.com/UniPro-tech/UniQUE-API/internal/utils"
)

const maxDiscordBodyBytes = 2048

func discordLogger() *slog.Logger {
	return slog.Default().With(slog.String("subsystem", "discord"))
}

// readDiscordBody reads a bounded amount of a Discord response body.
func readDiscordBody(reader io.Reader) []byte {
	body, _ := io.ReadAll(io.LimitReader(reader, maxDiscordBodyBytes+1))
	if len(body) > maxDiscordBodyBytes {
		body = body[:maxDiscordBodyBytes]
	}
	return body
}

// logDiscordResponse records a Discord API response. Response bodies are passed
// through utils.MaskedBody so tokens and other credentials never reach the log.
func logDiscordResponse(operation string, status int, body []byte, attrs ...any) {
	args := []any{
		slog.String("operation", operation),
		slog.Int("status", status),
	}
	if body != nil {
		args = append(args, slog.String("response_body", strings.ReplaceAll(strings.ReplaceAll(utils.MaskBody(body), "\n", ""), "\r", "")))
	}
	args = append(args, attrs...)

	logger := discordLogger()
	if status >= 200 && status < 300 {
		logger.Debug("discord_api_response", args...)
		return
	}
	logger.Error("discord_api_error", args...)
}
