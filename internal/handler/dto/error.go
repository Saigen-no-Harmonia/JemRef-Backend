package dto

type ErrorResponse struct {
	// エラーコード（業務エラーコード）
	Code string `json:"code"`
	// エラーメッセージ
	Message string `json:"message"`
}
