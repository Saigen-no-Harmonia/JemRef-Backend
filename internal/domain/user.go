package domain

import "time"

// TODO 最低限のスケルトン
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
