package dto

type CreateUserInput struct {
	FirebaseUserId string
	Email          string
}

type CreateUserOutput struct {
	PublicUserId string
}
