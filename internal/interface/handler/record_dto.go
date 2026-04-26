package handler

type CreateRecordRequest struct {
	Record RecordRequest `json:"record" binding:"required"`
}

// 書誌情報登録リクエスト
type RecordRequest struct {
	UserId     string `json:"user_id"`
	RecordType string `json:"record_type"`
	MainTitle  string `json:"main_title"`

	// 著者
	Authors []AuthorRequest `json:"authors"`

	// 編者情報
	Containor *RecordRequest `json:"contenor,omitempty"`

	// その他
	OtherDetails map[string]any `json:"other_details"`
}

type AuthorRequest struct {
	Priority int    `json:"priority"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type RecordResponse struct {
	UserId string `json:"user_id"`
}
