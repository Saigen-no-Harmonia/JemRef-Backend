package apperr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// テスト用の定義(台帳の値に依存しないよう独自に用意する)
var (
	testDef      = Def{KindBadRequest, "T0001", "公開メッセージ"}
	testOtherDef = Def{KindNotFound, "T0002", "別の公開メッセージ"}
)

func TestDef_New(t *testing.T) {

	tests := []struct {
		name           string
		def            Def
		detail         string
		expectedDetail string
	}{
		{
			name:           "正常_ログ用の詳細が公開メッセージと異なる",
			def:            testDef,
			detail:         "ログにだけ残す原因",
			expectedDetail: "ログにだけ残す原因",
		},
		{
			name:           "正常_ログ用の詳細が空文字",
			def:            testDef,
			detail:         "",
			expectedDetail: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.def.New(tt.detail)

			// Error()はログ用の詳細を返す(公開メッセージは返さない)
			assert.Equal(t, tt.expectedDetail, actual.Error())
			// 公開用の情報はDef()から取る
			assert.Equal(t, tt.def, actual.Def())
		})
	}
}

// TestDef_Err 公開メッセージをログ用の詳細としても使う
func TestDef_Err(t *testing.T) {
	actual := testDef.Err()

	assert.Equal(t, testDef.Message, actual.Error())
	assert.Equal(t, testDef, actual.Def())
}

// TestDef_同じDefから生成しても別の値になる
// パッケージ変数として共有するため、生成ごとに独立した値であることを確認する。
func TestDef_同じDefから生成しても別の値になる(t *testing.T) {
	a := testDef.New("詳細A")
	b := testDef.New("詳細B")

	assert.NotSame(t, a, b)
	assert.Equal(t, "詳細A", a.Error())
	assert.Equal(t, "詳細B", b.Error())
}

func TestError_errorsIs(t *testing.T) {

	sentinel := testDef.New("sentinel")
	other := testOtherDef.New("other")

	tests := []struct {
		name     string
		err      error
		target   error
		expected bool
	}{
		{
			name:     "正常_同一のsentinelは一致する",
			err:      sentinel,
			target:   sentinel,
			expected: true,
		},
		{
			name:     "正常_ラップしても一致する",
			err:      fmt.Errorf("文脈: %w", sentinel),
			target:   sentinel,
			expected: true,
		},
		{
			name:     "正常_多段にラップしても一致する",
			err:      fmt.Errorf("階層1: %w", fmt.Errorf("階層2: %w", sentinel)),
			target:   sentinel,
			expected: true,
		},
		{
			name:     "異常_別のsentinelとは一致しない",
			err:      sentinel,
			target:   other,
			expected: false,
		},
		{
			// 同じDefから生成しても別の値なので一致しない(Codeでの一致判定はしない)
			name:     "異常_同じDefから別途生成した値とは一致しない",
			err:      sentinel,
			target:   testDef.New("sentinel"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, errors.Is(tt.err, tt.target))
		})
	}
}

func TestError_errorsAs(t *testing.T) {

	sentinel := testDef.New("詳細")

	tests := []struct {
		name        string
		err         error
		expectFound bool
		expectedDef Def
	}{
		{
			name:        "正常_そのままのエラーから取り出せる",
			err:         sentinel,
			expectFound: true,
			expectedDef: testDef,
		},
		{
			name:        "正常_ラップされたエラーから取り出せる",
			err:         fmt.Errorf("文脈: %w", sentinel),
			expectFound: true,
			expectedDef: testDef,
		},
		{
			name:        "異常_apperrでないエラーからは取り出せない",
			err:         errors.New("ただのエラー"),
			expectFound: false,
		},
		{
			// %v でラップした場合は連鎖に入らないため取り出せない
			name:        "異常_文字列化されたエラーからは取り出せない",
			err:         fmt.Errorf("文脈: %v", sentinel),
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual *Error
			found := errors.As(tt.err, &actual)

			assert.Equal(t, tt.expectFound, found)
			if !tt.expectFound {
				return
			}
			require.NotNil(t, actual)
			assert.Equal(t, tt.expectedDef, actual.Def())
		})
	}
}

// TestKind_ゼロ値はUnspecified
// Kindを指定し忘れた定義を不変条件テストで検出できるよう、ゼロ値は未設定を表す。
func TestKind_ゼロ値はUnspecified(t *testing.T) {
	var zero Kind

	assert.Equal(t, KindUnspecified, zero)
}

// -------------台帳(catalog.go)の不変条件-------------

// TestAll_台帳の不変条件
//
// 台帳を走査して検証するため、エラーを追加してもこのテストは修正不要。
//
// 期待値は台帳自身から読むことになるため「E0009が400か」のような値の正しさは
// 検証できない(トートロジーになる)。ここで守るのは構造的な健全性のみ。
// コードとステータスの対応はフロントとの契約であり、レビューと実機確認で担保する。
func TestAll_台帳の不変条件(t *testing.T) {

	all := All()
	require.NotEmpty(t, all, "台帳が空(def()経由で宣言されていない可能性がある)")

	seen := make(map[string]bool, len(all))

	for _, d := range all {
		t.Run(d.Code, func(t *testing.T) {
			// Kindの指定漏れ(ゼロ値)を検出する。実行時はapi側で500に倒れるため気づけない
			assert.NotEqual(t, KindUnspecified, d.Kind, "Kindが未設定")

			// システムエラーはF、業務エラーはEで始まる4桁
			assert.Regexp(t, `^[EF]\d{4}$`, d.Code, "エラーコードの形式が不正")

			// レスポンスに空文字を返さない
			assert.NotEmpty(t, d.Message, "公開メッセージが空")
		})

		assert.False(t, seen[d.Code], "エラーコードが重複している: %s", d.Code)
		seen[d.Code] = true
	}
}
