package context

import "github.com/gin-gonic/gin"

func GetUserId(c *gin.Context) (string, bool) {
	// 存在チェック
	val, exists := c.Get(ContextKeyUserId)
	if !exists {
		return "", false
	}

	// 型チェック
	uid, ok := val.(string)
	return uid, ok
}

func GetEmail(c *gin.Context) (string, bool) {
	// 存在チェック
	val, exists := c.Get(ContextKeyEmail)
	if !exists {
		return "", false
	}

	// 型チェック
	email, ok := val.(string)
	return email, ok
}
