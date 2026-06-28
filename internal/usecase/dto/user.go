package dto

type CreateUserInput struct {
	FirebaseUserId string
	Email          string
}

type CreateUserOutput struct {
	PublicUserId string
}

type DeleteUserInput struct {
	InternalUserid uint64
	FirebaseUserId string
}

type UserLoginInput struct {
	InternalUserId uint64
}

type UserLoginOutput struct {
	PublicUserId string
}

type AuthUserOutput struct {
	InternalUserId uint64
	FirebaseUserId string
	PublicUserId   string
}
