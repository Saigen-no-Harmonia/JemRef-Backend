package middleware

import (
	"bytes"
	"encoding/json"
	"jemref_go/internal/config"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected slog.Level
	}{
		{name: "DEBUG", logLevel: "DEBUG", expected: slog.LevelDebug},
		{name: "小文字でも判定できる", logLevel: "debug", expected: slog.LevelDebug},
		{name: "WARN", logLevel: "WARN", expected: slog.LevelWarn},
		{name: "ERROR", logLevel: "ERROR", expected: slog.LevelError},
		{name: "INFO", logLevel: "INFO", expected: slog.LevelInfo},
		{name: "未知の文字列はINFO扱い", logLevel: "UNKNOWN", expected: slog.LevelInfo},
		{name: "空文字はINFO扱い", logLevel: "", expected: slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseLogLevel(tt.logLevel))
		})
	}
}

func TestMiddleware_CommonLogger(t *testing.T) {
	path := "/api/v0/test"

	tests := []struct {
		name          string
		routePath     string
		rawQuery      string
		statusCode    int
		expectSkipLog bool
		expectedLevel string
	}{
		{
			name:          "正常_2xxはINFOレベルで出力される",
			routePath:     path,
			statusCode:    http.StatusOK,
			expectedLevel: "INFO",
		},
		{
			name:          "正常_4xxはWARNレベルで出力される",
			routePath:     path,
			statusCode:    http.StatusNotFound,
			expectedLevel: "WARN",
		},
		{
			name:          "正常_5xxはERRORレベルで出力される",
			routePath:     path,
			statusCode:    http.StatusInternalServerError,
			expectedLevel: "ERROR",
		},
		{
			name:          "正常_クエリパラメータがpathに連結される",
			routePath:     path,
			rawQuery:      "foo=bar",
			statusCode:    http.StatusOK,
			expectedLevel: "INFO",
		},
		{
			name:          "正常_healthエンドポイントはログ出力されない",
			routePath:     "/api/v0/health",
			statusCode:    http.StatusOK,
			expectSkipLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			cfg := config.Config{Env: "test", LogLevel: "DEBUG"}
			var buf bytes.Buffer
			router.Use(CommonLoggerWithWriter(cfg, &buf))

			router.GET(tt.routePath, func(c *gin.Context) {
				c.Status(tt.statusCode)
			})

			req := httptest.NewRequest(http.MethodGet, tt.routePath, nil)
			req.URL.RawQuery = tt.rawQuery
			req.Header.Set("User-Agent", "test-agent")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if tt.expectSkipLog {
				assert.Empty(t, buf.String())
				return
			}

			var logLine map[string]any
			require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &logLine))

			expectedPath := tt.routePath
			if tt.rawQuery != "" {
				expectedPath += "?" + tt.rawQuery
			}

			assert.Equal(t, "HTTP Request", logLine["msg"])
			assert.Equal(t, tt.expectedLevel, logLine["level"])
			assert.Equal(t, http.MethodGet, logLine["method"])
			assert.Equal(t, expectedPath, logLine["path"])
			assert.Equal(t, float64(tt.statusCode), logLine["status"])
			assert.Equal(t, "192.0.2.1", logLine["ip"])
			assert.Equal(t, "test-agent", logLine["user_agent"])
			assert.Equal(t, "", logLine["errors"])
			assert.Contains(t, logLine, "latency")
		})
	}
}

// newLoggerのlocal(tint)分岐は人間可読テキスト形式でパース困難なため、
// panicせず実行できることのみ簡易的に確認する。
func TestMiddleware_CommonLogger_LocalEnvUsesTintHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	cfg := config.Config{Env: "local", LogLevel: "DEBUG"}
	var buf bytes.Buffer

	assert.NotPanics(t, func() {
		router.Use(CommonLoggerWithWriter(cfg, &buf))
	})

	router.GET("/api/v0/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v0/test", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		router.ServeHTTP(rec, req)
	})

	assert.NotEmpty(t, buf.String())
}
