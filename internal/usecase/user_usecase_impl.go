package usecase

// ユーザUsecase実装

import (
	"context"
	"database/sql"
	"errors"
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
	log.Printf("ユーザ情報登録usecase 処理開始: firebase uid: %s", cu.FirebaseUserId)

	publicUserId := uu.idGen.Generate()

	// 規約マスタチェック
	_, err := uu.generalRepo.SelectPolicyByPrimaryKey(
		ctx,
		policy.PolicyIdTermsOfService,
		cu.TermsAgreedVersion,
	)
	if err != nil {
		log.Println("ユーザ規約が取得できませんでした")
		return nil, ErrPolicyNotFound
	}

	_, err = uu.generalRepo.SelectPolicyByPrimaryKey(
		ctx,
		policy.PolicyIdPrivacyPolicy,
		cu.PrivacyPolicyAgreedVersion,
	)
	if err != nil {
		log.Println("プライバシーポリシーが取得できませんでした")
		return nil, ErrPolicyNotFound
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
		return nil, ErrUserAlreadyExists
	}

	if err != nil {
		return nil, err
	}

	log.Printf("ユーザ情報登録usecase 処理完了: firebase uid: %s", cu.FirebaseUserId)
	return &dto.CreateUserOutput{
		PublicUserId: publicUserId,
	}, nil

}

// Delete DBのユーザと Firebaseユーザを削除する
func (uu *UserUsecaseImpl) Delete(ctx context.Context, du dto.DeleteUserInput) error {

	internalUid := du.InternalUserid
	firebaseUid := du.FirebaseUserId
	log.Printf("ユーザ削除usecase 処理開始: internal uid: %d", internalUid)

	user, err := uu.userRepo.SelectByInternalUid(ctx, internalUid)
	// ユーザデータが存在しない場合
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("削除対象ユーザが存在しません。internal uid: %d", internalUid)
			return ErrUserNotFound
		}
		// その他のDBエラー
		log.Printf("database error: %s", err)
		return err
	}

	// マスタ整合性チェック
	if user.FirebaseUserId != du.FirebaseUserId {
		log.Printf("ユーザデータに不整合があります。internal uid: %d, firebase uid: %s", internalUid, firebaseUid)
		return ErrInvalidUser
	}

	if user.DeletedAt == nil {
		if err := uu.userRepo.Delete(ctx, internalUid); err != nil {
			log.Print(err)
			log.Printf("DBのユーザデータ削除に失敗しました。internal uid: %d", internalUid)
			return ErrUserDeleteFailed
		}
		// すでに論理削除済みならDB更新のみスキップ
	} else {
		log.Printf("すでに退会処理済みのため、FirebaseUser削除のみ実施します。firebase uid: %s", firebaseUid)
	}

	// firebaseユーザ削除
	if err = uu.firebaseRepo.DeleteUser(ctx, firebaseUid); err != nil {
		log.Printf("firebaseユーザの削除に失敗しました。internal uid: %d, firebase uid: %s", internalUid, firebaseUid)
		// 失敗時は、共通認証で自動リカバリ（FirebaseUser削除）する設計のため、ログ出力のみで終了
		return nil
	}

	log.Printf("ユーザ削除usecase 処理完了: internal uid: %d", internalUid)
	return nil
}

// GetUserAgreements 指定されたユーザについて、規約同意状況を判定して返却する
func (uu *UserUsecaseImpl) GetUserAgreements(ctx context.Context, uid int64) (*dto.GetUserAgreementsOutput, error) {

	u, err := uu.userRepo.SelectByInternalUid(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Ph0では、P001とP002だけがある想定なので、それぞれ取得・判定
	terms, err := uu.generalRepo.SelectLatestPolicyById(ctx, policy.PolicyIdTermsOfService)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			log.Println("ユーザ利用規約情報が取得できませんでした。")
			return nil, ErrPolicyNotFound
		}
		return nil, err
	}

	privacyPolicy, err := uu.generalRepo.SelectLatestPolicyById(ctx, policy.PolicyIdPrivacyPolicy)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			log.Println("プライバシーポリシー情報が取得できませんでした。")
			return nil, ErrPolicyNotFound
		}
		return nil, err
	}

	return createGetUserAgreementsOutput(u, terms, privacyPolicy)
}

// UpdateUserAgreements 引数で指定したユーザ情報に基づき、ユーザ同意状況を更新する
func (uu *UserUsecaseImpl) UpdateUserAgreements(ctx context.Context, ua dto.UpdateUserAgreementsInput) error {

	// 現時点の同意状況をベースにして更新処理するため、１度SELECTする
	u, err := uu.userRepo.SelectByInternalUid(ctx, ua.InternalUid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	now := time.Now()

	for _, a := range ua.Agreements {

		if !a.PolicyType.IsValid() {
			return ErrInvalidPolicyType
		}

		p, err := uu.generalRepo.SelectPolicyByPrimaryKey(
			ctx,
			a.PolicyType.GetId(),
			a.AgreedVersion,
		)
		if err != nil {
			// ここでのnot foundはバージョン指定不正のみ想定される
			if errors.Is(err, repository.ErrNotFound) {
				return ErrInvalidPolicyVersion
			}
			return err
		}

		switch p.Id {
		case policy.PolicyIdTermsOfService:
			u.TermsAgreedDt = &now
			u.TermsVersion = p.Version

		case policy.PolicyIdPrivacyPolicy:
			u.PrivacyPolicyAgreedDt = &now
			u.PrivacyPolicyVersion = p.Version

		default:
			return ErrUnexpectedPolicy
		}
	}

	rows, err := uu.userRepo.UpdateUserAgreement(ctx, u)

	if err != nil {
		return err
	}

	// 更新件数が0件でもエラーにはしない（更新対象ユーザがDBに存在することは共通認証処理で確認ずみ）
	_ = rows

	return nil
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

// Login ユーザログイン処理 Ph0では公開用UIDを返却するだけ
func (uu *UserUsecaseImpl) Login(ctx context.Context, ui dto.UserLoginInput) (*dto.UserLoginOutput, error) {
	uid := ui.InternalUserId
	log.Printf("ユーザログインusecase 処理開始: internal uid: %d", uid)

	u, err := uu.userRepo.SelectByInternalUid(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	log.Printf("ユーザログインusecase 処理完了: internal uid: %d", uid)
	return &dto.UserLoginOutput{PublicUserId: u.PublicUserId}, err
}
