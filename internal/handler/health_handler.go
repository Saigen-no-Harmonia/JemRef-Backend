package handler

// Healthハンドラ実装

import (
	"database/sql"
	"jemref_go/internal/handler/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// Healthエンドポイント
//
// @Summary Health
// @Description Health
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} dto.StatusResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {

	// DB疎通確認
	err := h.db.PingContext(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.StatusResponse{
			Status: "NG",
		})
		return
	}

	c.JSON(200, dto.StatusResponse{Status: "OK"})
}
