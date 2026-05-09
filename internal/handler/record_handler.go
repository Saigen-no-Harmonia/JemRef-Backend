package handler

import (
	"jemref_go/internal/handler/dto"
	"jemref_go/internal/usecase"
	"log"

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
// @Summary [REF-API-001] 書誌一覧参照
// @Description ユーザに紐づく書誌情報をリストで返却する。ユーザ情報はfirebase tokenから取得する。
// @Tags records
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param param query dto.GetRecordsRequest true "書誌一覧取得リクエスト"
// @Success 200 {object} dto.GetRecordsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /records [get]
func (h *RecordHandler) GetRecords(c *gin.Context) {
}

// [REF-API-002] 書誌詳細参照 /records/:id GET
// @Summary [REF-API-002] 書誌詳細参照
// @Description ユーザに紐づく書誌詳細情報を１件返却する。ユーザ情報はfirebase tokenから取得する。
// @Tags records
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param id path string true "ユーザ書誌ID"
// @Success 200 {object} dto.GetRecordResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /records/{id} [get]
func (h *RecordHandler) GetRecord(c *gin.Context) {}

// [REF-API-003] 書誌情報登録 /records POST
// @Summary [REF-API-003] 書誌情報登録
// @Description ユーザに紐づく書誌情報を登録する。ユーザ情報はfirebase tokenから取得する。
// @Tags records
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param record body dto.CreateRecordRequest true "書誌情報登録リクエスト"
// @Success 201 {object} dto.CreateRecordResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /records [post]
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

	// 仮
	log.Print(output)

	res := dto.CreateRecordResponse{
		PublicUserId:   "sample_public_user_id",
		PublicRecordId: "sample_public_record_id",
	}

	c.JSON(201, res)
}

// [REF-API-004] 書誌情報更新 /records/:id PUT
// @Summary [REF-API-004] 書誌情報更新
// @Description ユーザに紐づく書誌情報を更新する。ユーザ情報はfirebase tokenから取得する。<br>
// @Tags records
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param id path string true "ユーザ書誌ID"
// @Param Policies body dto.UpdateRecordRequest true "書誌情報更新リクエスト"
// @Success 200 {object} dto.StatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /records/{id} [put]
func (h *RecordHandler) UpdateRecord(c *gin.Context) {}

// [REF-API-005] 書誌情報削除 /records/:id DELETE
// @Summary [REF-API-005] 書誌情報削除
// @Description ユーザに紐づく書誌情報を論理削除する。ユーザ情報はfirebase tokenから取得する。<br>
// @Tags records
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param id path string true "ユーザ書誌ID"
// @Success 200 {object} dto.StatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /records/{id} [delete]
func (h *RecordHandler) DeleteRecord(c *gin.Context) {}

// Usecase用の構造体にマッピングする
func toCreateRecordInput(req dto.CreateRecordRequest) usecase.CreateRecordInput {
	//. TODO not implemented
	return usecase.CreateRecordInput{
		MainTitle: req.RecordDetailItem.MainTitle,
	}
}
