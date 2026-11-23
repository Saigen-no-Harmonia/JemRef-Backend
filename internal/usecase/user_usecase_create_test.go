package usecase

import (
	"context"
	"testing"
	"time"

	"jemref_go/internal/domain/policy"
	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"
	"jemref_go/internal/testutil/mock"
	"jemref_go/internal/usecase/dto"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------Create(context.Context, dto.CreateUserInput)のテスト-------------

// userCreateInput 会員登録の入力(全ケース共通)
var userCreateInput = dto.CreateUserInput{
	FirebaseUserId:             "firebase_uid",
	Email:                      "sample@example.com",
	TermsAgreedVersion:         "0.01",
	PrivacyPolicyAgreedVersion: "0.02",
}

// userCreateGeneratedId idGenが返す公開用ユーザID
const userCreateGeneratedId = "_sample_id_"

// userCreateMocks Createに渡すmock一式
type userCreateMocks struct {
	userRepo    *mock.MockUserRepository
	generalRepo *mock.MockGeneralRepository
	idGen       *mock.MockIdGen
}

// newUserCreateMocks 規約の存在確認とユーザ登録が、指定したエラー(nilなら成功)を返すmock一式を返す
func newUserCreateMocks(termsErr, privacyErr, createErr error) userCreateMocks {
	return userCreateMocks{
		userRepo: &mock.MockUserRepository{
			CreateFunc: func(ctx context.Context, u *user.User) error {
				return createErr
			},
		},
		generalRepo: newGeneralRepoMock(map[string]policyResult{
			policy.PolicyIdTermsOfService: {err: termsErr},
			policy.PolicyIdPrivacyPolicy:  {err: privacyErr},
		}),
		idGen: &mock.MockIdGen{
			GenerateFunc: func() string {
				return userCreateGeneratedId
			},
		},
	}
}

// runUserCreate Createを実行し、呼び出し回数(ID生成は常に1回、Create以外のユーザ操作は0回)を検証する
func runUserCreate(
	t *testing.T,
	m userCreateMocks,
	expectedGeneralRepoCalls int,
	expectedUserRepoCalls int,
) (*dto.CreateUserOutput, error) {
	t.Helper()

	u := NewUserUsecaseImpl(m.userRepo, m.generalRepo, nil, m.idGen)
	output, err := u.Create(context.Background(), userCreateInput)

	assert.Equal(t, expectedGeneralRepoCalls, m.generalRepo.SelectByPrimaryKeyCalled)
	assert.Equal(t, expectedUserRepoCalls, m.userRepo.CreateCalled)
	assert.Equal(t, 1, m.idGen.GenerateCalls)
	assert.Equal(t, 0, m.userRepo.DeleteCalled)
	assert.Equal(t, 0, m.userRepo.SelectByFirebaseUidCalled)
	assert.Equal(t, 0, m.userRepo.SelectByInternalUidCalled)

	return output, err
}

func TestUserUsecaseImpl_CreateUser_正常(t *testing.T) {
	t.Run("正常_会員登録", func(t *testing.T) {
		m := newUserCreateMocks(nil, nil, nil)
		timestampBefore := time.Now()

		output, err := runUserCreate(t, m, 2, 1)
		timestampAfter := time.Now()

		require.NoError(t, err)
		assert.Equal(t, userCreateGeneratedId, output.PublicUserId)

		// repositoryへ渡したユーザ
		inserted := m.userRepo.LastInsertUser
		assert.Equal(t, userCreateGeneratedId, inserted.PublicUserId)
		assert.Equal(t, userCreateInput.FirebaseUserId, inserted.FirebaseUserId)
		assert.Equal(t, userCreateInput.Email, inserted.Email)
		assert.Equal(t, "0.01", inserted.TermsVersion)
		assert.Equal(t, "0.02", inserted.PrivacyPolicyVersion)
		assert.WithinRange(t, *inserted.TermsAgreedDt, timestampBefore, timestampAfter)
		assert.WithinRange(t, *inserted.PrivacyPolicyAgreedDt, timestampBefore, timestampAfter)
		assert.Equal(t, time.UTC, inserted.TermsAgreedDt.Location())
	})
}

func TestUserUsecaseImpl_CreateUser_異常(t *testing.T) {

	tests := []struct {
		name string
		// termsErr / privacyErr / createErr 規約の存在確認・ユーザ登録が返すエラー(nilは成功)
		termsErr                 error
		privacyErr               error
		createErr                error
		expectedGeneralRepoCalls int
		expectedUserRepoCalls    int
		expectedErrorBody        error
	}{
		{
			name:     "異常_ユーザ規約チェック失敗",
			termsErr: ErrPolicyNotFound,
			// 利用規約の確認で中断する
			expectedGeneralRepoCalls: 1,
			expectedUserRepoCalls:    0,
			expectedErrorBody:        ErrPolicyNotFound,
		},
		{
			name:                     "異常_プラポリ規約チェック失敗",
			privacyErr:               ErrPolicyNotFound,
			expectedGeneralRepoCalls: 2,
			expectedUserRepoCalls:    0,
			expectedErrorBody:        ErrPolicyNotFound,
		},
		{
			// 一意制約違反(1062)は既存ユーザのエラーに変換される
			name:                     "異常_ユーザ重複エラー",
			createErr:                &mysql.MySQLError{Number: 1062},
			expectedGeneralRepoCalls: 2,
			expectedUserRepoCalls:    1,
			expectedErrorBody:        ErrUserAlreadyExists,
		},
		{
			name:                     "異常_その他のユーザRepoエラー",
			createErr:                repository.ErrTestUnexpected,
			expectedGeneralRepoCalls: 2,
			expectedUserRepoCalls:    1,
			expectedErrorBody:        repository.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newUserCreateMocks(tt.termsErr, tt.privacyErr, tt.createErr)

			output, err := runUserCreate(t, m, tt.expectedGeneralRepoCalls, tt.expectedUserRepoCalls)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErrorBody)
			assert.Nil(t, output)
		})
	}
}
