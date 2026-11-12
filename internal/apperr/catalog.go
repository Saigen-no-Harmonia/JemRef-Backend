package apperr

// エラーコードと公開メッセージの台帳
//
// ここがコード・メッセージの唯一の正。各層はこのDefを参照してエラーを定義する。
// 同じコードを複数の層で使う場合も、メッセージはここ1箇所で管理される。
//
// 新しいエラーを追加する手順:
//  1. ここに def(...) で1行追加する(コードは連番)
//  2. 発生元の層の errors.go で Xxx.Err() または Xxx.New("ログ用の詳細") を定義する
//
// internal/api に変換表はなく、テストは台帳を走査するため、1 と 2 だけで完結する。

// registry def()で宣言された全定義
//
// All()との同期ズレを防ぐため、Defは必ず def() 経由で宣言する
// (直接 Def{...} と書くと台帳の不変条件テストの対象から漏れる)。
var registry []Def

// def 定義を台帳へ登録しつつ返す
func def(kind Kind, code, message string) Def {
	d := Def{Kind: kind, Code: code, Message: message}
	registry = append(registry, d)

	return d
}

// All 台帳の全定義を返す
//
// 呼び出し側での書き換えを防ぐため複製を返す。順序は宣言順。
func All() []Def {
	all := make([]Def, len(registry))
	copy(all, registry)

	return all
}

var (
	// システムエラー
	Internal = def(KindInternal, "F0001", "予期せぬエラーが発生しました。")

	// 認証・ユーザ
	Unauthorized      = def(KindUnauthorized, "E0001", "認証情報が不正です。")
	UserAlreadyExists = def(KindConflict, "E0006", "すでに存在するFirebase UIDまたはEmailです。")
	UserDeleted       = def(KindUnauthorized, "E0007", "退会済みのユーザです。")
	UserNotFound      = def(KindUnauthorized, "E0008", "未登録ユーザです。")

	// リクエスト全般
	BadRequest = def(KindBadRequest, "E0002", "リクエストが不正です。")

	// 規約
	PolicyNotFound        = def(KindBadRequest, "E0003", "規約が存在しません。")
	InvalidPolicyType     = def(KindBadRequest, "E0004", "規約タイプが不正です。")
	InvalidPolicyVersion  = def(KindBadRequest, "E0005", "規約バージョンが不正です。")
	PolicyRequired        = def(KindBadRequest, "E0017", "規約情報は必須です。")
	AgreedVersionRequired = def(KindBadRequest, "E0018", "規約同意バージョンは必須です。")

	// 書誌
	MainTitleRequired      = def(KindBadRequest, "E0009", "主題は必須項目です。")
	RecordTypeRequired     = def(KindBadRequest, "E0010", "書誌タイプ指定は必須です。")
	InvalidRecordType      = def(KindBadRequest, "E0011", "書誌タイプ指定が不正です。")
	ContainerNotAllowed    = def(KindBadRequest, "E0012", "この書誌タイプは収録物を指定できません。")
	InvalidPublicationDate = def(KindBadRequest, "E0013", "出版日付の指定が不正です。")
	InvalidReadStatus      = def(KindBadRequest, "E0014", "読了ステータスが不正です。")
	RecordIdRequired       = def(KindBadRequest, "E0023", "書誌IDは必須です。")

	// RecordNotFound 他ユーザの書誌を指定された場合もこれを返す
	// (403と区別すると他ユーザの書誌IDの存在が漏れるため)
	RecordNotFound = def(KindNotFound, "E0024", "指定された書誌情報が存在しません。")

	// 書誌の構成要素
	InvalidContributorRole  = def(KindBadRequest, "E0015", "貢献者役割が不正です。")
	ContributorNameRequired = def(KindBadRequest, "E0019", "貢献者の氏名は必須です。")
	ContributorRoleRequired = def(KindBadRequest, "E0020", "貢献者の役割は必須です。")
	MemoContentRequired     = def(KindBadRequest, "E0021", "メモの内容は必須です。")
	UrlValueRequired        = def(KindBadRequest, "E0022", "URLの値は必須です。")

	// 登録上限
	ReachRegistrationLimit = def(KindBadRequest, "E0016", "DB登録件数上限を超えています。")
)
