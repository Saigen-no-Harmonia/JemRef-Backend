package handler

// 一般ハンドラ実装

import (
	"jemref_go/internal/usecase"

	"github.com/gin-gonic/gin"
)

type GeneralHandler struct {
	usecase *usecase.GeneralUsecase
}

func NewGeneralHandler(uc *usecase.GeneralUsecase) *GeneralHandler {
	return &GeneralHandler{usecase: uc}
}

// [GEN-API-001] ユーザ規約参照 /policies POST
// @Summary [GEN-API-001] ユーザ規約参照
// @Description 指定された規約情報を１件返却する。規約本文は改行文字入りのテキストとなる。
// @Tags policies
// @Accept json
// @Produce json
// @Success 200 {object} dto.GetPoliciesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /policies [get]
func (h *GeneralHandler) GetPolicies(c *gin.Context) {
}
