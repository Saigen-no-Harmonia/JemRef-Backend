package user

type PolicyAgreementStatus string

const (
	PolicyAgreementStatusAgreed         PolicyAgreementStatus = "agreed"
	PolicyAgreementStatusUpdateRequired PolicyAgreementStatus = "update_required"
)

// ChkAgreementStat 同意した規約と最新の規約のバージョンを比較し、同意状況を判定する
// Ph0では最新版≠同意版であれば全て要更新として扱う。
func ChkAgreementStat(agreedVer string, latestVer string) PolicyAgreementStatus {
	var stat PolicyAgreementStatus
	if agreedVer == latestVer {
		stat = PolicyAgreementStatusAgreed
	} else {
		stat = PolicyAgreementStatusUpdateRequired
	}
	return stat
}
