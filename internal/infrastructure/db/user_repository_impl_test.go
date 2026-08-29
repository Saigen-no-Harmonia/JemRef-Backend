package db

import (
	"context"
	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------SelectByInternalUid(context.Context, int64)のテスト-------------
func TestUserRepository_SelectByInternalUid_正常(t *testing.T) {
	repo := testUsersSetup(t, "testdata/users_select_001.sql")

	expect, _ := testutil.SelectUserByPk(t, testDb, int64(1001))

	actual, err := repo.SelectByInternalUid(context.Background(), int64(1001))
	require.NoError(t, err)
	assert.Equal(t, expect.InternalUserId, actual.InternalUserId)
	assert.Equal(t, expect.FirebaseUserId, actual.FirebaseUserId)
	assert.Equal(t, expect.Email, actual.Email)
	assert.Equal(t, expect.TermsAgreedDt, actual.TermsAgreedDt)
	assert.Equal(t, expect.TermsVersion, actual.TermsVersion)
	assert.Equal(t, expect.PrivacyPolicyAgreedDt, actual.PrivacyPolicyAgreedDt)
	assert.Equal(t, expect.PrivacyPolicyVersion, actual.PrivacyPolicyVersion)
	assert.Equal(t, expect.DeletedAt, actual.DeletedAt)
}

func TestUserRepository_SelectByInternalUid_論理削除済みは抽出対象外(t *testing.T) {
	repo := testUsersSetup(t, "testdata/users_select_001.sql")

	// 論削ずみのIDを指定
	actual, err := repo.SelectByInternalUid(context.Background(), int64(1003))
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, actual)
}

func TestUserRepository_SelectByInternalUid_対象ユーザなし(t *testing.T) {
	repo := testUsersSetup(t, "testdata/users_select_001.sql")

	// 論削ずみのIDを指定
	actual, err := repo.SelectByInternalUid(context.Background(), int64(9999))
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, actual)
}

// -------------SelectByFirebaseUid(context.Context, string)のテスト-------------
func TestUserRepository_SelectByFirebaseUid_正常(t *testing.T) {
	repo := testUsersSetup(t, "testdata/users_select_001.sql")
	expect, _ := testutil.SelectUserByPk(t, testDb, int64(1001))

	actual, err := repo.SelectByFirebaseUid(context.Background(), "_firebase_uid001_")
	require.NoError(t, err)
	assert.Equal(t, expect.InternalUserId, actual.InternalUserId)
	assert.Equal(t, expect.FirebaseUserId, actual.FirebaseUserId)
	assert.Equal(t, expect.Email, actual.Email)
	assert.Equal(t, expect.TermsAgreedDt, actual.TermsAgreedDt)
	assert.Equal(t, expect.TermsVersion, actual.TermsVersion)
	assert.Equal(t, expect.PrivacyPolicyAgreedDt, actual.PrivacyPolicyAgreedDt)
	assert.Equal(t, expect.PrivacyPolicyVersion, actual.PrivacyPolicyVersion)
	assert.Equal(t, expect.DeletedAt, actual.DeletedAt)
}

func TestUserRepository_SelectByFirebaseUid_論理削除済みは抽出対象外(t *testing.T) {
	repo := testUsersSetup(t, "testdata/users_select_001.sql")

	// 論削ずみのIDを指定
	actual, err := repo.SelectByFirebaseUid(context.Background(), "_firebase_uid003_")
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, actual)
}

func TestUserRepository_SelectByFirebaseUid_対象ユーザなし(t *testing.T) {
	repo := testUsersSetup(t, "testdata/users_select_001.sql")

	// 論削ずみのIDを指定
	actual, err := repo.SelectByFirebaseUid(context.Background(), "_firebase_uid999_")
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, actual)
}

