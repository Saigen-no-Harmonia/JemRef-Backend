package handler

// Healthハンドラ実装

import (
	"database/sql"
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

func (h *HealthHandler) Health(c *gin.Context) {

	// DB疎通確認
	err := h.db.PingContext(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "NG",
		})
		return
	}

	c.JSON(200, gin.H{"status": "OK"})
}
