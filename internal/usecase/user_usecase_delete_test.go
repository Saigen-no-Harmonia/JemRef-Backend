package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase/dto"

	"github.com/stretchr/testify/assert"
)

// -------------Delete(context.Context, dto.DeleteUserInput)のテスト-------------

// userDeleteCase Deleteのテストケース
type userDeleteCase struct {
	name string
	// firebaseUid 入力のFirebase UID。空文字の場合は登録済みユーザと同じ値を使う
	firebaseUid string
	// deletedUser 取得したユーザが退会済み(DeletedAtあり)である
	deletedUser bool
	// selectErr / deleteErr / firebaseErr 各repositoryが返すエラー(nilは成功)
	selectErr                   error
	deleteErr                   error
	firebaseErr                 error
	expectedSelectCalls         int
	expectedDeleteCalls         int
	expectedFirebaseDeleteCalls int
	expectedErrorBody           error
}

// runUserDeleteCase Deleteを実行し、repositoryの呼び出しと結果を検証する
func runUserDeleteCase(t *testing.T, tt userDeleteCase) {
	t.Helper()

	registered := &user.User{
		InternalUserId: userTestUid,
		FirebaseUserId: userTestFirebaseUid,
	}
	if tt.deletedUser {
		registered.DeletedAt = testutil.Ptr(time.Now())
	}
	userRepo := newSelectUserRepoMock(registered, tt.selectErr)
	userRepo.DeleteFunc = func(ctx context.Context, uid int64) error {
		return tt.deleteErr
	}
	firebaseRepo := &mock.MockFirebaseRepository{
		DeleteUserFunc: func(ctx context.Context, uid string) error {
			return tt.firebaseErr
		},
	}

	input := dto.DeleteUserInput{
		InternalUserid: userTestUid,
		FirebaseUserId: userTestFirebaseUid,
	}
	if tt.firebaseUid != "" {
		input.FirebaseUserId = tt.firebaseUid
	}

	u := NewUserUsecaseImpl(userRepo, nil, firebaseRepo, nil)
	err := u.Delete(context.Background(), input)

	// 呼び出し回数
	assert.Equal(t, tt.expectedSelectCalls, userRepo.SelectByInternalUidCalled)
	assert.Equal(t, tt.expectedDeleteCalls, userRepo.DeleteCalled)
	assert.Equal(t, tt.expectedSelectCalls+tt.expectedDeleteCalls, userRepo.Called)
	assert.Equal(t, tt.expectedFirebaseDeleteCalls, firebaseRepo.DeleteUserCalled)

	if tt.expectedErrorBody != nil {
		assert.Error(t, err)
		assert.ErrorIs(t, err, tt.expectedErrorBody)
		return
	}

	assert.NoError(t, err)
	// 退会済みのユーザはDB側の削除を行わない
	if !tt.deletedUser {
		assert.Equal(t, input.InternalUserid, userRepo.LastDeleteUserId)
	}
	assert.Equal(t, input.FirebaseUserId, firebaseRepo.DeletedFirebaseUid)
}

func TestUserUsecaseImpl_DeleteUser_正常(t *testing.T) {

	tests := []userDeleteCase{
		{
			name:                        "正常_退会",
			expectedSelectCalls:         1,
			expectedDeleteCalls:         1,
			expectedFirebaseDeleteCalls: 1,
		},
		{
			// DB側はすでに論理削除済みなので、Firebaseユーザの削除だけを行う
			name:                        "正常_論理削除済みのユーザはFirebaseユーザのみ削除",
			deletedUser:                 true,
			expectedSelectCalls:         1,
			expectedDeleteCalls:         0,
			expectedFirebaseDeleteCalls: 1,
		},
		{
			// DB側の退会は完了しているため、Firebaseユーザの削除に失敗しても成功として返す
			name:                        "正常_Firebaseユーザ削除に失敗しても成功を返す",
			firebaseErr:                 repository.ErrTestUnexpected,
			expectedSelectCalls:         1,
			expectedDeleteCalls:         1,
			expectedFirebaseDeleteCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runUserDeleteCase(t, tt)
		})
	}
}

func TestUserUsecaseImpl_DeleteUser_異常(t *testing.T) {

	tests := []userDeleteCase{
		{
			name:                "異常_該当ユーザなし",
			selectErr:           sql.ErrNoRows,
			expectedSelectCalls: 1,
			expectedErrorBody:   ErrUserNotFound,
		},
		{
			name:                "異常_ユーザSELECT失敗",
			selectErr:           repository.ErrTestUnexpected,
			expectedSelectCalls: 1,
			expectedErrorBody:   repository.ErrTestUnexpected,
		},
		{
			// 入力のFirebase UIDが登録済みユーザと一致しない
			name:                "異常_パラメータ不正",
			firebaseUid:         "_invalid_firebase_uid_",
			expectedSelectCalls: 1,
			expectedErrorBody:   ErrUserDataInconsistent,
		},
		{
			name:                "異常_DBデータ削除失敗",
			deleteErr:           errors.New("error delete failed"),
			expectedSelectCalls: 1,
			expectedDeleteCalls: 1,
			expectedErrorBody:   ErrUserDeleteFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runUserDeleteCase(t, tt)
		})
	}
}
