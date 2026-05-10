package handler

// 一般ハンドラ実装

import (
	"fmt"
	"jemref_go/internal/domain/policy"
	"jemref_go/internal/handler/dto"
	"jemref_go/internal/usecase"
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
		c.JSON(400, dto.ErrorResponse{
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
		c.JSON(400, dto.ErrorResponse{
			Code: "E0003",
			Message: fmt.Sprintf(
				"規約IDが不正です。policy_id=%s",
				targetPolicy,
			),
		})
		return
	}

	// Usecase呼び出し
	input, err := createGetPolicyInput(policyType)
	// TODO エラー処理
	if err != nil {
		log.Println(err)
		c.JSON(500, dto.ErrorResponse{
			Code:    "A001",
			Message: "Server Error",
		})
		return
	}

	output, err := h.usecase.GetPolicies(input)
	// TODO エラー処理
	if err != nil {
		log.Println(err)
		// TODO レスポンスをswitch?
		return
	}

	res, err := createGetPoliciesResponse(*output)
	// TODO エラー処理
	if err != nil {
		log.Println(err)
		c.JSON(500, dto.ErrorResponse{
			Code:    "A001",
			Message: "Server Error",
		})
		return
	}

	c.JSON(200, res)
}

func createGetPolicyInput(tp policy.PolicyType) (usecase.GetPoliciesInput, error) {
	// 規約タイプが正しいことはチェック済み
	policyId, _ := tp.Code()
	return usecase.GetPoliciesInput{
		PolicyType: policyId,
	}, nil
}

func createGetPoliciesResponse(o usecase.GetPoliciesOutput) (dto.GetPoliciesResponse, error) {
	// DBの規約IDを規約タイプ名称に変換
	policyType, err := policy.PolicyTypeFromCode(o.PolicyType)
	if err != nil {
		log.Println(err)
		return dto.GetPoliciesResponse{}, err
	}
	policyTypeStr := string(policyType)

	// レスポンスオブジェクトに詰めて返却
	return dto.GetPoliciesResponse{
		PolicyType:    policyTypeStr,
		Label:         o.Label,
		LatestVersion: o.LatestVersion,
		EffectiveDate: o.EffectiveDate.Format("yyyy-mm-dd"),
		Content:       o.Content,
	}, nil
}
