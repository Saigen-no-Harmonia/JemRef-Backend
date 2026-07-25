package policy

import "time"

type Policy struct {
	// ID
	Id string
	// バージョン
	Version string
	// 名前（表示名）
	Name string
	// 内容（改行文字入り本文）
	Content string
	// 発行日
	EffectiveDate time.Time
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
