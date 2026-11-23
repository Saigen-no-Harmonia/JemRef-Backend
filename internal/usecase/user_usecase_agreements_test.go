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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------GetUserAgreements(context.Context, int64)のテスト-------------

// newUserAgreementsGetMocks 同意済みのバージョン(利用規約1.0・プライバシーポリシー2.0)を持つユーザと、
// 最新版(どちらも2.0)の規約を返すmock一式を返す
//
// selectErr / termsErr / privacyErr を指定すると、その取得が失敗する。
func newUserAgreementsGetMocks(
	selectErr, termsErr, privacyErr error,
) (*mock.MockUserRepository, *mock.MockGeneralRepository) {
	userRepo := newSelectUserRepoMock(&user.User{
		InternalUserId:       userTestUid,
		TermsVersion:         "1.0",
		PrivacyPolicyVersion: "2.0",
	}, selectErr)

	terms := policyResult{policy: &policy.Policy{
		Id:      policy.PolicyIdTermsOfService,
		Version: "2.0",
		Name:    "_terms_label_",
	}}
	if termsErr != nil {
		terms = policyResult{err: termsErr}
	}
	privacy := policyResult{policy: &policy.Policy{
		Id:      policy.PolicyIdPrivacyPolicy,
		Version: "2.0",
		Name:    "_privacy_policy_label_",
	}}
	if privacyErr != nil {
		privacy = policyResult{err: privacyErr}
	}
	generalRepo := newGeneralRepoMock(map[string]policyResult{
		policy.PolicyIdTermsOfService: terms,
		policy.PolicyIdPrivacyPolicy:  privacy,
	})

	return userRepo, generalRepo
}

// runUserGetAgreements GetUserAgreementsを実行し、呼び出し回数を検証する
func runUserGetAgreements(
	t *testing.T,
	userRepo *mock.MockUserRepository,
	generalRepo *mock.MockGeneralRepository,
	expectedGeneralRepoCalls int,
) (*dto.GetUserAgreementsOutput, error) {
	t.Helper()

	u := NewUserUsecaseImpl(userRepo, generalRepo, nil, nil)
	output, err := u.GetUserAgreements(context.Background(), userTestUid)

	assert.Equal(t, 1, userRepo.SelectByInternalUidCalled)
	assert.Equal(t, expectedGeneralRepoCalls, generalRepo.SelectLatestByIdFuncCalled)
	assert.Equal(t, 0, userRepo.CreateCalled)
	assert.Equal(t, 0, userRepo.DeleteCalled)
	assert.Equal(t, 0, userRepo.SelectByFirebaseUidCalled)
	assert.Equal(t, 0, generalRepo.SelectByPrimaryKeyCalled)

	return output, err
}

func TestUserUsecaseImpl_GetUserAgreements_正常(t *testing.T) {
	t.Run("正常_同意状況を取得", func(t *testing.T) {
		userRepo, generalRepo := newUserAgreementsGetMocks(nil, nil, nil)

		output, err := runUserGetAgreements(t, userRepo, generalRepo, 2)

		require.NoError(t, err)
		assert.Equal(t, userTestUid, userRepo.LastSelectInternalUid)
		assert.Equal(t, []string{policy.PolicyIdTermsOfService, policy.PolicyIdPrivacyPolicy}, generalRepo.SelectedIds)
		assert.Equal(t, &dto.GetUserAgreementsOutput{
			Agreements: []dto.GetUserAgreement{
				{
					// 同意済み1.0 < 最新2.0 のため再同意が必要
					PolicyType:    policy.PolicyTypeTermsOfService,
					Label:         "_terms_label_",
					LatestVersion: "2.0",
					AgreedVersion: "1.0",
					Status:        user.PolicyAgreementStatusUpdateRequired,
				},
				{
					PolicyType:    policy.PolicyTypePrivacyPolicy,
					Label:         "_privacy_policy_label_",
					LatestVersion: "2.0",
					AgreedVersion: "2.0",
					Status:        user.PolicyAgreementStatusAgreed,
				},
			},
		}, output)
	})
}

