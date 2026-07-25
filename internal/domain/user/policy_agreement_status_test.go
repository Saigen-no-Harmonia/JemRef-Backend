package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolicyAgreementStatus_ChkAgreementStat(t *testing.T) {

	tests := []struct {
		name          string
		agreedVersion string
		latestVersion string
		expected      PolicyAgreementStatus
	}{
		{
			name:          "ChkAgreementStat_正常",
			agreedVersion: "1.0",
			latestVersion: "1.0",
			expected:      PolicyAgreementStatusAgreed,
		},
		{
			name:          "ChkAgreementStat_正常_同意バージョンがより古い",
			agreedVersion: "0.5",
			latestVersion: "1.0",
			expected:      PolicyAgreementStatusUpdateRequired,
		},
		{
			name:          "ChkAgreementStat_正常_同意バージョンがより新しい",
			agreedVersion: "1.5",
			latestVersion: "1.0",
			expected:      PolicyAgreementStatusUpdateRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ChkAgreementStat(tt.agreedVersion, tt.latestVersion)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
