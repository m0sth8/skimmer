package skimmer

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"github.com/codegangsta/martini-contrib/sessions"
)

// mock session.Session interface
type MockedSession struct {
	mock.Mock
}

func (m *MockedSession) Get(key interface{}) interface{} {
	args := m.Mock.Called(key)
	return args.Get(0)
}

func (m *MockedSession) Set(key, val interface{}) {
	m.Mock.Called(key, val)
	return
}

func (m *MockedSession) Delete(key interface{}) {
	m.Mock.Called(key)
	return
}

func (m *MockedSession) AddFlash(val interface{}, vars ...string) {
	m.Mock.Called(val, vars)
	return
}

func (m *MockedSession) Flashes(vars ...string) []interface{} {
	args := m.Mock.Called(vars)
	return args.Get(0).([]interface{})
}

func (m *MockedSession) Options(options sessions.Options) {
	m.Mock.Called(options)
	return
}

func TestHistory(t *testing.T) {
	session := new(MockedSession)
	data := []string{"one", "two", "three"}
	history := SessionHistory{session: session, size: 2, name:"test"}
	session.On("Get", "test").Return(data).Once()
	session.On("Set", "test", data[:2]).Return().Once()
	history.load()
	assert.Equal(t, history.All(), data)
	history.save()
	session.Mock.AssertExpectations(t)

	history = SessionHistory{session: session, size: 2, name:"test"}
	session.On("Get", "test").Return(nil).Once()
	history.load()
	assert.Empty(t, history.All())
	session.On("Set", "test", []string{"one"}).Return().Once()
	history.Add("one")
	session.On("Set", "test", []string{"two", "one"}).Return().Once()
	history.Add("two")
	session.On("Set", "test", []string{"three", "two"}).Return().Once()
	history.Add("three")
	session.Mock.AssertExpectations(t)



}
