package middleware

import (
	"context"
	"errors"
	"fmt"
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/usecase"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

type FirebaseAuthClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// FirebaseAuth 要ログインAPI呼び出し時にコールする共通認証処理。
func FirebaseAuth(client FirebaseAuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorizationヘッダの取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(ErrRequireAuthHeader)
			c.Abort()
			return
		}

		// Bearer<token>のチェック
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Error(ErrRequireBearerToken)
			c.Abort()
			return
		}
		idToken := parts[1]

		// Firebaseでの認証
		token, err := client.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			c.Error(fmt.Errorf("%w: %v", ErrInvalidFirebaseToken, err))
			c.Abort()
			return
		}

		// メールアドレス取得
		email, ok := token.Claims["email"].(string)
		if !ok {
			c.Error(ErrRequireEmail)
			c.Abort()
			return
		}

		c.Set(ctxutil.CtxKeyEmail, email)
		c.Set(ctxutil.CtxKeyFirebaseUid, token.UID)
	}
}

// ChkUnregistered ユーザ未登録であることを検証する
func ChkUnregistered(au usecase.AuthUsecase) gin.HandlerFunc {

	return func(c *gin.Context) {
		firebaseUid, ok := ctxutil.GetFirebaseUid(c)
		if !ok {
			c.Error(ErrRequireFirebaseUid)
			c.Abort()
			return
		}

		_, err := au.Authenticate(c.Request.Context(), firebaseUid)
		// 正常（既存ユーザなし）
		if err == usecase.ErrUserNotFound {
			return

			// エラーがない場合は会員登録済み
		} else if err == nil {
			c.Error(fmt.Errorf("会員登録済みのユーザです。firebase uid=%s: %w : %v", firebaseUid, ErrUserAlreadyExists, err))
			c.Abort()
			return

			// その他のエラー
		} else if !errors.Is(err, usecase.ErrUserNotFound) {
			c.Error(fmt.Errorf("予期せぬエラー firebase uid=%s : %w", firebaseUid, err))
			c.Abort()
			return
		}
	}
}

// FindCurrentUser DBからユーザ情報を取得しContextに格納する
func FindCurrentUser(au usecase.AuthUsecase) gin.HandlerFunc {

	return func(c *gin.Context) {
		firebaseUid, ok := ctxutil.GetFirebaseUid(c)
		if !ok {
			c.Error(ErrRequireFirebaseUid)
			c.Abort()
			return
		}

		// ユーザ情報をDBから取得
		authUser, err := au.Authenticate(c.Request.Context(), firebaseUid)
		if err != nil {
			// 退会済みユーザだった場合、Firebaseユーザを削除して終了
			if errors.Is(err, usecase.ErrUserDeleted) {
				_ = au.DeleteUser(
					c.Request.Context(),
					firebaseUid,
				)
				c.Error(fmt.Errorf("すでに退会済みのユーザです。 FIrebaseユーザを削除しました。firebaseUid=%s: %w", firebaseUid, ErrUserDeleted))
				c.Abort()
				return
			}

			c.Error(fmt.Errorf("予期せぬエラー firebase uid=%s :%w", firebaseUid, err))
			c.Abort()
			return
		}

		c.Set(ctxutil.CtxKeyUid, authUser.InternalUserId)
		c.Set(ctxutil.CtxKeyPublicUid, authUser.PublicUserId)
	}
}