// -------------Create(context.Context, *user.User)のテスト-------------
func TestUserRepository_Create_正常(t *testing.T) {
	repo := testUsersSetup(t, "")
	now := time.Now().Truncate(time.Microsecond)
	termsAgreedDt := time.Now()
	PrivacyPolicyAgreedDt := time.Now()
	u := &user.User{
		FirebaseUserId:        "_firebase_uid_",
		PublicUserId:          "_public_uid_",
		Email:                 "_email_",
		TermsAgreedDt:         &termsAgreedDt,
		TermsVersion:          "_terms_ver_",
		PrivacyPolicyAgreedDt: &PrivacyPolicyAgreedDt,
		PrivacyPolicyVersion:  "_prpl_ver_",
		InsPg:                 "_ins_pg_",
		InsId:                 "_ins_id_",
		UpdPg:                 "_upd_pg_",
		UpdId:                 "_upd_id_",
	}

	beforeCnt := testutil.CountUsers(t, testDb)
	err := repo.Create(
		context.Background(),
		u,
	)
	require.NoError(t, err)
	afterCnt := testutil.CountUsers(t, testDb)
	assert.Equal(t, 1, afterCnt-beforeCnt)

	actual, err := testutil.SelectUserByFirebaseUid(t, testDb, "_firebase_uid_")
	require.NoError(t, err)
	assert.Equal(t, u.FirebaseUserId, actual.FirebaseUserId)
	assert.Equal(t, u.PublicUserId, actual.PublicUserId)
	assert.Equal(t, u.Email, actual.Email)
	assert.WithinDuration(t, *u.TermsAgreedDt, *actual.TermsAgreedDt, time.Microsecond)
	assert.Equal(t, u.TermsVersion, actual.TermsVersion)
	assert.WithinDuration(t, *u.PrivacyPolicyAgreedDt, *actual.PrivacyPolicyAgreedDt, time.Microsecond)
	assert.Equal(t, u.PrivacyPolicyVersion, actual.PrivacyPolicyVersion)
	assert.Equal(t, 0, actual.DelFlg)
	assert.NotZero(t, actual.InternalUserId)
	assert.Nil(t, actual.DeletedAt)
	assert.WithinDuration(t, now, *actual.InsDt, 3*time.Second)
	assert.WithinDuration(t, now, *actual.UpdDt, 3*time.Second)
}

// -------------Delete(context.Context, int64)のテスト-------------
func TestUserRepository_Delete_正常(t *testing.T) {
	repo := testUsersSetup(t, "testdata/users_delete_001.sql")

	before, err := testutil.SelectUserByPk(t, testDb, int64(1001))
	require.NoError(t, err)

	now := time.Now().Truncate(time.Microsecond).UTC()
	err = repo.Delete(context.Background(), int64(1001))
	require.NoError(t, err)

	after, err := testutil.SelectUserByPk(t, testDb, int64(1001))
	require.NoError(t, err)

	assert.Equal(t, before.InternalUserId, after.InternalUserId)
	assert.Equal(t, before.PublicUserId, after.PublicUserId)
	assert.Equal(t, before.FirebaseUserId, after.FirebaseUserId)
	// メアドは冒頭と末尾文字数を検証（細かい期待値が取りづらいので）
	assert.Regexp(t, `^__deleted__.+__\d{14}$`, after.Email)
	assert.Equal(t, before.TermsAgreedDt, after.TermsAgreedDt)
	assert.Equal(t, before.TermsVersion, after.TermsVersion)
	assert.Equal(t, before.PrivacyPolicyAgreedDt, after.PrivacyPolicyAgreedDt)
	assert.Equal(t, before.PrivacyPolicyVersion, after.PrivacyPolicyVersion)
	assert.WithinDuration(t, now, *after.DeletedAt, 3*time.Second)
	assert.Equal(t, 1, after.DelFlg)
	assert.Equal(t, before.InsPg, after.InsPg)
	assert.Equal(t, before.InsId, after.InsId)
	assert.Equal(t, "MEM-API-004", after.UpdPg)
	assert.Equal(t, "system", after.UpdId)
	assert.WithinDuration(t, now, *after.UpdDt, 3*time.Second)
}

