package handler

type CreateUserRequest struct {
	// リクエストパラメータは現状特になし
}

type CreateUserResponse struct {
	PublicUserId string `json:"user_id"`
}
