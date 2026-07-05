package dto

type CreateUserInput struct {
	FirebaseUserId string
	Email          string
}

type CreateUserOutput struct {
	PublicUserId string
}

type DeleteUserInput struct {
	InternalUserid int64
	FirebaseUserId string
}

type UserLoginInput struct {
	InternalUserId int64
}

type UserLoginOutput struct {
	PublicUserId string
}

type AuthUserOutput struct {
	InternalUserId int64
	FirebaseUserId string
	PublicUserId   string
}
