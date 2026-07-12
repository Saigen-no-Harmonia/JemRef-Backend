package middleware

import (
	"context"
	"errors"
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/usecase"
	"log"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
)

// FirebaseAuth 要ログインAPI呼び出し時にコールする共通認証処理。
func FirebaseAuth(app *firebase.App) gin.HandlerFunc {

	client, _ := app.Auth(context.Background())

	return func(c *gin.Context) {
		// Authorizationヘッダの取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("fatal: Anthorizationヘッダ取得に失敗しました")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Bearer<token>のチェック
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Println("fatal: Bearerヘッダの形式が不正です")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid header format"})
			return
		}
		idToken := parts[1]

		// Firebaseでの認証
		token, err := client.VerifyIDToken(c.Request.Context(), idToken)
		if err != nil {
			log.Println("Firebase認証に失敗しました")
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// メールアドレス取得
		email, ok := token.Claims["email"].(string)
		if !ok {
			log.Printf("メールアドレスの取得に失敗しました。firebaseUid: %s", token.UID)
		}

		c.Set(ctxutil.CtxKeyEmail, email)
		c.Set(ctxutil.CtxKeyFirebaseUid, token.UID)
	}
}

// ChkUnregistered ユーザ未登録であることを検証する
func ChkUnregistered(au *usecase.AuthUsecaseImpl) gin.HandlerFunc {

	return func(c *gin.Context) {
		firebaseUid, ok := ctxutil.GetFirebaseUid(c)
		if !ok {
			log.Printf("fatal: firebase uidの取得処理に異常があります")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "failed to get firebase uid"})
			return
		}

		_, err := au.Authenticate(c.Request.Context(), firebaseUid)
		// DBにユーザデータがある場合（会員登録済み）
		if err == nil {
			log.Printf("会員登録済みのユーザです。firebase uid: %s", firebaseUid)
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "user has been registered"})
			return
			// not found以外のエラー
		} else if !errors.Is(err, usecase.ErrUserNotFound) {
			log.Printf("fatal: db error")
			log.Print(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
	}
}

// FindCurrentUser DBからユーザ情報を取得しContextに格納する
func FindCurrentUser(au *usecase.AuthUsecaseImpl) gin.HandlerFunc {

	return func(c *gin.Context) {
		firebaseUid, ok := ctxutil.GetFirebaseUid(c)
		if !ok {
			log.Printf("fatal: firebase uidの取得処理に異常があります")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "failed to get firebase uid"})
			return
		}

		// ユーザ情報をDBから取得
		authUser, err := au.Authenticate(c.Request.Context(), firebaseUid)
		if err != nil {
			// ユーザ情報が存在しない場合
			if errors.Is(err, usecase.ErrUserNotFound) {
				log.Printf("DBのユーザ情報取得に失敗しました。firebaseUid: %s", firebaseUid)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user has not registered"})
				return
			}

			// 退会済みユーザだった場合、Firebaseユーザを削除して終了
			if errors.Is(err, usecase.ErrUserDeleted) {
				log.Printf("ERROR: すでに退会済みのユーザです。 FIrebaseユーザ削除処理を実施します。firebaseUid: %s", firebaseUid)
				_ = au.DeleteUser(
					c.Request.Context(),
					firebaseUid,
				)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user has deleted"})
			}

			// その他のエラー
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.Set(ctxutil.CtxKeyUid, authUser.InternalUserId)
		c.Set(ctxutil.CtxKeyPublicUid, authUser.PublicUserId)
	}
}

// StubAuth 旧共通認証スタブ
func StubAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO firebaseトークン認証に置き換える
		c.Set(ctxutil.CtxKeyFirebaseUid, "firebaseuid")

		// 仮のユーザID
		c.Set(ctxutil.CtxKeyUid, "dummy-user")

		// 仮のメールアドレス
		c.Set(ctxutil.CtxKeyEmail, "dummy@example.com")

		c.Next()
	}
}
