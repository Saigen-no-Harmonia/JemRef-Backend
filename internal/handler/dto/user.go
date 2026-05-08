package dto

import "jemref_go/internal/domain/user"

type CreateUserResponse struct {
	// 公開用ユーザID(28文字固定ULID）
	PublicUserId string `json:"user_id" binding:"required"`
}

type GetUserAgreementsResponse struct {
	Agreements []UserAgreementResponse `json:"agreements"`
}

type UserAgreementResponse struct {
	// 規約ID
	PolicyId string `json:"policy_id" binding:"required" required:"true" example:"privacy_policy"`
	// 規約ラベル（表示名）
	Label string `json:"label" binding:"required" example:"プライバシーポリシー"`
	// 最新バージョン
	LatestVersion string `json:"latest_version" binding:"required" example:"1.2"`
	// 同意バージョン
	AgreedVersion string `json:"agreed_version" binding:"required" example:"1.0"`
	// 同意ステータス
	Status user.PolicyAgreementStatus `json:"status" binding:"required" example:"update_required"`
}

type UpdateUserAgreementsRequest struct {
	Policies []AgreementPolicyRequest `json:"policies" binding:"required"`
}

type AgreementPolicyRequest struct {
	// 規約ID
	PolicyId string `json:"policy_id" binding:"required" example:"privacy_policy"`
	// 規約バージョン
	Version string `json:"version" binding:"required" example:"1.4"`
}

type UserLoginResponse struct {
	// 公開用ユーザID（28文字固定ULID）
	UserId string `json:"user_id" binding:"required" example:"001JSUFP94SNN~"`
}