func TestUserUsecaseImpl_GetUserAgreements_異常(t *testing.T) {

	tests := []struct {
		name string
		// selectErr / termsErr / privacyErr ユーザ・利用規約・プライバシーポリシーの取得が返すエラー
		selectErr                error
		termsErr                 error
		privacyErr               error
		expectedGeneralRepoCalls int
		expectedErrorBody        error
	}{
		{
			name:      "異常_ユーザマスタ取得失敗",
			selectErr: repository.ErrNotFound,
			// ユーザの取得で中断する
			expectedGeneralRepoCalls: 0,
			expectedErrorBody:        ErrUserNotFound,
		},
		{
			name:                     "異常_ユーザ情報取得時に予期せぬエラー",
			selectErr:                repository.ErrTestUnexpected,
			expectedGeneralRepoCalls: 0,
			expectedErrorBody:        repository.ErrTestUnexpected,
		},
		{
			name:     "異常_ユーザ規約取得失敗",
			termsErr: repository.ErrNotFound,
			// 利用規約の取得で中断する
			expectedGeneralRepoCalls: 1,
			expectedErrorBody:        ErrPolicyNotFound,
		},
		{
			name:                     "異常_ユーザ規約取得時に予期せぬエラー",
			termsErr:                 repository.ErrTestUnexpected,
			expectedGeneralRepoCalls: 1,
			expectedErrorBody:        repository.ErrTestUnexpected,
		},
		{
			name:                     "異常_プラポリ取得失敗",
			privacyErr:               repository.ErrNotFound,
			expectedGeneralRepoCalls: 2,
			expectedErrorBody:        ErrPolicyNotFound,
		},
		{
			name:                     "異常_プラポリ取得時に予期せぬエラー",
			privacyErr:               repository.ErrTestUnexpected,
			expectedGeneralRepoCalls: 2,
			expectedErrorBody:        repository.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, generalRepo := newUserAgreementsGetMocks(tt.selectErr, tt.termsErr, tt.privacyErr)

			output, err := runUserGetAgreements(t, userRepo, generalRepo, tt.expectedGeneralRepoCalls)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErrorBody)
			assert.Nil(t, output)
		})
	}
}

// -------------UpdateUserAgreements(context.Context, dto.UpdateUserAgreementsInput)のテスト-------------

// userAgreementsBoth 利用規約1.0・プライバシーポリシー2.0の両方に同意する入力
var userAgreementsBoth = []dto.UpdateUserAgreement{
	{PolicyType: policy.PolicyTypeTermsOfService, AgreedVersion: "1.0"},
	{PolicyType: policy.PolicyTypePrivacyPolicy, AgreedVersion: "2.0"},
}

// userAgreementsTermsOnly 利用規約1.0だけに同意する入力
var userAgreementsTermsOnly = []dto.UpdateUserAgreement{
	{PolicyType: policy.PolicyTypeTermsOfService, AgreedVersion: "1.0"},
}

// userAgreementsUpdateCase UpdateUserAgreementsのテストケース
type userAgreementsUpdateCase struct {
	name string
	// agreements 入力の同意内容。nilの場合は userAgreementsBoth を使う
	agreements []dto.UpdateUserAgreement
	// policies 規約IDごとの取得結果。nilの場合は入力どおりのバージョンの規約が見つかる
	policies map[string]policyResult
	// selectErr / updateErr ユーザの取得・更新が返すエラー(nilは成功)
	selectErr                error
	updateErr                error
	expectedGeneralRepoCalls int
	expectedUpdateCalls      int
	expectedErrorBody        error
}

// runUserAgreementsUpdateCase UpdateUserAgreementsを実行し、repositoryの呼び出しと結果を検証する
func runUserAgreementsUpdateCase(t *testing.T, tt userAgreementsUpdateCase) {
	t.Helper()

	agreedDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)
	userRepo := newSelectUserRepoMock(&user.User{
		InternalUserId:        userTestUid,
		TermsAgreedDt:         &agreedDate,
		TermsVersion:          "_tmp1_",
		PrivacyPolicyAgreedDt: &agreedDate,
		PrivacyPolicyVersion:  "_tmp2_",
	}, tt.selectErr)
	userRepo.UpdateUserAgreementFunc = func(ctx context.Context, u *user.User) (int64, error) {
		if tt.updateErr != nil {
			return 0, tt.updateErr
		}
		return 1, nil
	}

	policies := tt.policies
	if policies == nil {
		policies = map[string]policyResult{
			policy.PolicyIdTermsOfService: {policy: &policy.Policy{Id: policy.PolicyIdTermsOfService, Version: "1.0"}},
			policy.PolicyIdPrivacyPolicy:  {policy: &policy.Policy{Id: policy.PolicyIdPrivacyPolicy, Version: "2.0"}},
		}
	}
	generalRepo := newGeneralRepoMock(policies)

	agreements := tt.agreements
	if agreements == nil {
		agreements = userAgreementsBoth
	}
	input := dto.UpdateUserAgreementsInput{InternalUid: userTestUid, Agreements: agreements}

	before := time.Now()
	u := NewUserUsecaseImpl(userRepo, generalRepo, nil, nil)
	err := u.UpdateUserAgreements(context.Background(), input)
	after := time.Now()

	// 呼び出し回数
	assert.Equal(t, tt.expectedGeneralRepoCalls, generalRepo.SelectByPrimaryKeyCalled)
	assert.Equal(t, tt.expectedUpdateCalls, userRepo.UpdateUserAgreementCalled)
	assert.Equal(t, 1, userRepo.SelectByInternalUidCalled)
	assert.Equal(t, 0, generalRepo.SelectLatestByIdFuncCalled)
	assert.Equal(t, 0, userRepo.CreateCalled)
	assert.Equal(t, 0, userRepo.DeleteCalled)
	assert.Equal(t, 0, userRepo.SelectByFirebaseUidCalled)

	if tt.expectedErrorBody != nil {
		assert.Error(t, err)
		assert.ErrorIs(t, err, tt.expectedErrorBody)
		return
	}

	assert.NoError(t, err)
	assert.Equal(t, []mock.PolicyPrimaryKeyCall{
		{Id: policy.PolicyIdTermsOfService, Version: "1.0"},
		{Id: policy.PolicyIdPrivacyPolicy, Version: "2.0"},
	}, generalRepo.SelectedPrimaryKeys)
	assert.Equal(t, userTestUid, userRepo.LastSelectInternalUid)

	// 同意したバージョンと同意日時(UTC)で更新される
	updated := userRepo.LastUpdatedUser
	assert.Equal(t, userTestUid, updated.InternalUserId)
	assert.Equal(t, "1.0", updated.TermsVersion)
	assert.Equal(t, "2.0", updated.PrivacyPolicyVersion)
	assert.WithinRange(t, *updated.TermsAgreedDt, before, after)
	assert.WithinRange(t, *updated.PrivacyPolicyAgreedDt, before, after)
	assert.Equal(t, time.UTC, updated.TermsAgreedDt.Location())
}

