package middleware

import (
	ctxutil "jemref_go/internal/context"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO firebaseトークン認証に置き換える
		c.Set(ctxutil.CtxKeyFirebaseUid, "firebaseuid")

		// 仮のユーザID
		c.Set(ctxutil.CtxKeyUid, "dummy-user")

		// 仮のメールアドレス
		c.Set(ctxutil.CtxtKeyEmail, "dummy@example.com")

		c.Next()
	}
}

func Delete(firebaseUid string) error {
	// 仮
	return nil
}
