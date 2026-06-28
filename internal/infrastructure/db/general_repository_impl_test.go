package db

// GeneralRepository実装のユニットテスト
// 下記のメソッドを検証する
// SelectLatestById(context.Context, String)

import (
	"context"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------SelectLatestById(context.Context, String)のテスト-------------
func TestGeneralRepository_SelectLatestById_正常(t *testing.T) {

	repo := testPoliciesSetup(t, "testdata/policies_select_success_001.sql")
	actual, err := repo.SelectLatestById(
		context.Background(),
		"P001",
	)
	require.NoError(t, err)
	assert.Equal(t, "P001", actual.Id)
	assert.Equal(t, "test_0.01", actual.Version)
	assert.Equal(t, "_test_name_0.01_", actual.Name)
	assert.Equal(t, "_test_content_0.01_", actual.Content)
	assert.Equal(t, "2026-05-10", actual.EffectiveDate.Format("2006-01-02"))
}

func TestGeneralRepository_SelectLatestById_準正常_該当レコードなし(t *testing.T) {

	repo := testPoliciesSetup(t, "testdata/policies_select_success_001.sql")
	actual, err := repo.SelectLatestById(
		context.Background(),
		"P999",
	)
	require.ErrorIs(t, err, repository.ErrPolicyNotFound)
	assert.Nil(t, actual)
}

func TestGeneralRepository_SelectLatestById_準正常_引数なし(t *testing.T) {

	repo := testPoliciesSetup(t, "testdata/policies_select_success_001.sql")
	actual, err := repo.SelectLatestById(
		context.Background(),
		"",
	)
	require.ErrorIs(t, err, repository.ErrPolicyNotFound)
	assert.Nil(t, actual)
}

// testPoliciesSetup Policiesテーブルのトランケートとfixtureの投入
func testPoliciesSetup(t *testing.T, fixturePath string) *GeneralRepositoryImpl {
	err := testutil.TruncateTable(
		testDb,
		"m_policies",
	)
	require.NoError(t, err)

	err = testutil.LoadFixture(
		testDb,
		fixturePath,
	)
	require.NoError(t, err)

	return NewGeneralRepositoryImpl(testDb)
}
