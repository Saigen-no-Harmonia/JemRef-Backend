package usecase

// ユーザUsecase実装

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"jemref_go/internal/domain/id"
	"jemref_go/internal/domain/policy"
	"jemref_go/internal/domain/user"
	"jemref_go/internal/repository"
	"jemref_go/internal/usecase/dto"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

type UserUsecaseImpl struct {
	userRepo     repository.UserRepository
	firebaseRepo repository.FirebaseRepository
	idGen        id.Generator
	generalRepo  repository.GeneralRepository
	// txManager repository.TxManager
}

func NewUserUsecase(ur repository.UserRepository, gr repository.GeneralRepository, fr repository.FirebaseRepository, g id.Generator) UserUsecase {
	return &UserUsecaseImpl{
		userRepo:     ur,
		generalRepo:  gr,
		firebaseRepo: fr,
		idGen:        g,
	}
}

// Create ユーザを作成し、公開用UIDを返却する
func (uu *UserUsecaseImpl) Create(ctx context.Context, cu dto.CreateUserInput) (*dto.CreateUserOutput, error) {

	publicUserId := uu.idGen.Generate()

	// 規約マスタチェック
	_, err := uu.generalRepo.SelectPolicyByPrimaryKey(
		ctx,
		policy.PolicyIdTermsOfService,
		cu.TermsAgreedVersion,
	)
	if err != nil {
		return nil, fmt.Errorf("ユーザ規約マスタ情報が取得できませんでした。:%w", ErrPolicyNotFound)
	}

	_, err = uu.generalRepo.SelectPolicyByPrimaryKey(
		ctx,
		policy.PolicyIdPrivacyPolicy,
		cu.PrivacyPolicyAgreedVersion,
	)
	if err != nil {
		return nil, fmt.Errorf("プライバシーポリシーマスタ情報が取得できませんでした。:%w", ErrPolicyNotFound)
	}

	sysDate := time.Now()

	user := &user.User{
		PublicUserId:          publicUserId,
		FirebaseUserId:        cu.FirebaseUserId,
		Email:                 cu.Email,
		TermsAgreedDt:         &sysDate,
		TermsVersion:          cu.TermsAgreedVersion,
		PrivacyPolicyAgreedDt: &sysDate,
		PrivacyPolicyVersion:  cu.PrivacyPolicyAgreedVersion,
	}

	err = uu.userRepo.Create(ctx, user)

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return nil, fmt.Errorf(
			"すでにユーザが存在します。firebaseuid=%s, email=%s :%w",
			user.FirebaseUserId,
			user.Email,
			ErrUserAlreadyExists,
		)
	}

	if err != nil {
		return nil, fmt.Errorf(
			"create user by firebaseuid=%s, email=%s :%w",
			user.FirebaseUserId,
			user.Email,
			err,
		)
	}

	return &dto.CreateUserOutput{
		PublicUserId: publicUserId,
	}, nil

}

// Delete DBのユーザと Firebaseユーザを削除する
func (uu *UserUsecaseImpl) Delete(ctx context.Context, du dto.DeleteUserInput) error {

	internalUid := du.InternalUserid
	firebaseUid := du.FirebaseUserId

	user, err := uu.userRepo.SelectByInternalUid(ctx, internalUid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf(
				"削除対象ユーザが存在しません。internal uid=%d :%w",
				internalUid,
				ErrUserNotFound,
			)
		}
		return fmt.Errorf("select user by internal uid=%d :%w", internalUid, err)
	}

	// マスタ整合性チェック
	if user.FirebaseUserId != du.FirebaseUserId {
		return fmt.Errorf(
			"ユーザデータに不整合があります。internal uid=%d, firebase uid=%s :%w",
			internalUid,
			firebaseUid,
			ErrInvalidUser,
		)
	}

	if user.DeletedAt == nil {
		if err := uu.userRepo.Delete(ctx, internalUid); err != nil {
			return fmt.Errorf(
				"DBのユーザデータ削除に失敗しました。internal uid=%d :%w",
				internalUid,
				ErrUserDeleteFailed,
			)
		}
	} else {
		// すでに論理削除済みならDB更新のみスキップ
		log.Printf("すでに退会処理済みのため、FirebaseUser削除のみ実施します。firebase uid: %s", firebaseUid)
	}

	// firebaseユーザ削除
	if err = uu.firebaseRepo.DeleteUser(ctx, firebaseUid); err != nil {
		log.Printf("firebaseユーザの削除に失敗しました。internal uid: %d, firebase uid: %s", internalUid, firebaseUid)
		// 失敗時は、共通認証で自動リカバリ（FirebaseUser削除）する設計のため、ログ出力のみで終了
		return nil
	}

	return nil
}