func TestUserUsecaseImpl_UpdateUserAgreements_正常(t *testing.T) {
	t.Run("正常_同意状況を更新", func(t *testing.T) {
		runUserAgreementsUpdateCase(t, userAgreementsUpdateCase{
			expectedGeneralRepoCalls: 2,
			expectedUpdateCalls:      1,
		})
	})
}

func TestUserUsecaseImpl_UpdateUserAgreements_異常(t *testing.T) {

	tests := []userAgreementsUpdateCase{
		{
			name:      "異常_ユーザが存在しない",
			selectErr: repository.ErrNotFound,
			// ユーザの取得で中断する
			expectedGeneralRepoCalls: 0,
			expectedErrorBody:        ErrUserNotFound,
		},
		{
			name:                     "異常_ユーザ取得処理で予期せぬエラー",
			selectErr:                repository.ErrTestUnexpected,
			expectedGeneralRepoCalls: 0,
			expectedErrorBody:        repository.ErrTestUnexpected,
		},
		{
			// 2件目の規約タイプが不正なため、1件目の確認だけが行われる
			name: "異常_リクエストされた規約タイプが不正",
			agreements: []dto.UpdateUserAgreement{
				{PolicyType: policy.PolicyTypeTermsOfService, AgreedVersion: "1.0"},
				{PolicyType: "不正な規約", AgreedVersion: "2.0"},
			},
			expectedGeneralRepoCalls: 1,
			expectedErrorBody:        ErrInvalidPolicyType,
		},
		{
			name:       "異常_規約が存在しない（バージョン指定誤り）",
			agreements: userAgreementsTermsOnly,
			policies: map[string]policyResult{
				policy.PolicyIdTermsOfService: {err: repository.ErrNotFound},
			},
			expectedGeneralRepoCalls: 1,
			expectedErrorBody:        ErrInvalidPolicyVersion,
		},
		{
			name:       "異常_規約取得時に予期せぬエラー",
			agreements: userAgreementsTermsOnly,
			policies: map[string]policyResult{
				policy.PolicyIdTermsOfService: {err: repository.ErrTestUnexpected},
			},
			expectedGeneralRepoCalls: 1,
			expectedErrorBody:        repository.ErrTestUnexpected,
		},
		{
			// 要求と異なる規約IDが返ってきた
			name:       "異常_DBが不正な規約情報を返却",
			agreements: userAgreementsTermsOnly,
			policies: map[string]policyResult{
				policy.PolicyIdTermsOfService: {policy: &policy.Policy{Id: "P999", Version: "1.0"}},
			},
			expectedGeneralRepoCalls: 1,
			expectedErrorBody:        ErrUnexpectedPolicy,
		},
		{
			name:                     "異常_ユーザマスタ更新に失敗",
			updateErr:                repository.ErrTestUnexpected,
			expectedGeneralRepoCalls: 2,
			expectedUpdateCalls:      1,
			expectedErrorBody:        repository.ErrTestUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runUserAgreementsUpdateCase(t, tt)
		})
	}
}
