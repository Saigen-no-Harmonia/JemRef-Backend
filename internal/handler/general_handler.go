package handler

// 一般ハンドラ実装

import (
	"errors"
	"fmt"
	"jemref_go/internal/domain/policy"
	handlerdto "jemref_go/internal/handler/dto"
	"jemref_go/internal/usecase"
	usecasedto "jemref_go/internal/usecase/dto"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GeneralHandler struct {
	usecase usecase.GeneralUsecase
}

func NewGeneralHandler(uc usecase.GeneralUsecase) *GeneralHandler {
	return &GeneralHandler{usecase: uc}
}

// [GEN-API-001] 規約情報参照 /policies POST
// @Summary [GEN-API-001] 規約情報参照
// @Description 指定された規約について、最新版の情報を１件返却する。規約本文は改行文字入りのテキストとなる。
// @Tags policies
// @Accept json
// @Produce json
// @Param policytype path string true "規約タイプ"
// @Success 200 {object} dto.GetPoliciesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /policies/:policy_type [get]
func (h *GeneralHandler) GetPolicies(c *gin.Context) {
	// パスパラメータ取得
	targetPolicy := c.Param(ParamPolicyType)
	policyType := policy.PolicyType(targetPolicy)

	if !policyType.IsValid() {
		c.JSON(
			http.StatusBadRequest,
			handlerdto.ErrorResponse{
				Code: "E0003",
				Message: fmt.Sprintf(
					"規約IDが不正です。policy_id=%s",
					targetPolicy,
				),
			})
		return
	}

	// Usecase呼び出し
	input := createGetPolicyInput(policyType)
	output, err := h.usecase.GetPolicies(c.Request.Context(), input)
	if err != nil {
		log.Println(err)

		// リクエストされた規約が存在しない場合
		if errors.Is(err, usecase.ErrPolicyNotFound) {
			c.JSON(
				http.StatusNotFound,
				handlerdto.ErrorResponse{
					Code: "E0006",
					Message: fmt.Sprintf("リクエストされた規約が存在しません。policy_type=%s",
						targetPolicy,
					),
				},
			)
		} else {
			// 意図しないエラー
			c.JSON(
				http.StatusInternalServerError,
				handlerdto.ErrorResponse{
					Code:    "A0001",
					Message: "Fatal: internal server error",
				},
			)
		}

		return
	}

	res := createGetPoliciesResponse(*output)
	c.JSON(http.StatusOK, res)
}

// createGetPolicyInput Usecase用Input構造体を作成
func createGetPolicyInput(tp policy.PolicyType) usecasedto.GetPoliciesInput {
	return usecasedto.GetPoliciesInput{
		PolicyId: tp.GetId(),
	}
}

// createGetPoliciesResponse レスポンス用構造体を作成
func createGetPoliciesResponse(o usecasedto.GetPoliciesOutput) handlerdto.GetPoliciesResponse {
	// DBの規約IDを、レスポンス用に規約タイプ名称へと変換
	policyType := policy.PolicyTypeFromCode(o.PolicyId)

	return handlerdto.GetPoliciesResponse{
		PolicyType:    string(policyType),
		Label:         o.Label,
		LatestVersion: o.LatestVersion,
		EffectiveDate: o.EffectiveDate.Format("2006-01-02"),
		Content:       o.Content,
	}
}
