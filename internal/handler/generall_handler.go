package handler

// 一般ハンドラ実装

import (
	"errors"
	"fmt"
	"jemref_go/internal/domain/policy"
	handlerdto "jemref_go/internal/handler/dto"
	"jemref_go/internal/infrastructure/db"
	"jemref_go/internal/usecase"
	usecasedto "jemref_go/internal/usecase/dto"
	"log"

	"github.com/gin-gonic/gin"
)

type GeneralHandler struct {
	usecase *usecase.GeneralUsecase
}

func NewGeneralHandler(uc *usecase.GeneralUsecase) *GeneralHandler {
	return &GeneralHandler{usecase: uc}
}

// [GEN-API-001] 規約情報参照 /policies POST
// @Summary [GEN-API-001] 規約情報参照
// @Description 指定された規約について、最新版の情報を１件返却する。規約本文は改行文字入りのテキストとなる。
// @Tags policies
// @Accept json
// @Produce json
// @Success 200 {object} dto.GetPoliciesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /policies [get]
func (h *GeneralHandler) GetPolicies(c *gin.Context) {
	targetPolicy := c.Param(ParamPolicyId)
	// パラメータに規約IDがなかった場合
	if targetPolicy == "" {
		c.JSON(400, handlerdto.ErrorResponse{
			Code: "E0002",
			Message: fmt.Sprintf(
				"必須パラメータがありません。%s",
				targetPolicy,
			),
		})
		return
	}

	// 規約IDが不正だった場合
	policyType := policy.PolicyType(targetPolicy)
	if !policyType.IsValid() {
		c.JSON(400, handlerdto.ErrorResponse{
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
		if errors.Is(err, db.ErrPolicyNotFound) {
			c.JSON(404, handlerdto.ErrorResponse{
				Code:    "E0006",
				Message: "リクエストされた規約は存在しません。policy_type={}",
			})
		} else {
			// DB接続エラーなど
			c.JSON(500, handlerdto.ErrorResponse{
				Code:    "A0001",
				Message: "Fatal: internal server error",
			})
		}

		return
	}

	res := createGetPoliciesResponse(*output)

	c.JSON(200, res)
}

// createGetPolicyInput Usecase用Input構造体を作成
func createGetPolicyInput(tp policy.PolicyType) usecasedto.GetPoliciesInput {
	policyId, _ := tp.GetId()
	return usecasedto.GetPoliciesInput{
		PolicyId: policyId,
	}
}

// createGetPoliciesResponse レスポンス用構造体を作成
func createGetPoliciesResponse(o usecasedto.GetPoliciesOutput) handlerdto.GetPoliciesResponse {
	// DBの規約IDを、レスポンス用に規約タイプ名称へと変換
	policyType, err := policy.PolicyTypeFromCode(o.PolicyId)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	return handlerdto.GetPoliciesResponse{
		PolicyType:    string(policyType),
		Label:         o.Label,
		LatestVersion: o.LatestVersion,
		EffectiveDate: o.EffectiveDate.Format("yyyy-mm-dd"),
		Content:       o.Content,
	}
}
