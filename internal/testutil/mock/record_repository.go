package mock

import (
	"context"
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
)

type MockRecordRepository struct {
	CreateFunc func(
		ctx context.Context,
		tx repository.DBTX,
		ur record.UserRecordEntry,
	) error

	Called         int
	CreateCalled   int
	CreatedEntries []record.UserRecordEntry
}

func (m *MockRecordRepository) Create(
	ctx context.Context,
	tx repository.DBTX,
	ur record.UserRecordEntry,
) error {
	m.Called++
	m.CreateCalled++
	m.CreatedEntries = append(m.CreatedEntries, ur)
	return m.CreateFunc(ctx, tx, ur)
}

var _ repository.RecordRepository = (*MockRecordRepository)(nil)
