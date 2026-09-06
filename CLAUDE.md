# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

JemRef(和文文献の収集・整理に特化した書誌情報管理ツール)のバックエンド。Go + gin。Firebase Authでの認証、MySQLでの永続化、Swaggerでのドキュメント公開を行う。

## コマンド

```sh
make run                 # go run cmd/server/main.go でローカル起動
make test                # go test ./...
make test-repository     # DB統合テストのみ(./internal/infrastructure/db)
make swagger             # swag init でSwagger定義を再生成(要: go install github.com/swaggo/swag/cmd/swag@latest)
make docs                # public/docs/api/ にSwagger UI一式を配置
make cover-record        # ./internal/domain/record のカバレッジ計測 + go tool cover -func
make cover-usecase       # 同上(./internal/usecase)
make cover-handler       # 同上(./internal/handler)
make cover-middleware    # 同上(./internal/middleware)
```

単一パッケージ・単一テストの実行:

```sh
go test ./internal/middleware/...
go test ./internal/middleware/... -run TestMiddleware_CommonLogger -v
```

`internal/infrastructure/db` 配下のテストは `testcontainers-go` で実MySQLコンテナを起動する統合テストのため、**Dockerが起動している環境が必須**(`go test ./...` はここも含む)。`sql/ddl.sql` を読み込んでスキーマを作成する(`internal/infrastructure/db/main_test.go`)。

ハンドラのSwaggerアノテーション(`// @Summary` 等)を変更したら `make swagger` で `docs/` を再生成する。

## 環境変数

`internal/config` の `Load()` が `godotenv` で `.env` を読み込み、`DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASS` / `DB_NAME` / `ENV` / `LOG_LEVEL` を必須として取得する。**未設定の変数が1つでもあると `log.Fatalf` で即座にプロセスが落ちる**。ほかに Firebase Admin SDK が `GOOGLE_CLOUD_PROJECT` / `FIREBASE_AUTH_EMULATOR_HOST`(ローカルはエミュレータ接続)を参照する。雛形は `.env.example`。

- `ENV=local` のとき、ログは `tint` の人間可読形式。それ以外はJSON形式(`internal/middleware/logger.go`)。
- `LOG_LEVEL` は `DEBUG` / `WARN` / `ERROR` を解釈し、それ以外(未知の値・空文字)は `INFO` にフォールバックする。

## CI

- `.github/workflows/test.yml`: push / pull_request で `go test ./...`。**DB統合テストもCIで実行される**ため、CIでだけ落ちる差異(時刻の丸め等)に注意。
- `.github/workflows/docs.yml`: `main` へのpushで `make swagger` を実行し、Swagger UIをGitHub Pagesへデプロイ。

## アーキテクチャ

レイヤードアーキテクチャ。DIコンテナは使わず `cmd/server/main.go` 内の `newXxxHandler`/`newXxxUsecase` 関数で手動配線している。

```
handler (gin) → usecase → repository (interface) → infrastructure/db (実装) or infrastructure/mock (実装)
```

- **internal/domain**: 純粋なドメインモデル・値オブジェクト(`record`, `user`, `policy`, `id`)。
  - 列挙的な値オブジェクトの多く(`record.ReadStatus`, `record.ContributerRole`, `policy.PolicyType`)は `type Xxx string` + id⇔値のマップで実装され、`GetId()` / `XxxFromId()` / `IsValid()` を持つ。`XxxFromId()` は未知のidに対して `panic` する — **呼び出し側が事前に `IsValid()` で検証していることを前提とした設計**(不変条件違反はバグとして落とす方針)。
  - `user.PolicyAgreementStatus` はこの型に当てはまらない例外で、id変換を持たず `ChkAgreementStat(agreedVer, latestVer)` で状態を判定するだけ。
- **internal/repository**: リポジトリのインターフェースと、リポジトリ層共通の sentinel error(`ErrNotFound`, `ErrDuplicateEntry`)を定義。実装は2種類:
  - `internal/infrastructure/db`: `database/sql` を使った実MySQL実装
  - `internal/infrastructure/mock`: **未実装機能を動かすためのスタブ**。状態を保持せず(「mockなので保存はしない」)、固定のサンプル値を返すだけ。実装完了時に削除する。
