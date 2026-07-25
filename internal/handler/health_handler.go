package handler

// Healthハンドラ実装

import (
	"context"
	"jemref_go/internal/handler/dto"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db DbPinger
}

func NewHealthHandler(db DbPinger) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// @Summary Health
// @Description Health
// @Tags health
// @Accept json
// @Produce json
// @Success 200
// @Failure 500 {object} dto.ErrorResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {

	// DB疎通確認
	err := h.db.PingContext(c.Request.Context())

	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError,
			dto.ErrorResponse{
				Code:    "F0001",
				Message: "fatal:Internal Server Error",
			},
		)
		return
	}

	c.Status(http.StatusOK)
}

type DbPinger interface {
	PingContext(ctx context.Context) error
}
