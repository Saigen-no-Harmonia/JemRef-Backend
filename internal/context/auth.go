package context

import (
	"github.com/gin-gonic/gin"
)

func GetFirebaseUid(c *gin.Context) (string, bool) {
	// 存在チェック
	val, exists := c.Get(CtxKeyFirebaseUid)
	if !exists {
		return "", false
	}

	firebaseUid, ok := val.(string)
	return firebaseUid, ok
}

func GetUserId(c *gin.Context) (string, bool) {
	// 存在チェック
	val, exists := c.Get(CtxKeyUid)
	if !exists {
		return "", false
	}

	// 型チェック
	uid, ok := val.(string)
	return uid, ok
}

func GetEmail(c *gin.Context) (string, bool) {
	// 存在チェック
	val, exists := c.Get(CtxtKeyEmail)
	if !exists {
		return "", false
	}

	// 型チェック
	email, ok := val.(string)
	return email, ok
}