- **internal/usecase**: ビジネスロジック。インターフェースは `interface.go`、実装は `*_impl.go`、入出力DTOは `usecase/dto`、独自の sentinel error は `usecase/errors.go`。
- **internal/handler**: ginハンドラ。リクエスト/レスポンスDTOは `handler/dto`、独自の sentinel error は `handler/error.go`、ルートパラメータ名は `handler/params.go`。
- **internal/middleware**: 各ミドルウェア固有の sentinel error は同ディレクトリの `error.go`。
  - `auth.go`: `FirebaseAuth`(IDトークン検証 → email/firebaseUid をcontextへ)、`ChkUnregistered`(会員登録APIで「未登録であること」を検証)、`FindCurrentUser`(DBからユーザを引いて internal uid / public uid をcontextへ。退会済みならFirebaseユーザを削除して中断)。
  - `logger.go`: リクエストログ。`/api/v0/health` はログ対象外。テストからログ出力先を差し替えられるよう `CommonLoggerWithWriter(cfg, io.Writer)` を用意し、`CommonLogger(cfg)` はその `os.Stdout` 版。
- **internal/context**: gin.Context に格納する値のキー定数(`keys.go`)と、型安全に取り出す `GetUserId`/`GetFirebaseUid`/`GetPublicUid`/`GetEmail`(`auth.go`)。ミドルウェアが `c.Set` し、以降のハンドラ/ミドルウェアがこれらのgetterで読む。
- **internal/api**: エラーをHTTPレスポンスへ変換する層。`error.go`(`ApiError` 定義)と `error_handler.go`(`ErrorHandler()` ミドルウェアと `toApiError`)。

### エラーハンドリングの流れ

各レイヤーは `c.Error(err)` で gin の `c.Errors` にエラーを積むだけで、HTTPステータスやレスポンス形式を意識しない。最後に `internal/api.ErrorHandler()`(ginミドルウェア)が `c.Errors.Last()` を見て `toApiError` でHTTPレスポンスに変換する。

新しいエラーを追加する際の流れ:
1. 発生元のレイヤー(`middleware`/`handler`/`usecase`)の `errors.go` に sentinel error を定義
2. `internal/api/error_handler.go` の `toApiError` に `errors.Is(err, ...)` の分岐を追加
3. 対応する `ApiError`(ステータスコード・エラーコード・メッセージ)が無ければ `internal/api/error.go` に追加

**2 の追加漏れはコンパイルエラーにならず、`default` の 500(`ErrInternal`)に落ちる。** 別レイヤーの sentinel error を包み直して返す場合(例: usecase のエラーを middleware 側の sentinel error でラップする)は特に、包み直した後のエラーが `toApiError` に登録されているかを必ず確認すること。

### ミドルウェアの登録順序に依存関係がある

`cmd/server/main.go` の `setupMiddleware`: `gin.Recovery()` → `middleware.CommonLogger(cfg)` → `api.ErrorHandler()` の順で登録すること。Loggerは `c.Next()` でハンドラチェーンの完了を待ってからログ出力するため、ErrorHandlerより先に登録し、後続で確定するレスポンス内容(ステータスコード等)を拾えるようにする必要がある。

## テストの型

- **handler / usecase / middleware / domain**: テーブル駆動テスト + `t.Run`(サブテスト名は日本語で `正常_xxx` / `異常_xxx`)。
- **internal/infrastructure/db**: モックを使わず testcontainers-go で実MySQLを起動する統合テスト。こちらはテーブル駆動にせず、`TestUserRepository_Create_正常` のように**1ケース1関数**でケース名を関数名に含める。
- モックは `MockXxx` 構造体に `XxxFunc func(...) (...)` フィールドを持たせ、実メソッドは呼び出し回数(`Called`, `XxxCalled`)と直近の引数(`LastXxx`)を記録してから `XxxFunc` に委譲する、という型で統一されている。テストで使うのは `internal/testutil/mock`(usecase/repositoryインターフェース向け)。`internal/infrastructure/mock` はテスト用ではなく本番配線に差し込むスタブなので混同しない。
- 各レイヤーに `ErrTestUnexpected` という専用の sentinel error があり、「分岐でカバーしていない予期しないエラー」のテストに使う。

## 守ること

- エラーの比較は常に `errors.Is` を使う(`err == ErrXxx` の直接比較はラップされた瞬間に壊れる)。テストのエラー検証も `assert.ErrorIs(t, actualErr, tt.expectedErrorBody)`。
- **DB往復した `time.Time` の比較は `assert.WithinDuration(t, expected, actual, time.Microsecond)` を使う**(`assert.True(t, a.Equal(b))` は不可)。MySQLのカラム精度による丸めで、ローカルでは通ってもCI環境でだけ失敗することがある(実績: コミット `6c571e6`)。
- `XxxFromId()` 系を呼ぶ前に `IsValid()` で検証する(未知のidは `panic` する)。
- 新しい sentinel error を追加したら `toApiError` の分岐追加までをセットで行う。
