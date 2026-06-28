package user

import "time"

type User struct {
	InternalUserId        uint64
	PublicUserId          string
	FirebaseUserId        string
	Email                 string
	TermsAgreedDt         *time.Time
	TermsVersion          string
	PrivacyPolicyAgreedDt *time.Time
	PrivacyPolicyVersion  string
	DeletedAt             *time.Time
}
