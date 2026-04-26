package middleware

import "github.com/gin-gonic/gin"

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO firebaseトークン認証に置き換える

		// 仮のユーザID
		c.Set("user_id", "dummy-user")

		c.Next()
	}

}