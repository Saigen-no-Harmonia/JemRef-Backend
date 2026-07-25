package mock

import "context"

type MockDB struct {
	PingErr error
}

func (m *MockDB) PingContext(c context.Context) error {
	return m.PingErr
}
