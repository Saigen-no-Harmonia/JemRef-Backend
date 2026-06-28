package handler

import (
	"errors"
	"jemref_go/internal/context"
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/handler/dto"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/usecase"
	usecaseDto "jemref_go/internal/usecase/dto"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	usecase *usecase.UserUsecase
}

// コンストラクタ
func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{usecase: uc}
}

// 共通認証ルート
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	// r.GET("/users/:id", h.GetUser)
	// r.PUT("/users/:id", h.UpdateUser)
	r.DELETE("/users/:id", h.DeleteUser)
	r.GET("/users/agreements", h.GetUserAgreements)
	r.PUT("/users/agreement", h.UpdateUserAgreements)
	r.PUT("/users/login", h.Login)
}

// [MEM-API-002] ユーザ情報登録 /join POST
//
// @Summary [MEM-API-002] ユーザ情報登録
// @Description ユーザを作成する。ユーザ情報はfirebase tokenから取得する。
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param param body dto.CreateUserRequest true "ユーザ情報登録リクエスト"
// @Success 201 {object} dto.CreateUserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {

	// firebaseユーザIDを受け取る
	firebaseUserId, ok := ctxutil.GetFirebaseUid(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "F0001",
			Message: "FirebaseUIDの取得処理に異常があります。",
		})
		return
	}

	// メールアドレスを受け取る
	email, ok := ctxutil.GetEmail(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "F0002",
			Message: "Eメールアドレスの取得処理に異常があります。"})
		return
	}

	// usecase用構造体を作成
	input := createUserInput(firebaseUserId, email)

	// usecaseを呼び出し
	output, err := h.usecase.CreateUser(c.Request.Context(), input)
	if err != nil {
		log.Println(err)

		// 既存ユーザがいた場合

		// 上記以外の場合
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "F0003",
			Message: "ユーザ作成処理に失敗しました。",
		})
		return
	}

	// レスポンスを生成
	res := dto.CreateUserResponse{
		PublicUserId: output.PublicUserId,
	}

	c.JSON(http.StatusCreated, res)
}

// [MEM-API-004] ユーザ退会 /users/:id DELETE
//
// @Summary [MEM-API-004] ユーザ退会
// @Description ユーザ情報を論理削除する。
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param id path string true "(公開用)ユーザID"
// @Success 200 {object}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	internalUid, ok := ctxutil.GetUserId(c)
	if !ok {
		log.Print("internal user idの取得に失敗")
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "F0001",
			Message: "failed to get internal user id",
		})
	}
	firebaseUid, ok := ctxutil.GetFirebaseUid(c)
	if !ok {
		log.Print("firebase user idの取得に失敗")
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "F0001",
			Message: "failed to get firebase user id",
		})
	}

	input := createDeleteUserInput(internalUid, firebaseUid)

	err := h.usecase.DeleteUser(c.Request.Context(), input)
	if err != nil {
		// FirebaseユーザはいるがDBにデータが物理的に存在しない：未入会
		if errors.Is(err, usecase.ErrUserNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Code:    "Exxxx",
				Message: "Unauthorized",
			})
			return
		}
		// ユーザ情報に不整合がある場合
		if errors.Is(err, usecase.ErrInvalidUser) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
				Code:    "F0001",
				Message: "fatal: invalid user data",
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "F0001",
			Message: "fatal: internal server error",
		})
	}

	c.Status(http.StatusOK)
}

// [MEM-API-005] ユーザ規約同意状況参照 /users/agreements GET
//
// @Summary [MEM-API-005] ユーザ規約同意状況参照
// @Description 指定したユーザの規約同意状況を参照する。ユーザ情報はヘッダから取得する。
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Success 200 {object} dto.GetUserAgreementsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/agreements [get]
func (h *UserHandler) GetUserAgreements(c *gin.Context) {}

// [MEM-API-006] ユーザ規約同意状況更新 users/agreements PUT
//
// @Summary [MEM-API-006] ユーザ規約同意状況更新
// @Description 指定したユーザの規約同意状況を更新する。ユーザ情報はヘッダから取得する。
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param policies body dto.UpdateUserAgreementsRequest true "規約同意更新リクエスト"
// @Success 200 {object} dto.StatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/agreements [put]
func (h *UserHandler) UpdateUserAgreements(c *gin.Context) {}

// [MEM-API-007]ユーザログイン users/login PUT
//
// @Summary [MEM-API-007]ユーザログイン
// @Description ユーザログイン処理を実施する。Ph0では公開用ユーザIDを返却するだけ。
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Success 200 {object} dto.UserLoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/login [put]
func (h *UserHandler) Login(c *gin.Context) {

	// id取得（共通認証処理でcontextにセット済み）
	publicUid, ok := context.GetPublicUid(c)
	if !ok {
		// 意図しないエラー
		log.Println("fatal:ctxに公開用ユーザIDが存在しませんでした。")
		c.JSON(
			http.StatusInternalServerError,
			handlerDto.ErrorResponse{
				Code:    "F0001",
				Message: "fatal: internal server error",
			},
		)
		return
	}

	res := createLoginResponse(publicUid)
	c.JSON(http.StatusOK, res)
}

// createUserInput ユーザ作成Usecaseの引数を作成
func createUserInput(firebaseUserId string, email string) usecaseDto.CreateUserInput {
	return usecaseDto.CreateUserInput{
		FirebaseUserId: firebaseUserId,
		Email:          email,
	}
}

// createDeleteUserInput ユーザ削除Usecaseの引数を作成
func createDeleteUserInput(internalUid uint64, firebaseUid string) usecaseDto.DeleteUserInput {
	return usecaseDto.DeleteUserInput{
		InternalUserid: internalUid,
		FirebaseUserId: firebaseUid,
	}
}

// createLoginResponse Loginのレスポンスを生成
func createLoginResponse(publicUid string) handlerDto.UserLoginResponse {
	return handlerDto.UserLoginResponse{
		PublicUserId: publicUid,
	}
}
