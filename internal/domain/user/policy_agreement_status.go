package user

type PolicyAgreementStatus string

const (
	PolicyAgreementStatusAgreed         = "agreed"
	PolicyAgreementStatusUpdateRequired = "update_required"
	PolicyAgreementStatusNotAgreed      = "not_agreed"
	PolicyAgreementStatusWithdrown      = "withdrown"
)

func (s PolicyAgreementStatus) IsValid() bool {
	switch s {
	case PolicyAgreementStatusAgreed,
		PolicyAgreementStatusUpdateRequired,
		PolicyAgreementStatusNotAgreed,
		PolicyAgreementStatusWithdrown:
		return true
	default:
		return false
	}
}
