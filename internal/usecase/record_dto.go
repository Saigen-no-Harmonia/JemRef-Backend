package usecase

type CreateRecordInput struct {
	UserId    string
	MainTitle string
}

type CreateRecordOutput struct {
	UserId string
}
