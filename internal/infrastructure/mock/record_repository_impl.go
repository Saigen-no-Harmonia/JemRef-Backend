package mock

import "jemref_go/internal/domain/record"

// TODO 仮実装用のモック。実装時にディレクトリごと削除すること

type RecordRepositoryMock struct {
	UserId string
	Title  string
}

func NewRecordRepositoryMock() *RecordRepositoryMock {
	return &RecordRepositoryMock{
		UserId: "sample_id",
		Title:  "sample_title",
	}
}

func (m *RecordRepositoryMock) Create(ur *record.UserRecordEntry) error {
	// mockなので保存はしない
	ur.PublicUserId = m.UserId
	ur.MainTitle = m.Title
	return nil
}
