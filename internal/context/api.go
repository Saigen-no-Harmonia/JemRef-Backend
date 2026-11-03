package context

import (
	"context"

	"github.com/gin-gonic/gin"
)

// 書誌情報API
const REF_API_001 = "REF-API-001"
const REF_API_002 = "REF-API-002"
const REF_API_003 = "REF-API-003"
const REF_API_004 = "REF-API-004"
const REF_API_005 = "REF-API-005"

// 会員情報API
// MEM-API-001(ユーザ情報参照) / MEM-API-003(ユーザ情報更新) はPh0では使用しない
const MEM_API_002 = "MEM-API-002"
const MEM_API_004 = "MEM-API-004"
const MEM_API_005 = "MEM-API-005"
const MEM_API_006 = "MEM-API-006"
const MEM_API_007 = "MEM-API-007"

// 一般API
const GEN_API_001 = "GEN-API-001"

// SetApiId 以降の処理で参照するAPI IDをcontextに格納する
// 監査列更新のためだが、参照のみのエンドポイントでも統一のため設定する。
func SetApiId(c *gin.Context, apiId string) {
	c.Set(CtxKeyApiId, apiId)
}

// GetApiId 監査列へ設定するAPI IDを取り出す
func GetApiId(ctx context.Context) string {
	apiId, ok := ctx.Value(CtxKeyApiId).(string)
	if !ok || apiId == "" {
		panic("api id is not set in context")
	}

	return apiId
}
