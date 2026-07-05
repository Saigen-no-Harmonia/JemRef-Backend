package dto

type ErrorResponse struct {
	// エラーコード（業務エラーコード）
	Code string `json:"code"`
	// エラーメッセージ
	Message string `json:"message"`
}

func NewErrorResponse(code string, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}
