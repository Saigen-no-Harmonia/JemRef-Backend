package dto

// [REF-API-001] 書誌一覧参照 リクエスト
//
// クエリパラメータで受け取る。必須項目はない。
// 検索条件は指定されたものをANDで組み合わせ、空文字(検索語は前後の空白を除いて空)は未指定として扱う。
// defaultタグはSwaggerの表記用で、実際のデフォルト値はhandlerの定数(defaultRecordsXxx)が正。
// 値を変える場合は両方をそろえること。
type GetRecordsRequest struct {
	// 貢献者名（検索用：書誌または収録物の貢献者名の部分一致）
	ContributorName string `form:"contributor_name" example:"木村"`
	// タイトル（検索用：書誌または収録物の主題・副題の部分一致）
	Title string `form:"title" example:"音楽"`
	// 出版年（検索用：西暦4桁）
	PublishYear string `form:"publish_year" example:"2026"`
	// 並びかえキー (出版年、登録日)
	SortKey string `form:"sort_key" enums:"publish_year,insert_date" default:"publish_year" example:"insert_date"`
	// 並び順（asc または desc）
	SortType string `form:"sort_type" enums:"asc,desc" default:"asc" example:"desc"`
	// リミット（1〜100）
	Limit *int `form:"limit" minimum:"1" maximum:"100" default:"20" example:"50"`
	// オフセット
	Offset *int `form:"offset" minimum:"0" default:"0" example:"20"`
}

// [REF-API-001] 書誌一覧参照 レスポンス
type GetRecordsResponse struct {
	// 検索条件に一致する総件数（limit/offset適用前）
	Count int `json:"count" binding:"required" example:"1"`
	// 書誌一覧（該当なしの場合は空配列）
	Records *[]RecordListItemResponse `json:"records" binding:"required"`
}

// [REF-API-002] 書誌詳細参照　レスポンス
type GetRecordResponse struct {
	// ユーザ書誌ID(36桁固定文字列)
	RecordId string `json:"record_id" binding:"required" example:"AHIF98DJ0SJUFY874H..."`
	// 書誌詳細情報
	RecordDetailItem `binding:"required"`
	// 収録物詳細情報
	ContainerDetailItem *ContainerDetailItem `json:"container"`
}

// [REF-API-003] 書誌情報登録　リクエスト
type CreateRecordRequest struct {
	// 書誌詳細情報
	RecordDetailItem `binding:"required"`
	// 収録物詳細情報
	ContainerDetailItem *ContainerDetailItem `json:"container"`
}

// [REF-API-003] 書誌情報登録　レスポンス
type CreateRecordResponse struct {
	// 公開用ユーザID（26桁固定ULID）
	PublicUserId string `json:"user_id" binding:"required"`
	// 公開用書誌ID
	PublicRecordId string `json:"record_id" binding:"required"`
}

// [REF-API-004] 書誌情報更新　リクエスト
type UpdateRecordRequest struct {
	// 書誌詳細情報
	RecordDetailItem `binding:"required"`
	// 収録物詳細情報
	ContainerDetailItem *ContainerDetailItem `json:"container"`
}

// レコード関係共通DTO-------------------------------------------
// 書誌リスト用レコード情報
type RecordListItemResponse struct {
	// ユーザ書誌ID(36桁固定文字列)
	RecordId string `json:"record_id" binding:"required" example:"AHIF98DJ0SJUFY874H..."`
	// 書誌タイプ
	Type string `json:"type" binding:"required" enums:"authored_book,journal_article,contributed_book_article"`
	// 主題
	MainTitle string `json:"main_title" binding:"required" example:"第1章 「学校」制度の境界線"`
	// 副題
	SubTitle *string `json:"sub_title" example:"その形成と展開"`
	// ページ数
	PageRange *string `json:"page_range" example:"15-46"`
	// 出版社/出版者
	Publisher *string `json:"publisher" example:"東京大学出版会"`
	// 出版年（月日）
	PublicationDate *string `json:"publication_date" example:"2022-11-30"`
	// 既読ステータス
	ReadStatus *string `json:"read_status" enums:"unread,reading,partially_read,read"`
	// 貢献者情報
	Contributors []ContributorInfo `json:"contributors"`
	// 収録物情報
	ContainerListItem *ContainerListItemResponse `json:"container"`
}

