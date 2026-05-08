package handler

import (
	"jemref_go/internal/handler/dto"
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

// 共通認証ルート
func (h *RecordHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/records", h.GetRecords)
	r.GET("/records/:id", h.GetRecord)
	r.POST("/records", h.CreateRecord)
	r.PUT("/records/:id", h.UpdateRecord)
	r.DELETE("/records/:id", h.DeleteRecord)
}

// [REF-API-001] 書誌一覧参照 /records GET
func (h *RecordHandler) GetRecords(c *gin.Context) {
}

// [REF-API-002] 書誌詳細参照 /records/:id GET
func (h *RecordHandler) GetRecord(c *gin.Context) {}

// [REF-API-003] 書誌情報登録 /records POST
// リクエストされた書誌情報を作成する
func (h *RecordHandler) CreateRecord(c *gin.Context) {
	var req dto.CreateRecordRequest

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

	res := dto.CreateRecordResponse{
		UserId: output.UserId,
	}

	c.JSON(201, res)
}

// [REF-API-004] 書誌情報更新 /records/:id PUT
func (h *RecordHandler) UpdateRecord(c *gin.Context) {}

// [REF-API-005] 書誌情報削除 /records/:id DELETE
func (h *RecordHandler) DeleteRecord(c *gin.Context) {}

// Usecase用の構造体にマッピングする
func toCreateRecordInput(req dto.CreateRecordRequest) usecase.CreateRecordInput {
	//. TODO not implemented
	return usecase.CreateRecordInput{
		UserId:    req.Record.UserId,
		MainTitle: req.Record.MainTitle,
	}
}
