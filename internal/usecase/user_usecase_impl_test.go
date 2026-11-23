package usecase

import (
	"context"

	"jemref_go/internal/domain/policy"
	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil/mock"
)

// UserUsecaseImplのテストは、メソッドごとに user_usecase_<メソッド>_test.go に置く。
// このファイルには、それらが共有する定数とmockのビルダを置く。

const (
	// userTestUid テスト用の内部ユーザID
	userTestUid = int64(1001)
	// userTestFirebaseUid テスト用のFirebase UID
	userTestFirebaseUid = "_firebase_uid_"
)

// policyResult 規約の取得結果(SelectByPrimaryKey / SelectLatestById の戻り値)
type policyResult struct {
	policy *policy.Policy
	err    error
}

// newGeneralRepoMock 規約IDごとに決めた結果を返すMockGeneralRepositoryを返す
//
// SelectByPrimaryKey と SelectLatestById のどちらも同じ表を引く(バージョンは見ない)。
// 表にない規約IDが渡された場合は repository.ErrTestUnexpected を返す。
func newGeneralRepoMock(results map[string]policyResult) *mock.MockGeneralRepository {
	lookup := func(id string) (*policy.Policy, error) {
		if r, ok := results[id]; ok {
			return r.policy, r.err
		}
		return nil, repository.ErrTestUnexpected
	}

	return &mock.MockGeneralRepository{
		SelectByPrimaryKeyFunc: func(ctx context.Context, id string, version string) (*policy.Policy, error) {
			return lookup(id)
		},
		SelectLatestByIdFunc: func(ctx context.Context, id string) (*policy.Policy, error) {
			return lookup(id)
		},
	}
}

// newSelectUserRepoMock SelectByInternalUidで指定したユーザ(またはエラー)を返すMockUserRepositoryを返す
//
// 他のメソッドは呼び出し側で必要に応じて設定する。
func newSelectUserRepoMock(u *user.User, err error) *mock.MockUserRepository {
	return &mock.MockUserRepository{
		SelectByInternalUidFunc: func(ctx context.Context, uid int64) (*user.User, error) {
			if err != nil {
				return nil, err
			}
			return u, nil
		},
	}
}
