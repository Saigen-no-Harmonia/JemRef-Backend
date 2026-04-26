package handler

import (
	"jemref_go/internal/usecase"

	"github.com/gin-gonic/gin"
)

type RecordHandler struct {
	usecase *usecase.RecordUsecase
}

// コンストラクタ
func NewRecordHandler(uc *usecase.RecordUsecase) *RecordHandler {
	return &RecordHandler{usecase: uc}
}

func (h *RecordHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 	r.GET("/records", h.GetRecords)
	// 	r.GET("/records/:id", h.GetRecord)
	r.POST("/records", h.CreateRecord)
	// r.PUT("/records/:id", h.UpdateRecord)
	// r.DELETE("/records/:id", h.DeleteRecord)
}

func (h *RecordHandler) CreateRecord(c *gin.Context) {
	var req CreateRecordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// usecase用の構造体にマッピング
	input := toCreateRecordInput(req)

	output, err := h.usecase.CreateRecord(input)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	res := RecordResponse{
		UserId: output.UserId,
	}

	c.JSON(201, res)
}

func toCreateRecordInput(req CreateRecordRequest) usecase.CreateRecordInput {
	// TODO not implemented
	return usecase.CreateRecordInput{
		UserId:    req.Record.UserId,
		MainTitle: req.Record.MainTitle,
	}
}
