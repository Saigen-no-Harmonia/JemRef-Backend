package user

import "time"

type User struct {
	// 内部用ユーザID
	InternalUserId int64
	// 公開用ユーザID
	PublicUserId string
	// FirebaseユーザID
	FirebaseUserId string
	// メールアドレス
	Email string
	// ユーザ規約同意日時
	TermsAgreedDt *time.Time
	// ユーザ規約同意バージョン
	TermsVersion string
	// プライバシーポリシー同意日時
	PrivacyPolicyAgreedDt *time.Time
	PrivacyPolicyVersion  string
	DeletedAt             *time.Time
	// 削除フラグ
	DelFlg int
	// 登録プログラム
	InsPg string
	// 登録者識別ID
	InsId string
	// 登録日
	InsDt *time.Time
	// 更新プログラム
	UpdPg string
	// 更新者識別ID
	UpdId string
	// 更新日
	UpdDt *time.Time
}
