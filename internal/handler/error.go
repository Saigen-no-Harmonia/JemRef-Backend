package handler

import (
	"errors"

	"jemref_go/internal/apperr"
)

// ErrTestUnexpected 分岐でカバーしていない予期しないエラーのテスト用
// apperr.Errorにしないことで、変換されず500に落ちることを検証できる。
var ErrTestUnexpected = errors.New("handler test unexpected")

// contextからの値取得失敗はいずれもサーバ側の不具合なので、
// レスポンスは共通の500に寄せ、どの値が取れなかったかはログに残す。
var (
	ErrInternalUidNotFound = apperr.Internal.New("contextからinternal uidが取得できませんでした。")
	ErrFirebaseUidNotFound = apperr.Internal.New("contextからfirebase uidが取得できませんでした。")
	ErrPublicUidNotFound   = apperr.Internal.New("contextから公開用UIDが取得できませんでした。")
	ErrEmailNotFound       = apperr.Internal.New("contextからメールアドレスが取得できませんでした。")
)

// リクエストの不備。フロント側に見せる文言がそのまま原因の説明になるものはErr()を使う。
var (
	ErrInvalidRequestBody    = apperr.BadRequest.Err()
	ErrPolicyRequired        = apperr.PolicyRequired.Err()
	ErrInvalidPolicyType     = apperr.InvalidPolicyType.Err()
	ErrAgreedVersionRequired = apperr.AgreedVersionRequired.Err()

	ErrMainTitleRequired      = apperr.MainTitleRequired.Err()
	ErrRecordTypeRequired     = apperr.RecordTypeRequired.Err()
	ErrInvalidRecordType      = apperr.InvalidRecordType.Err()
	ErrContainerNotAllowed    = apperr.ContainerNotAllowed.Err()
	ErrInvalidPublicationDate = apperr.InvalidPublicationDate.Err()
	ErrInvalidReadStatus      = apperr.InvalidReadStatus.Err()
	ErrRecordIdRequired       = apperr.RecordIdRequired.Err()

	ErrInvalidContributorRole  = apperr.InvalidContributorRole.Err()
	ErrContributorNameRequired = apperr.ContributorNameRequired.Err()
	ErrContributorRoleRequired = apperr.ContributorRoleRequired.Err()
	ErrMemoContentRequired     = apperr.MemoContentRequired.Err()
	ErrUrlValueRequired        = apperr.UrlValueRequired.Err()
)
