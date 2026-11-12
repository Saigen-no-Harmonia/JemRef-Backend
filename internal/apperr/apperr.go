// Package apperr アプリケーションエラーの表現を提供する
//
// エラーには3種類の読み手があり、1つのエラー値でそれぞれに必要な情報を運ぶ。
//   - アプリケーション(フロント): Def.Code  … 機械可読なエラーコード
//   - エンドユーザ:               Def.Message … 公開して安全な文言
//   - 運用者:                     Error()     … ログに出す詳細
//
// HTTPステータスは持たない(Kindという意味的な分類までに留める)。
// Kind からHTTPステータスへの変換は internal/api が行う。
package apperr

// Kind エラーの意味的な分類
//
// ゼロ値は KindUnspecified(未設定)とし、Defの定義漏れをテストで検出できるようにしてある
// (catalog.goの全定義に対する不変条件テストが TestAll_台帳の不変条件)。
// 実行時に KindUnspecified が渡された場合は、api側で安全側に倒して500になる。
type Kind int

const (
	KindUnspecified Kind = iota
	KindInternal
	KindBadRequest
	KindUnauthorized
	KindNotFound
	KindConflict
)

// Def エラーコードと公開メッセージの定義
//
// コードと公開メッセージの唯一の正。catalog.go にすべての定義を置く。
// 複数の層・複数の原因が同じDefを共有してよい(例: 認証失敗の各原因はすべてUnauthorized)。
type Def struct {
	Kind    Kind
	Code    string
	Message string
}

// Error アプリケーションエラー
//
// パッケージ変数として各層で定義し、errors.Is での比較対象にする。
// 共有される値なので、生成後に書き換えてはならない(detailは生成時に確定させる)。
type Error struct {
	def    Def
	detail string
}

// Error ログ用の詳細を返す
//
// レスポンスに出す文言は Def().Message であり、こちらは出力しない。
// 呼び出し側が fmt.Errorf("...: %w", err) でラップした文脈は、
// ラップ後のエラーの Error() に含まれるためログには両方が残る。
func (e *Error) Error() string { return e.detail }

// Def エラーコードと公開メッセージの定義を返す
func (e *Error) Def() Def { return e.def }

// New ログ用の詳細を指定してエラーを生成する
//
// 公開メッセージとログ用の詳細が異なる場合に使う
// (例: 認証失敗の原因はログにだけ残し、レスポンスは原因を明かさない)。
func (d Def) New(detail string) *Error {
	return &Error{def: d, detail: detail}
}

// Err 公開メッセージをログ用の詳細としても使うエラーを生成する
//
// 入力値の不備など、ユーザに見せる文言がそのまま原因の説明になる場合に使う。
func (d Def) Err() *Error {
	return &Error{def: d, detail: d.Message}
}
