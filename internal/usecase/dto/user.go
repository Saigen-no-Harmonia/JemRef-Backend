package dto

import (
	"jemref_go/internal/domain/policy"
	"jemref_go/internal/domain/user"
)

type CreateUserInput struct {
	FirebaseUserId             string
	Email                      string
	TermsAgreedVersion         string
	PrivacyPolicyAgreedVersion string
}

type CreateUserOutput struct {
	PublicUserId string
}

type DeleteUserInput struct {
	InternalUserid int64
	FirebaseUserId string
}

type GetUserAgreementsOutput struct {
	Agreements []GetUserAgreement
}

type GetUserAgreement struct {
	PolicyType    policy.PolicyType
	Label         string
	LatestVersion string
	AgreedVersion string
	Status        user.PolicyAgreementStatus
}

type UpdateUserAgreementsInput struct {
	InternalUid int64
	Agreements  []UpdateUserAgreement
}

type UpdateUserAgreement struct {
	PolicyType    policy.PolicyType
	AgreedVersion string
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
