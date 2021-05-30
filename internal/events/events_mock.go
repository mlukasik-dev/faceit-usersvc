package events

import "github.com/stretchr/testify/mock"

type Mock struct {
	mock.Mock
}

var _ Client = (*Mock)(nil)

func (m *Mock) Publish(eventName string, data interface{}) {
	if eventName == CreateUserEvent {
		m.Called(eventName, "<id>")
		return
	}
	m.Called(eventName, data)
}
