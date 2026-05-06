package middleware

import (
	ctxutil "jemref_go/internal/context"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO firebaseトークン認証に置き換える

		// 仮のユーザID
		c.Set(ctxutil.ContextKeyUserId, "dummy-user")

		// 仮のメールアドレス
		c.Set(ctxutil.ContextKeyEmail, "dummy@example.com")

		c.Next()
	}

}
