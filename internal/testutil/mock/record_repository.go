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

	SelectByPkFunc func(
		ctx context.Context,
		tx repository.DBTX,
		recordId string,
		internalUid int64,
	) (*record.UserRecordEntry, error)

	CountActiveChildrenFunc func(
		ctx context.Context,
		tx repository.DBTX,
		parentRecordId string,
	) (int, error)

	LogicalDeleteFunc func(
		ctx context.Context,
		tx repository.DBTX,
		recordId string,
	) error

	Called         int
	CreateCalled   int
	CreatedEntries []record.UserRecordEntry

	SelectByPkCalled          int
	CountActiveChildrenCalled int
	LogicalDeleteCalled       int
	LastSelectByPkRecordId    string
	LastSelectByPkInternalUid int64
	// CountedRecordIds CountActiveChildrenに渡された書誌IDを呼び出し順に保持する
	CountedRecordIds []string
	// DeletedRecordIds LogicalDeleteに渡された書誌IDを呼び出し順に保持する
	DeletedRecordIds []string
}

func (m *MockRecordRepository) SelectByPk(
	ctx context.Context,
	tx repository.DBTX,
	recordId string,
	internalUid int64,
) (*record.UserRecordEntry, error) {
	m.Called++
	m.SelectByPkCalled++
	m.LastSelectByPkRecordId = recordId
	m.LastSelectByPkInternalUid = internalUid
	return m.SelectByPkFunc(ctx, tx, recordId, internalUid)
}

func (m *MockRecordRepository) CountActiveChildren(
	ctx context.Context,
	tx repository.DBTX,
	parentRecordId string,
) (int, error) {
	m.Called++
	m.CountActiveChildrenCalled++
	m.CountedRecordIds = append(m.CountedRecordIds, parentRecordId)
	return m.CountActiveChildrenFunc(ctx, tx, parentRecordId)
}

func (m *MockRecordRepository) LogicalDelete(
	ctx context.Context,
	tx repository.DBTX,
	recordId string,
) error {
	m.Called++
	m.LogicalDeleteCalled++
	m.DeletedRecordIds = append(m.DeletedRecordIds, recordId)
	return m.LogicalDeleteFunc(ctx, tx, recordId)
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
