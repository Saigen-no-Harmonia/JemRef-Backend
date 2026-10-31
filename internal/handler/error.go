package handler

import (
	"errors"
)

var (
	ErrTestUnexpected          = errors.New("handler test unexpected")
	ErrInternalUidNotFound     = errors.New("contextからinternal uidが取得できませんでした。")
	ErrFirebaseUidNotFound     = errors.New("contextからfirebase uidが取得できませんでした。")
	ErrPublicUidNotFound       = errors.New("contextから公開用UIDが取得できませんでした。")
	ErrEmailNotFound           = errors.New("contextからメールアドレスが取得できませんでした。")
	ErrPolicyRequired          = errors.New("リクエストに規約情報がありませんでした。")
	ErrAgreedVersionRequired   = errors.New("規約同意バージョンは必須です。")
	ErrInvalidRequestBody      = errors.New("リクエストボディの形式が不正です。")
	ErrPolicyTypeInvalid       = errors.New("規約タイプ指定が不正です。")
	ErrMainTitleRequired       = errors.New("主題は必須項目です。")
	ErrRecordTypeRequired      = errors.New("書誌タイプ指定は必須です。")
	ErrInvalidRecordType       = errors.New("書誌タイプ指定が不正です。")
	ErrContainerNotAllowed     = errors.New("この書誌タイプは親レコードを持つことができません。")
	ErrInvalidPublicationDate  = errors.New("出版日付の指定が不正です。")
	ErrInvalidReadStatus       = errors.New("読了ステータスが不正です。")
	ErrInvalidContributorRole  = errors.New("貢献者役割が不正です。")
	ErrContributorNameRequired = errors.New("貢献者の氏名は必須です。")
	ErrContributorRoleRequired = errors.New("貢献者の役割は必須です。")
	ErrMemoContentRequired     = errors.New("メモの内容は必須です。")
	ErrUrlValueRequired        = errors.New("URLの値は必須です。")
)
