package record

type ReadStatus string

const (
	ReadStatusRead        = "read"
	ReadStatusUnread      = "unread"
	ReadStatusReading     = "reading"
	ReadStatusPertialRead = "partial_read"
)

func (s ReadStatus) IsValid() bool {
	switch s {
	case ReadStatusRead,
		ReadStatusUnread,
		ReadStatusReading,
		ReadStatusPertialRead:
		return true
	default:
		return false
	}
}
