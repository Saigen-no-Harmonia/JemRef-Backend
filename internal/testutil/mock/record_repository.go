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

	CountActiveArticlesFunc func(
		ctx context.Context,
		tx repository.DBTX,
		containerRecordId string,
	) (int, error)

	LogicalDeleteFunc func(
		ctx context.Context,
		tx repository.DBTX,
		recordId string,
	) error

	SelectComponentsFunc func(
		ctx context.Context,
		tx repository.DBTX,
		entries []*record.UserRecordEntry,
	) error

	Called         int
	CreateCalled   int
	CreatedEntries []record.UserRecordEntry

	SelectByPkCalled          int
	CountActiveArticlesCalled int
	LogicalDeleteCalled       int
	SelectComponentsCalled    int
	LastSelectByPkRecordId    string
	LastSelectByPkInternalUid int64
	// SelectedRecordIds SelectByPkに渡された書誌IDを呼び出し順に保持する
	SelectedRecordIds []string
	// ComponentTargetIds SelectComponentsに渡された書誌IDを呼び出しごとに保持する
	// nilの要素は "" として記録し、収録物なしで呼ばれたことを検証できるようにする。
	ComponentTargetIds [][]string
	// CountedRecordIds CountActiveArticlesに渡された書誌IDを呼び出し順に保持する
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
	m.SelectedRecordIds = append(m.SelectedRecordIds, recordId)
	return m.SelectByPkFunc(ctx, tx, recordId, internalUid)
}

func (m *MockRecordRepository) CountActiveArticles(
	ctx context.Context,
	tx repository.DBTX,
	containerRecordId string,
) (int, error) {
	m.Called++
	m.CountActiveArticlesCalled++
	m.CountedRecordIds = append(m.CountedRecordIds, containerRecordId)
	return m.CountActiveArticlesFunc(ctx, tx, containerRecordId)
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

func (m *MockRecordRepository) SelectComponents(
	ctx context.Context,
	tx repository.DBTX,
	entries []*record.UserRecordEntry,
) error {
	m.Called++
	m.SelectComponentsCalled++

	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		if e == nil {
			ids = append(ids, "")
			continue
		}
		ids = append(ids, e.RecordEntryId)
	}
	m.ComponentTargetIds = append(m.ComponentTargetIds, ids)

	return m.SelectComponentsFunc(ctx, tx, entries)
}

var _ repository.RecordRepository = (*MockRecordRepository)(nil)
