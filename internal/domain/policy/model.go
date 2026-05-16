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
}
