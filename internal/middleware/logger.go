package middleware

import (
	"jemref_go/internal/config"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
)

// CommonLogger 共通ロガーミドルウェア
func CommonLogger(cfg config.Config) gin.HandlerFunc {
	logger := newLogger(cfg.Env, cfg.LogLevel)

	// ログ出力しない設定
	skipList := []string{"/api/v0/health"}
	skip := make(map[string]struct{}, len(skipList))
	for _, path := range skipList {
		skip[path] = struct{}{}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if _, ok := skip[path]; ok {
			return
		}

		end := time.Now()
		latency := end.Sub(start)

		if raw != "" {
			path = path + "?" + raw
		}

		status := c.Writer.Status()
		method := c.Request.Method
		level := slog.LevelInfo
		if status >= 400 && status < 500 {
			level = slog.LevelWarn
		} else if status >= 500 {
			level = slog.LevelError
		}

		logger.LogAttrs(
			c.Request.Context(),
			level,
			"HTTP Request",
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}

// newLogger ロガー生成 視認性向上のため、ローカルのみtintのロガーを使用する
func newLogger(env string, logLevel string) *slog.Logger {
	level := parseLogLevel(logLevel)

	if env == "local" {
		return slog.New(
			tint.NewTextHandler(os.Stdout, &tint.Options{
				Level: level,
			}),
		)
	}

	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		}),
	)
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
