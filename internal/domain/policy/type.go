package policy

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

// GetId 規約タイプのIDを返却する
func (t PolicyType) GetId() string {
	code, ok := typeToCodeMap[t]

	if !ok {
		panic("invalid policy type")
	}

	return code
}

// PolicyTypeFromId 規約タイプIDをもとに規約タイプを返却する
func PolicyTypeFromId(id string) PolicyType {
	policyType, ok := codeToTypeMap[id]

	if !ok {
		panic("invalid policy id")
	}

	return policyType
}

// IsValid 規約タイプが正しいものであればtrueを返却する
func (t PolicyType) IsValid() bool {
	_, ok := typeToCodeMap[t]
	return ok
}
