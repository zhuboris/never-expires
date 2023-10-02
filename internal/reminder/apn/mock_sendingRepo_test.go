// Code generated by mockery v2.32.0. DO NOT EDIT.

package apn

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MocksendingRepo is an autogenerated mock type for the sendingRepo type
type MocksendingRepo struct {
	mock.Mock
}

type MocksendingRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *MocksendingRepo) EXPECT() *MocksendingRepo_Expecter {
	return &MocksendingRepo_Expecter{mock: &_m.Mock}
}

// notifications provides a mock function with given fields: ctx, dataCh
func (_m *MocksendingRepo) notifications(ctx context.Context, dataCh chan<- notificationData) error {
	ret := _m.Called(ctx, dataCh)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, chan<- notificationData) error); ok {
		r0 = rf(ctx, dataCh)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MocksendingRepo_notifications_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'notifications'
type MocksendingRepo_notifications_Call struct {
	*mock.Call
}

// notifications is a helper method to define mock.On call
//   - ctx context.Context
//   - dataCh chan<- notificationData
func (_e *MocksendingRepo_Expecter) notifications(ctx interface{}, dataCh interface{}) *MocksendingRepo_notifications_Call {
	return &MocksendingRepo_notifications_Call{Call: _e.mock.On("notifications", ctx, dataCh)}
}

func (_c *MocksendingRepo_notifications_Call) Run(run func(ctx context.Context, dataCh chan<- notificationData)) *MocksendingRepo_notifications_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(chan<- notificationData))
	})
	return _c
}

func (_c *MocksendingRepo_notifications_Call) Return(_a0 error) *MocksendingRepo_notifications_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MocksendingRepo_notifications_Call) RunAndReturn(run func(context.Context, chan<- notificationData) error) *MocksendingRepo_notifications_Call {
	_c.Call.Return(run)
	return _c
}

// NewMocksendingRepo creates a new instance of MocksendingRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMocksendingRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *MocksendingRepo {
	mock := &MocksendingRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
