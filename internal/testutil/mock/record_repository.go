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

	UpdateFunc func(
		ctx context.Context,
		tx repository.DBTX,
		ur record.UserRecordEntry,
	) error

	SelectComponentsFunc func(
		ctx context.Context,
		tx repository.DBTX,
		entries []*record.UserRecordEntry,
	) error

	CountByConditionFunc func(
		ctx context.Context,
		tx repository.DBTX,
		cond repository.RecordListCondition,
	) (int, error)

	SelectByConditionFunc func(
		ctx context.Context,
		tx repository.DBTX,
		cond repository.RecordListCondition,
	) ([]*record.UserRecordEntry, error)

	SelectByPksFunc func(
		ctx context.Context,
		tx repository.DBTX,
		recordIds []string,
		internalUid int64,
	) ([]*record.UserRecordEntry, error)

	SelectContributorRelationsFunc func(
		ctx context.Context,
		tx repository.DBTX,
		entries []*record.UserRecordEntry,
	) error

	Called         int
	CreateCalled   int
	UpdateCalled   int
	CreatedEntries []record.UserRecordEntry
	// UpdatedEntries Updateに渡された書誌を呼び出し順に保持する
	UpdatedEntries []record.UserRecordEntry

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

	CountByConditionCalled           int
	SelectByConditionCalled          int
	SelectByPksCalled                int
	SelectContributorRelationsCalled int
	// CountedConditions CountByConditionに渡された検索条件を呼び出し順に保持する
	CountedConditions []repository.RecordListCondition
	// SelectedConditions SelectByConditionに渡された検索条件を呼び出し順に保持する
	SelectedConditions []repository.RecordListCondition
	// SelectedPkIds SelectByPksに渡された書誌IDを呼び出しごとに保持する
	SelectedPkIds [][]string
	// LastSelectByPksInternalUid SelectByPksに最後に渡された内部用ユーザID
	LastSelectByPksInternalUid int64
	// ContributorTargetIds SelectContributorRelationsに渡された書誌IDを呼び出しごとに保持する
	// nilの要素は "" として記録する(ComponentTargetIdsと同じ)。
	ContributorTargetIds [][]string
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
	m.ComponentTargetIds = append(m.ComponentTargetIds, entryIdsOf(entries))

	return m.SelectComponentsFunc(ctx, tx, entries)
}

func (m *MockRecordRepository) CountByCondition(
	ctx context.Context,
	tx repository.DBTX,
	cond repository.RecordListCondition,
) (int, error) {
	m.Called++
	m.CountByConditionCalled++
	m.CountedConditions = append(m.CountedConditions, cond)
	return m.CountByConditionFunc(ctx, tx, cond)
}

func (m *MockRecordRepository) SelectByCondition(
	ctx context.Context,
	tx repository.DBTX,
	cond repository.RecordListCondition,
) ([]*record.UserRecordEntry, error) {
	m.Called++
	m.SelectByConditionCalled++
	m.SelectedConditions = append(m.SelectedConditions, cond)
	return m.SelectByConditionFunc(ctx, tx, cond)
}

func (m *MockRecordRepository) SelectByPks(
	ctx context.Context,
	tx repository.DBTX,
	recordIds []string,
	internalUid int64,
) ([]*record.UserRecordEntry, error) {
	m.Called++
	m.SelectByPksCalled++
	// 呼び出し側がスライスを使い回しても記録が変わらないよう複製する
	m.SelectedPkIds = append(m.SelectedPkIds, append([]string(nil), recordIds...))
	m.LastSelectByPksInternalUid = internalUid
	return m.SelectByPksFunc(ctx, tx, recordIds, internalUid)
}

func (m *MockRecordRepository) SelectContributorRelations(
	ctx context.Context,
	tx repository.DBTX,
	entries []*record.UserRecordEntry,
) error {
	m.Called++
	m.SelectContributorRelationsCalled++
	m.ContributorTargetIds = append(m.ContributorTargetIds, entryIdsOf(entries))
	return m.SelectContributorRelationsFunc(ctx, tx, entries)
}

// entryIdsOf 渡された書誌のIDを順に返す。nilの要素は "" として記録する
func entryIdsOf(entries []*record.UserRecordEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, e := range entries {
		if e == nil {
			ids = append(ids, "")
			continue
		}
		ids = append(ids, e.RecordEntryId)
	}

	return ids
}

func (m *MockRecordRepository) Update(
	ctx context.Context,
	tx repository.DBTX,
	ur record.UserRecordEntry,
) error {
	m.Called++
	m.UpdateCalled++
	m.UpdatedEntries = append(m.UpdatedEntries, ur)
	return m.UpdateFunc(ctx, tx, ur)
}

var _ repository.RecordRepository = (*MockRecordRepository)(nil)
