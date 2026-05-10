package policy

type PolicyType string

const (
	PolicyTypeTermsOfService = "terms"
	PolicyTypePrivacyPolicy  = "privacy_policy"
)

func (t PolicyType) IsValid() bool {
	switch t {
	case PolicyTypeTermsOfService,
		PolicyTypePrivacyPolicy:
		return true
	default:
		return false
	}
}
