package handler

import (
	"encoding/json"
	"errors"
	"jemref_go/internal/handler/dto"
	"jemref_go/internal/testutil/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Health(t *testing.T) {
	tests := []struct {
		name              string
		path              string
		mockDb            DbPinger
		expectedStatus    int
		expectedError     bool
		expectedErrorBody dto.ErrorResponse
	}{
		{
			name: "Health_正常",
			path: "/api/v0/health",
			mockDb: &mock.MockDB{
				PingErr: nil,
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "Health_異常",
			path: "/api/v0/health",
			mockDb: &mock.MockDB{
				PingErr: errors.New("予期せぬエラー"),
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
			expectedErrorBody: dto.ErrorResponse{
				Code:    "F0001",
				Message: "fatal:Internal Server Error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			h := NewHealthHandler(tt.mockDb)

			r.GET(
				tt.path,
				h.Health,
			)

			req := httptest.NewRequest(
				http.MethodGet,
				tt.path,
				nil,
			)

			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			// 異常系のassertion
			if tt.expectedError {
				var actual dto.ErrorResponse
				err := json.Unmarshal(
					rec.Body.Bytes(),
					&actual,
				)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedErrorBody, actual)
				return
			}
		})
	}
}
