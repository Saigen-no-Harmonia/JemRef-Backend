package dto

// [GEN-API-001] ユーザ規約参照 リクエスト
type GetPoliciesRequest struct {
	// 規約ID
	PolicyId string `json:"policy_id" binding:"required" enums:"terms,privacy_policy"`
}

// [GEN-API-001] ユーザ規約参照 レスポンス
type GetPoliciesResponse struct {
	// 規約ID
	PolicyId string `json:"policy_id" binding:"required" enums:"terms,privacy_policy"`
	// ラベル（表示名）
	Label string `json:"label" binding:"required" example:"利用規約"`
	// 最新バージョン
	LatestVersion string `json:"latest_version" binding:"required" example:"1.2"`
	// 更新日時（yyyy-mm-dd hh:MM:ss）
	UpdateDate string `json:"update_date" binding:"required"  example:"2026-09-30 10:00:00"`
	// 規約本文（改行文字入りテキスト）
	Content string `json:"content" binding:"required" example:"第１条〜¥n第２条〜〜2026年9月30日"`
}
