package user

import "time"

type User struct {
	InternalUserId        string
	PublicUserId          string
	FirebaseUserId        string
	Email                 string
	TermsAgreedDt         time.Time
	TermsVersion          string
	PrivacyPolicyAgreedDt time.Time
	PrivacyPolicyVersion  string
}
