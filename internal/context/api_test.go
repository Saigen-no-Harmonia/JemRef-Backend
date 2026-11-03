package context

import (
	stdcontext "context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestApi_SetApiId(t *testing.T) {

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "正常_書誌情報登録API",
			input: REF_API_003,
		},
		{
			name:  "正常_ユーザ情報登録API",
			input: MEM_API_002,
		},
		{
			name:  "正常_規約情報参照API",
			input: GEN_API_001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			SetApiId(c, tt.input)

			assert.Equal(t, tt.input, c.Keys[CtxKeyApiId])
			// gin.Contextをcontext.Contextとして渡してもGetApiIdで取り出せること
			assert.Equal(t, tt.input, GetApiId(c))
		})
	}
}

func TestApi_GetApiId(t *testing.T) {

	tests := []struct {
		name string
		// setup 検証対象のcontextを組み立てる
		setup       func() stdcontext.Context
		expected    string
		expectPanic bool
	}{
		{
			name: "正常_contextから取得",
			setup: func() stdcontext.Context {
				return stdcontext.WithValue(stdcontext.Background(), CtxKeyApiId, MEM_API_006)
			},
			expected:    MEM_API_006,
			expectPanic: false,
		},
		{
			name: "異常_未設定",
			setup: func() stdcontext.Context {
				return stdcontext.Background()
			},
			expectPanic: true,
		},
		{
			name: "異常_空文字",
			setup: func() stdcontext.Context {
				return stdcontext.WithValue(stdcontext.Background(), CtxKeyApiId, "")
			},
			expectPanic: true,
		},
		{
			name: "異常_string以外の型",
			setup: func() stdcontext.Context {
				return stdcontext.WithValue(stdcontext.Background(), CtxKeyApiId, 123)
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()

			if tt.expectPanic {
				assert.Panics(t, func() {
					GetApiId(ctx)
				})
				return
			}

			assert.Equal(t, tt.expected, GetApiId(ctx))
		})
	}
}
