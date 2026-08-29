package handler

import (
	"fmt"
	"jemref_go/internal/context"
	ctxutil "jemref_go/internal/context"
	"jemref_go/internal/domain/policy"
	"jemref_go/internal/handler/dto"
	handlerDto "jemref_go/internal/handler/dto"
	"jemref_go/internal/usecase"
	usecaseDto "jemref_go/internal/usecase/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	usecase usecase.UserUsecase
}

// コンストラクタ
func NewUserHandler(uc usecase.UserUsecase) *UserHandler {
	return &UserHandler{usecase: uc}
}

// 共通認証ルート
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	// r.GET("/users/:id", h.GetUser)
	// r.PUT("/users/:id", h.UpdateUser)
	r.DELETE("/users", h.DeleteUser)
	r.GET("/users/agreements", h.GetUserAgreements)
	r.PUT("/users/agreements", h.UpdateUserAgreements)
	r.PUT("/users/login", h.Login)
}

// @Summary [MEM-API-002] ユーザ情報登録
// @Description ユーザを作成する。ユーザ情報はfirebase tokenから取得する。
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param param body dto.CreateUserRequest true "ユーザ情報登録リクエスト"
// @Success 201 {object} dto.CreateUserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /join [post]
func (h *UserHandler) CreateUser(c *gin.Context) {

	firebaseUserId, ok := ctxutil.GetFirebaseUid(c)
	if !ok {
		c.Error(ErrFirebaseUidNotFound)
		return
	}

	email, ok := ctxutil.GetEmail(c)
	if !ok {
		c.Error(fmt.Errorf("メールアドレス取得失敗 firebase uid = %s, %w", firebaseUserId, ErrEmailNotFound))
		return
	}

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(fmt.Errorf("%w: %v", ErrInvalidRequestBody, err))
		return
	}

	input := createUserInput(firebaseUserId, email, req)
	output, err := h.usecase.Create(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	res := dto.CreateUserResponse{
		PublicUserId: output.PublicUserId,
	}

	c.JSON(http.StatusCreated, res)
}

// @Summary [MEM-API-004] ユーザ退会
// @Description ユーザ情報を論理削除する。
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param id path string true "(公開用)ユーザID"
// @Success 200
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	internalUid, ok := ctxutil.GetUserId(c)
	if !ok {
		c.Error(ErrInternalUidNotFound)
		return
	}

	firebaseUid, ok := ctxutil.GetFirebaseUid(c)
	if !ok {
		c.Error(ErrFirebaseUidNotFound)
		return
	}

	input := createDeleteUserInput(internalUid, firebaseUid)

	err := h.usecase.Delete(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}

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
func (h *UserHandler) GetUserAgreements(c *gin.Context) {
	internalUid, ok := ctxutil.GetUserId(c)
	if !ok {
		c.Error(ErrInternalUidNotFound)
		return
	}

	output, err := h.usecase.GetUserAgreements(c.Request.Context(), internalUid)
	if err != nil {
		c.Error(err)
		return
	}

	res := createGetUserAgreementsResponse(*output)
	c.JSON(http.StatusOK, res)
}

// @Summary [MEM-API-006] ユーザ規約同意状況更新
// @Description 指定したユーザの規約同意状況を更新する。ユーザ情報はヘッダから取得する。
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization  header string true "Bearer <firebase_id_token>"
// @Param policies body dto.UpdateUserAgreementsRequest true "規約同意更新リクエスト"
// @Success 200
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/agreements [put]
func (h *UserHandler) UpdateUserAgreements(c *gin.Context) {

	uid, ok := ctxutil.GetUserId(c)
	if !ok {
		c.Error(ErrInternalUidNotFound)
		return
	}

	var req handlerDto.UpdateUserAgreementsRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.Error(fmt.Errorf("%w: %v", ErrInvalidRequestBody, err))
		return
	}

	if len(req.Policies) == 0 {
		c.Error(ErrPolicyRequired)
		return
	}

	input := createUpdateUserAgreementInput(uid, req)

	err := h.usecase.UpdateUserAgreements(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}

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
	publicUid, ok := context.GetPublicUid(c)
	if !ok {
		c.Error(ErrPublicUidNotFound)
		return
	}

	res := createLoginResponse(publicUid)
	c.JSON(http.StatusOK, res)
}

// createUserInput ユーザ作成Usecaseの引数を作成
func createUserInput(firebaseUid string, email string, req dto.CreateUserRequest) usecaseDto.CreateUserInput {
	return usecaseDto.CreateUserInput{
		FirebaseUserId:             firebaseUid,
		Email:                      email,
		TermsAgreedVersion:         req.TermsAgreedVersion,
		PrivacyPolicyAgreedVersion: req.PrivacyPolicyAgreedVersion,
	}
}

// createDeleteUserInput ユーザ削除Usecaseの引数を作成
func createDeleteUserInput(internalUid int64, firebaseUid string) usecaseDto.DeleteUserInput {
	return usecaseDto.DeleteUserInput{
		InternalUserid: internalUid,
		FirebaseUserId: firebaseUid,
	}
}

// createGetUserAgreementsResponse ユーザ規約同意状況取得レスポンスを生成
// Ph0では、ユーザ規約とプラポリだけがある想定で簡易実装
func createGetUserAgreementsResponse(output usecaseDto.GetUserAgreementsOutput) dto.GetUserAgreementsResponse {
	var res dto.GetUserAgreementsResponse

	terms := output.Agreements[0]
	privacyPolicy := output.Agreements[1]

	t := dto.UserAgreementResponse{
		PolicyType:    string(terms.PolicyType),
		Label:         terms.Label,
		LatestVersion: terms.LatestVersion,
		AgreedVersion: terms.AgreedVersion,
		Status:        terms.Status,
	}

	p := dto.UserAgreementResponse{
		PolicyType:    string(privacyPolicy.PolicyType),
		Label:         privacyPolicy.Label,
		LatestVersion: privacyPolicy.LatestVersion,
		AgreedVersion: privacyPolicy.AgreedVersion,
		Status:        privacyPolicy.Status,
	}

	res.Agreements = append(res.Agreements, t)
	res.Agreements = append(res.Agreements, p)

	return res
}

// createUpdateUserAgreementInput ユーザ規約同意状況更新usecaseの引数を作成
func createUpdateUserAgreementInput(internalUid int64, req handlerDto.UpdateUserAgreementsRequest) usecaseDto.UpdateUserAgreementsInput {
	input := usecaseDto.UpdateUserAgreementsInput{
		InternalUid: internalUid,
	}

	for _, p := range req.Policies {
		input.Agreements = append(input.Agreements,
			usecaseDto.UpdateUserAgreement{
				PolicyType:    policy.PolicyType(p.PolicyType),
				AgreedVersion: p.Version,
			},
		)
	}
	return input
}

// createLoginResponse Loginのレスポンスを生成
func createLoginResponse(publicUid string) handlerDto.UserLoginResponse {
	return handlerDto.UserLoginResponse{
		PublicUserId: publicUid,
	}
}
