package user

type PolicyAgreementStatus string

const (
	PolicyAgreementStatusAgreed         = "agreed"
	PolicyAgreementStatusUpdateRequired = "update_required"
)

func (s PolicyAgreementStatus) IsValid() bool {
	switch s {
	case PolicyAgreementStatusAgreed,
		PolicyAgreementStatusUpdateRequired:
		return true
	default:
		return false
	}
}
