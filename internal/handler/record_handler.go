package handler

import (
	ctxutil "jemref_go/internal/context"
	handlerdto "jemref_go/internal/handler/dto"
	"jemref_go/internal/usecase"
	usecasedto "jemref_go/internal/usecase/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

// statusOk StatusResponseで返すステータス文字列
const statusOk = "OK"

type RecordHandler struct {
	usecase usecase.RecordUsecase
}

// コンストラクタ
func NewRecordHandler(uc usecase.RecordUsecase) *RecordHandler {
	return &RecordHandler{usecase: uc}
}

// 共通認証ルート
func (h *RecordHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/records", h.GetRecords)
	r.GET("/records/:"+ParamRecordId, h.GetRecord)
	r.POST("/records", h.CreateRecord)
	r.PUT("/records/:"+ParamRecordId, h.UpdateRecord)
	r.DELETE("/records/:"+ParamRecordId, h.DeleteRecord)
}

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
	internalUid, ok := ctxutil.GetUserId(c)
	if !ok {
		c.Error(ErrInternalUidNotFound)
		return
	}

	publicUid, ok := ctxutil.GetPublicUid(c)
	if !ok {
		c.Error(ErrPublicUidNotFound)
		return
	}

	var req handlerdto.CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(toBindingError(err, createRecordRequiredErrors))
		return
	}

	input, err := toCreateRecordInput(internalUid, req)
	if err != nil {
		c.Error(err)
		return
	}

	ctxutil.SetApiId(c, ctxutil.REF_API_003)
	output, err := h.usecase.Create(c, input)
	if err != nil {
		c.Error(err)
		return
	}

	res := handlerdto.CreateRecordResponse{
		PublicUserId:   publicUid,
		PublicRecordId: output.RecordId,
	}

	c.JSON(http.StatusCreated, res)
}

// @Summary [REF-API-004] 書誌情報更新
// @Description ユーザに紐づく書誌情報を更新する。ユーザ情報はfirebase tokenから取得する。<br>
// @Tags records
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param id path string true "ユーザ書誌ID"
// @Param Record body dto.UpdateRecordRequest true "書誌情報更新リクエスト"
// @Success 200 {object} dto.StatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /records/{id} [put]
func (h *RecordHandler) UpdateRecord(c *gin.Context) {}

// @Summary [REF-API-005] 書誌情報削除
// @Description ユーザに紐づく書誌情報を論理削除する。ユーザ情報はfirebase tokenから取得する。<br>
// @Tags records
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param id path string true "ユーザ書誌ID"
// @Success 200 {object} dto.StatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /records/{id} [delete]
func (h *RecordHandler) DeleteRecord(c *gin.Context) {
	internalUid, ok := ctxutil.GetUserId(c)
	if !ok {
		c.Error(ErrInternalUidNotFound)
		return
	}

	recordId := c.Param(ParamRecordId)
	if recordId == "" {
		c.Error(ErrRecordIdRequired)
		return
	}

	ctxutil.SetApiId(c, ctxutil.REF_API_005)
	err := h.usecase.Delete(c, usecasedto.DeleteRecordInput{
		InternalUserId: internalUid,
		RecordId:       recordId,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, handlerdto.StatusResponse{Status: statusOk})
}
