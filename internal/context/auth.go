package context

import (
	"github.com/gin-gonic/gin"
)

func GetUserId(c *gin.Context) (int64, bool) {
	// 存在チェック
	val, exists := c.Get(CtxKeyUid)
	if !exists {
		return 0, false
	}

	// 型チェック
	uid, ok := val.(int64)
	if val == 0 {
		return 0, false
	}
	return uid, ok
}

func GetFirebaseUid(c *gin.Context) (string, bool) {
	// 存在チェック
	val, exists := c.Get(CtxKeyFirebaseUid)
	if !exists {
		return "", false
	}

	firebaseUid, ok := val.(string)
	if val == "" {
		return "", false
	}

	return firebaseUid, ok
}

func GetPublicUid(c *gin.Context) (string, bool) {
	// 存在チェック
	val, exists := c.Get(CtxKeyPublicUid)
	if !exists {
		return "", false
	}

	publicUid, ok := val.(string)
	if val == "" {
		return "", false
	}

	return publicUid, ok
}

func GetEmail(c *gin.Context) (string, bool) {
	// 存在チェック
	val, exists := c.Get(CtxKeyEmail)
	if !exists {
		return "", false
	}

	// 型チェック
	email, ok := val.(string)
	if val == "" {
		return "", false
	}
	return email, ok
}
