package dto

// [GEN-API-001] ユーザ規約参照 レスポンス
type GetPoliciesResponse struct {
	// 規約タイプ
	PolicyType string `json:"policy_type" binding:"required" enums:"terms,privacy_policy"`
	// ラベル（表示名）
	Label string `json:"label" binding:"required" example:"利用規約"`
	// 最新バージョン
	LatestVersion string `json:"latest_version" binding:"required" example:"1.2"`
	// 発効日（yyyy-mm-dd）
	EffectiveDate string `json:"effective_date" binding:"required"  example:"2026-09-30"`
	// 規約本文（改行文字入りテキスト）
	Content string `json:"content" binding:"required" example:"第１条〜¥n第２条〜〜2026年9月30日"`
}
