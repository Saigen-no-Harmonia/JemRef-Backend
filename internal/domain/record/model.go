package record

import (
	"time"
)

type UserRecordEntry struct {
	// ユーザ書誌ID
	RecordEntryId string
	// 内部用ユーザID
	InternalUid int64
	// 書誌タイプID
	RecordType RecordType
	// 主題
	MainTitle string
	// 副題
	SubTitle *string
	// ページ範囲
	PageRange *string
	// 巻
	Volume *string
	// 号
	Issue *string
	// 出版者
	Publisher *string
	// 出版年月日
	PublicationDate *time.Time
	// 出版年月日精度
	DatePrecision *DatePrecision
	// 読了ステータス
	ReadStatus *ReadStatus
	// コンテナ書誌ID
	ParentId *string
	// 書誌マスタID
	MRecordId *string
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

	Contributors  []ContributorEntry
	Contributions []ContributionEntry
	Urls          []UrlEntry
	Memos         []MemoEntry
	OptionValues  []OptionValue
}

type ContributorEntry struct {
	// ユーザ入力貢献者ID
	ContributorEntryId int64
	// ユーザ入力貢献者名
	RawName string
	// 内部用ユーザID
	InternalUid int64
	// 貢献者マスタID
	MContributorId *string
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

type ContributionEntry struct {
	// ユーザ入力書誌貢献ID
	ContributionEntryId int64
	// ユーザ入力書誌ID
	RecordEntryId string
	// ユーザ入力貢献者ID
	ContributorEntryId int64
	// 役割
	Role ContributorRole
	// 表示順
	DisplayNumber int
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

type UrlEntry struct {
	// ユーザ入力URL ID
	UrlEntryId int64
	// ユーザ入力書誌ID
	RecordEntryId string
	// 表示順
	DisplayNumber int
	// ユーザ入力見出し
	Label *string
	// ユーザ入力URL
	Url string
	// 内部用ユーザID
	InternalUid int64
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

type MemoEntry struct {
	// ユーザ入力メモID
	MemoEntryId int64
	// ユーザ入力書誌ID
	RecordEntryId string
	// 表示順
	DisplayNumber int
	// 見出し
	Title *string
	// 内容
	Content string
	// 内部用ユーザID
	InternalUid int64
	// 公開許諾レベル
	ShareLevel int
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

type OptionValue struct {
	// 書誌項目入力値ID
	OptionValueEntryId int64
	// ユーザ入力書誌ID
	RecordEntryId string
	// 入力項目
	RecordField RecordField
	// 表示順
	DisplayNumber int
	// 入力値
	RawValue string
	// 内部用ユーザID
	InternalUid int64
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
