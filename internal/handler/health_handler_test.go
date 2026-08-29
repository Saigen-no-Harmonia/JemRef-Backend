package handler

import (
	"jemref_go/internal/testutil/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_Health(t *testing.T) {
	tests := []struct {
		name              string
		path              string
		mockDb            DbPinger
		expectedError     bool
		expectedErrorBody error
	}{
		{
			name: "Health_正常",
			path: "/api/v0/health",
			mockDb: &mock.MockDB{
				PingErr: nil,
			},
			expectedError: false,
		},
		{
			name: "Health_異常",
			path: "/api/v0/health",
			mockDb: &mock.MockDB{
				PingErr: ErrTestUnexpected,
			},
			expectedError:     true,
			expectedErrorBody: ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			h := NewHealthHandler(tt.mockDb)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v0/health",
				nil,
			)

			c.Request = req

			h.Health(c)

			// 異常系のassertion
			if tt.expectedError {
				require.Len(t, c.Errors, 1)
				assert.ErrorIs(t, c.Errors.Last().Err, tt.expectedErrorBody)
			} else {
				assert.Empty(t, c.Errors)
			}
		})
	}
}
