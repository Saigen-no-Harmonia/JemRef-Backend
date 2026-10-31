package record

type ReadStatus string

const (
	ReadStatusRead          ReadStatus = "read"
	ReadStatusUnread        ReadStatus = "unread"
	ReadStatusReading       ReadStatus = "reading"
	ReadStatusPartiallyRead ReadStatus = "partially_read"
)

const (
	ReadStatusIdRead          = 1
	ReadStatusIdUnread        = 2
	ReadStatusIdReading       = 3
	ReadStatusIdPartiallyRead = 4
)

var statusToIdMap = map[ReadStatus]int{
	ReadStatusRead:          ReadStatusIdRead,
	ReadStatusUnread:        ReadStatusIdUnread,
	ReadStatusReading:       ReadStatusIdReading,
	ReadStatusPartiallyRead: ReadStatusIdPartiallyRead,
}

var idToStatusMap = map[int]ReadStatus{
	ReadStatusIdRead:          ReadStatusRead,
	ReadStatusIdUnread:        ReadStatusUnread,
	ReadStatusIdReading:       ReadStatusReading,
	ReadStatusIdPartiallyRead: ReadStatusPartiallyRead,
}

// GetId 既読ステータスのIDを返却する
func (s ReadStatus) GetId() int {
	id, ok := statusToIdMap[s]

	if !ok {
		panic("invalid status")
	}

	return id
}

// GetNullableId 未設定(nil)ならnilを返す。DBのNULL登録用。
func (s *ReadStatus) GetNullableId() any {
	if s == nil {
		return nil
	}
	return s.GetId()
}

// ReadStatusFromId 既読ステータスIDをもとに既読ステータスを返却する
func ReadStatusFromId(id int) ReadStatus {
	readStatus, ok := idToStatusMap[id]

	if !ok {
		panic("invalid status id")
	}

	return readStatus
}

// IsValid 既読ステータスが正しいものであればtrueを返却する
func (s ReadStatus) IsValid() bool {
	switch s {
	case ReadStatusRead,
		ReadStatusUnread,
		ReadStatusReading,
		ReadStatusPartiallyRead:
		return true
	default:
		return false
	}
}
