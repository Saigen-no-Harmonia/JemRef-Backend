package handler

import (
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/handler/dto"
	"jemref_go/internal/usecase"
	"log"

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
	r.POST("/users", h.CreateUser)
	// r.PUT("/users/:id", h.UpdateUser)
	r.DELETE("/users/:id", h.DeleteUser)
	r.GET("/users/agreements", h.GetUserAgreements)
	r.PUT("/users/agreement", h.UpdateUserAgreements)
	r.PUT("/users/login", h.Login)
}

// [MEM-API-002] ユーザ情報登録 /users POST
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
	firebaseUserId, ok := ctxutil.GetUserId(c)
	if !ok {
		c.JSON(500, dto.ErrorResponse{
			Code:    "F0001",
			Message: "Firebase UIDの取得に失敗しました。",
		})
		return
	}

	// メールアドレスを受け取る
	email, ok := ctxutil.GetEmail(c)
	if !ok {
		c.JSON(500, dto.ErrorResponse{
			Code:    "F0002",
			Message: "FirebaseからのEメールアドレス取得に失敗しました。"})
		return
	}

	// usecase用構造体を作成
	input := createUserInput(firebaseUserId, email)

	// usecaseを呼び出し
	output, err := h.usecase.CreateUser(c.Request.Context(), input)
	if err != nil {

		log.Println(err)

		c.JSON(500, dto.ErrorResponse{
			Code:    "F0003",
			Message: "DB更新処理に失敗しました。",
		})
		return
	}

	// レスポンスを生成
	res := dto.CreateUserResponse{
		PublicUserId: output.PublicUserId,
	}

	c.JSON(201, res)
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
// @Success 200 {object} dto.StatusResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {}

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
func (h *UserHandler) Login(c *gin.Context) {}

// Useacse用の構造体にマッピングする
func createUserInput(firebaseUserId string, email string) usecase.CreateUserInput {
	return usecase.CreateUserInput{
		FirebaseUserId: firebaseUserId,
		Email:          email,
	}
}
