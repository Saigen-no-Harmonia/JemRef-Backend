package handler

import (
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/usecase"

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
	// r.GET("users/:id, h.GetUser")
	r.POST("/users", h.CreateUser)
	// r.PUT("users/:id, h.UpdateUser")
	// r.DELETE("users/:id, h.DelteUser")
}

// [MEM-API-002] ユーザ情報登録 /user POST
// リクエストされたユーザを作成する
func (h *UserHandler) CreateUser(c *gin.Context) {

	var req CreateUserRequest

	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	c.JSON(400, gin.H{"error": err.Error()})
	// 	logger.Info("failed to bind request")
	// 	return
	// }

	// firebaseユーザIDを受け取る
	firebaseUserId, ok := ctxutil.GetUserId(c)
	if !ok {
		c.JSON(500, gin.H{"error": "failed to get firebase userid"})
		return
	}

	// メールアドレスを受け取る
	email, ok := ctxutil.GetEmail(c)
	if !ok {
		c.JSON(500, gin.H{"error": "failed to get email address"})
		return
	}

	// usecase用構造体を作成
	input := createUserInput(req, firebaseUserId, email)

	output, err := h.usecase.CreateUser(input)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// レスポンスを生成
	res := CreateUserResponse{
		PublicUserId: output.PublicUserId,
	}

	c.JSON(201, res)
}

// Useacse用の構造体にマッピングする
func createUserInput(req CreateUserRequest, firebaseUserId string, email string) usecase.CreateUserInput {
	return usecase.CreateUserInput{
		FirebaseUserId: firebaseUserId,
		Email:          email,
	}
}
