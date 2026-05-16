package user

type PolicyAgreementStatus string

const (
	PolicyAgreementStatusAgreed         PolicyAgreementStatus = "agreed"
	PolicyAgreementStatusUpdateRequired PolicyAgreementStatus = "update_required"
)
