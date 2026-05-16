package policy

import "fmt"

type PolicyType string

const (
	PolicyTypeTermsOfService PolicyType = "terms"
	PolicyTypePrivacyPolicy  PolicyType = "privacy_policy"
)

const (
	PolicyIdTermsOfService = "P001"
	PolicyIdPrivacyPolicy  = "P002"
)

var typeToCodeMap = map[PolicyType]string{
	PolicyTypeTermsOfService: PolicyIdTermsOfService,
	PolicyTypePrivacyPolicy:  PolicyIdPrivacyPolicy,
}

var codeToTypeMap = map[string]PolicyType{
	PolicyIdTermsOfService: PolicyTypeTermsOfService,
	PolicyIdPrivacyPolicy:  PolicyTypePrivacyPolicy,
}

func (t PolicyType) GetId() (string, error) {
	code, ok := typeToCodeMap[t]

	if !ok {
		return "", fmt.Errorf("invalid policy type: %s", t)
	}

	return code, nil
}

func PolicyTypeFromCode(code string) (PolicyType, error) {
	policyType, ok := codeToTypeMap[code]

	if !ok {
		return "", fmt.Errorf("invalid policy code: %s", code)
	}

	return policyType, nil
}

func (t PolicyType) IsValid() bool {
	_, ok := typeToCodeMap[t]
	return ok
}