// GetUserAgreements 指定されたユーザについて、規約同意状況を判定して返却する
func (uu *UserUsecaseImpl) GetUserAgreements(ctx context.Context, uid int64) (*dto.GetUserAgreementsOutput, error) {

	u, err := uu.userRepo.SelectByInternalUid(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("select user by internal uid=%d :%w", uid, ErrUserNotFound)
		}
		return nil, fmt.Errorf("select user by internal uid=%d :%w", uid, err)
	}

	// Ph0では、P001とP002だけがある想定なので、それぞれ取得・判定
	terms, err := uu.generalRepo.SelectLatestPolicyById(ctx, policy.PolicyIdTermsOfService)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("ユーザ利用規約情報が取得できませんでした。:%w", ErrPolicyNotFound)
		}
		return nil, fmt.Errorf("select terms master:%w", err)
	}

	privacyPolicy, err := uu.generalRepo.SelectLatestPolicyById(ctx, policy.PolicyIdPrivacyPolicy)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("プライバシーポリシー情報が取得できませんでした。:%w", ErrPolicyNotFound)
		}
		return nil, fmt.Errorf("select privacy policy master :%w", err)
	}

	return createGetUserAgreementsOutput(u, terms, privacyPolicy)
}

// UpdateUserAgreements 引数で指定したユーザ情報に基づき、ユーザ同意状況を更新する
func (uu *UserUsecaseImpl) UpdateUserAgreements(ctx context.Context, ua dto.UpdateUserAgreementsInput) error {

	// 現時点の同意状況をベースにして更新処理するため、１度SELECTする
	u, err := uu.userRepo.SelectByInternalUid(ctx, ua.InternalUid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("ユーザ情報が存在しません。 internal uid=%d :%w", ua.InternalUid, ErrUserNotFound)
		}
		return fmt.Errorf("select user by internal uid=%d :%w", ua.InternalUid, err)
	}

	now := time.Now()

	for _, a := range ua.Agreements {

		if !a.PolicyType.IsValid() {
			return fmt.Errorf(
				"規約タイプが不正です。policyType=%s :%w",
				a.PolicyType,
				ErrInvalidPolicyType,
			)
		}

		p, err := uu.generalRepo.SelectPolicyByPrimaryKey(
			ctx,
			a.PolicyType.GetId(),
			a.AgreedVersion,
		)
		if err != nil {
			// ここでのnot foundはバージョン指定不正のみ想定される
			if errors.Is(err, repository.ErrNotFound) {
				return fmt.Errorf(
					"規約バージョン指定が不正です。policy id=%s, agreed version=%s :%w",
					a.PolicyType.GetId(),
					a.AgreedVersion,
					ErrInvalidPolicyVersion,
				)
			}
			return fmt.Errorf(
				"select policy by policy id=%s, agreed version=%s :%w",
				a.PolicyType.GetId(),
				a.AgreedVersion,
				err,
			)
		}

		switch p.Id {
		case policy.PolicyIdTermsOfService:
			u.TermsAgreedDt = &now
			u.TermsVersion = p.Version

		case policy.PolicyIdPrivacyPolicy:
			u.PrivacyPolicyAgreedDt = &now
			u.PrivacyPolicyVersion = p.Version

		default:
			return fmt.Errorf("invalid policy id %s %w", p.Id, ErrUnexpectedPolicy)
		}
	}

	rows, err := uu.userRepo.UpdateUserAgreement(ctx, u)

	if err != nil {
		return fmt.Errorf("update user agreements by internal uid=%d :%w", u.InternalUserId, err)
	}

	// 更新件数が0件でもエラーにはしない（更新対象ユーザがDBに存在することは共通認証処理で確認ずみ）
	_ = rows

	return nil
}

// Login ユーザログイン処理 Ph0では公開用UIDを返却するだけ
func (uu *UserUsecaseImpl) Login(ctx context.Context, ui dto.UserLoginInput) (*dto.UserLoginOutput, error) {
	uid := ui.InternalUserId

	u, err := uu.userRepo.SelectByInternalUid(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf(
				"ユーザ情報が存在しません。internal uid =%d :%w",
				uid,
				ErrUserNotFound,
			)
		}
		return nil, fmt.Errorf(
			"select user by internal uid =%d :%w",
			uid,
			err,
		)
	}

	return &dto.UserLoginOutput{PublicUserId: u.PublicUserId}, err
}

// createGetUserAgreementsOutput ユーザの規約同意状況を判定し返却する。
// Ph0ではユーザ規約とプラポリだけが存在する想定で簡易実装している。
func createGetUserAgreementsOutput(u *user.User, terms, privacyPolicy *policy.Policy) (*dto.GetUserAgreementsOutput, error) {

	var res dto.GetUserAgreementsOutput

	termsStatus := dto.GetUserAgreement{
		PolicyType:    policy.PolicyTypeTermsOfService,
		Label:         terms.Name,
		LatestVersion: terms.Version,
		AgreedVersion: u.TermsVersion,
		Status:        user.ChkAgreementStat(u.TermsVersion, terms.Version),
	}

	privacyPolicyStatis := dto.GetUserAgreement{
		PolicyType:    policy.PolicyTypePrivacyPolicy,
		Label:         privacyPolicy.Name,
		LatestVersion: privacyPolicy.Version,
		AgreedVersion: u.PrivacyPolicyVersion,
		Status:        user.ChkAgreementStat(u.PrivacyPolicyVersion, privacyPolicy.Version),
	}

	res.Agreements = append(res.Agreements, termsStatus)
	res.Agreements = append(res.Agreements, privacyPolicyStatis)

	return &res, nil
}
