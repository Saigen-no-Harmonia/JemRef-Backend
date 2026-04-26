package mock

import "jemref_go/internal/domain"

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

func (m *RecordRepositoryMock) Create(ur *domain.UserRecordEntry) error {
	// mockなので保存はしない
	ur.UserId = m.UserId
	ur.MainTitle = m.Title
	return nil
}
