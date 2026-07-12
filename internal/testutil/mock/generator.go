package mock

type MockIdGen struct {
	GenerateFunc func() string

	GenerateCalls int
}

func (m *MockIdGen) Generate() string {
	m.GenerateCalls++

	if m.GenerateFunc != nil {
		return m.GenerateFunc()
	}

	return ""
}