func TestUserRepository_Delete_対象ユーザなし(t *testing.T) {
	repo := testUsersSetup(t, "testdata/users_delete_001.sql")

	before, err := testutil.SelectUserByPk(t, testDb, int64(1001))
	require.NoError(t, err)

	err = repo.Delete(context.Background(), int64(1000))
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrNotFound)

	after, err := testutil.SelectUserByPk(t, testDb, int64(1001))
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

// -------------UpdateUserAgreement(context.Context, *user.User)のテスト-------------
func TestUserRepository_UpdateUserAgreement_正常(t *testing.T) {
	repo := testUsersSetup(t, "testdata/users_update_001.sql")
	before, err := testutil.SelectUserByPk(t, testDb, int64(1001))
	require.NoError(t, err)

	now := time.Now().Truncate(time.Microsecond).UTC()
	u := *before
	u.TermsAgreedDt = &now
	u.TermsVersion = "_terms_updated_"
	u.PrivacyPolicyAgreedDt = &now
	u.PrivacyPolicyVersion = "_prpl_updated_"

	beforeUpd := time.Now().UTC()
	cnt, err := repo.UpdateUserAgreement(context.Background(), &u)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, cnt)

	after, err := testutil.SelectUserByPk(t, testDb, int64(1001))
	require.NoError(t, err)
	// 更新
	assert.Equal(t, u.TermsAgreedDt, after.TermsAgreedDt)
	assert.Equal(t, u.TermsVersion, after.TermsVersion)
	assert.Equal(t, u.PrivacyPolicyAgreedDt, after.PrivacyPolicyAgreedDt)
	assert.Equal(t, u.PrivacyPolicyVersion, after.PrivacyPolicyVersion)
	assert.Equal(t, "MEM-API-005", after.UpdPg)
	assert.Equal(t, "system", after.UpdId)
	assert.WithinDuration(t, beforeUpd, *after.UpdDt, 3*time.Second)
	// 非更新
	assert.Equal(t, before.InternalUserId, after.InternalUserId)
	assert.Equal(t, before.PublicUserId, after.PublicUserId)
	assert.Equal(t, before.FirebaseUserId, after.FirebaseUserId)
	assert.Equal(t, before.DelFlg, after.DelFlg)
	assert.Equal(t, before.InsPg, after.InsPg)
	assert.Equal(t, before.InsId, after.InsId)
	assert.Equal(t, before.InsDt, after.InsDt)
}

func TestUserRepository_UpdateUserAgreement_対象ユーザなし(t *testing.T) {
	repo := testUsersSetup(t, "testdata/users_update_001.sql")
	before, err := testutil.SelectUserByPk(t, testDb, int64(1001))
	require.NoError(t, err)

	now := time.Now().Truncate(time.Microsecond).UTC()
	u := *before
	u.InternalUserId = int64(1000) // 存在しないUID
	u.TermsAgreedDt = &now
	u.TermsVersion = "_terms_updated_"
	u.PrivacyPolicyAgreedDt = &now
	u.PrivacyPolicyVersion = "_prpl_updated_"

	cnt, err := repo.UpdateUserAgreement(context.Background(), &u)
	assert.NoError(t, err)
	assert.EqualValues(t, 0, cnt)

	after, err := testutil.SelectUserByPk(t, testDb, int64(1001))
	require.NoError(t, err)
	// 更新
	assert.Equal(t, before, after)
}

// testUsersSetup DBクリーンアップ＋テスト用Fixture投入を済ませたuserRepoを返却する
func testUsersSetup(t *testing.T, fixturePath string) *UserRepositoryImpl {
	testutil.CleanUpTables(t, testDb)

	if fixturePath != "" {
		err := testutil.LoadFixture(
			testDb,
			fixturePath,
		)
		require.NoError(t, err)
	}

	return NewUserRepositoryImpl(testDb)
}