// 書誌リスト用収録物情報
type ContainerListItemResponse struct {
	// 収録物主題
	MainTitle string `json:"main_title" example:"境界線の学校史"`
	// 収録物副題
	SubTitle *string `json:"sub_title" example:"戦後日本の学校か社会の周辺と周縁"`
	// 巻
	Volume *string `json:"volume" example:"日本教育史講座第1巻"`
	// 号
	Issue *string `json:"issue" example:"第11号"`
	// 収録物貢献者情報
	Contributors []ContributorInfo `json:"contributors"`
}

// 書誌詳細用レコード情報
type RecordDetailItem struct {
	// 書誌タイプ
	Type string `json:"type" binding:"required" enums:"authored_book,journal_article,contributed_book_article"`
	// 主題
	MainTitle string `json:"main_title" binding:"required" example:"第1章 「学校」制度の境界線"`
	// 副題
	SubTitle *string `json:"sub_title" example:"その形成と展開"`
	// ページ数
	PageRange *string `json:"page_range" example:"15-46"`
	// 出版社/出版者
	Publisher *string `json:"publisher" example:"東京大学出版会"`
	// 出版年（月日）
	PublicationDate *string `json:"publication_date" example:"2022-11-30"`
	// 既読ステータス
	ReadStatus *string `json:"read_status" enums:"unread,reading,partially_read,read"`
	// 貢献者情報
	Contributors []ContributorInfo `json:"contributors" binding:"dive"`
	// メモ情報
	Memos []MemoInfo `json:"memos" binding:"dive"`
	// URL情報
	Urls []UrlInfo `json:"urls" binding:"dive"`
	// オプション書誌項目値情報
	OptionValues *OptionValues `json:"option_values"`
}

// 書誌詳細用収録物情報
type ContainerDetailItem struct {
	// 主題
	MainTitle string `json:"main_title" binding:"required" example:"第1章 「学校」制度の境界線"`
	// 副題
	SubTitle *string `json:"sub_title" example:"その形成と展開"`
	// 巻
	Volume *string `json:"volume" example:"日本教育史講座第1巻"`
	// 号
	Issue *string `json:"issue" example:"第11号"`
	// 既読ステータス
	ReadStatus *string `json:"read_status" enums:"unread,reading,partially_read,read"`
	// 貢献者情報
	Contributors []ContributorInfo `json:"contributors" binding:"dive"`
	// オプション書誌項目値情報
	OptionValues *OptionValues `json:"option_values"`
}

// 貢献者情報
type ContributorInfo struct {
	// 貢献者名（人物名、研究会名など）
	Name string `json:"name" binding:"required" example:"木村元"`
	// 役割
	Role string `json:"role" binding:"required" enums:"author,interpreter,compiler,compiler_author"`
}

// メモ情報
type MemoInfo struct {
	// ラベル
	Label *string `json:"label" example:"要約"`
	// メモの内容
	Content string `json:"content" binding:"required" example:"学校の戦後史。少し難しい。"`
}

// URL情報
type UrlInfo struct {
	// ラベル
	Label *string `json:"label" example:"オンラインPDFのリンク"`
	// URL
	Value string `json:"value" binding:"required" example:"https://example.com/sample"`
}

// オプション書誌項目値情報
type OptionValues struct {
	Isbn          *string `json:"isbn" example:"9784130513555"`
	Edition       *string `json:"edition" example:"第２版"`
	OriginalTitle *string `json:"original_title" example:"School History of Boundary"`
}
