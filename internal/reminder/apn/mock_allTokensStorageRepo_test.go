// Code generated by mockery v2.32.0. DO NOT EDIT.

package apn

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockallTokensStorageRepo is an autogenerated mock type for the allTokensStorageRepo type
type MockallTokensStorageRepo struct {
	mock.Mock
}

type MockallTokensStorageRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *MockallTokensStorageRepo) EXPECT() *MockallTokensStorageRepo_Expecter {
	return &MockallTokensStorageRepo_Expecter{mock: &_m.Mock}
}

// removeDeviceTokens provides a mock function with given fields: ctx, tokens
func (_m *MockallTokensStorageRepo) removeDeviceTokens(ctx context.Context, tokens []string) error {
	ret := _m.Called(ctx, tokens)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) error); ok {
		r0 = rf(ctx, tokens)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockallTokensStorageRepo_removeDeviceTokens_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'removeDeviceTokens'
type MockallTokensStorageRepo_removeDeviceTokens_Call struct {
	*mock.Call
}

// removeDeviceTokens is a helper method to define mock.On call
//   - ctx context.Context
//   - tokens []string
func (_e *MockallTokensStorageRepo_Expecter) removeDeviceTokens(ctx interface{}, tokens interface{}) *MockallTokensStorageRepo_removeDeviceTokens_Call {
	return &MockallTokensStorageRepo_removeDeviceTokens_Call{Call: _e.mock.On("removeDeviceTokens", ctx, tokens)}
}

func (_c *MockallTokensStorageRepo_removeDeviceTokens_Call) Run(run func(ctx context.Context, tokens []string)) *MockallTokensStorageRepo_removeDeviceTokens_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]string))
	})
	return _c
}

func (_c *MockallTokensStorageRepo_removeDeviceTokens_Call) Return(_a0 error) *MockallTokensStorageRepo_removeDeviceTokens_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockallTokensStorageRepo_removeDeviceTokens_Call) RunAndReturn(run func(context.Context, []string) error) *MockallTokensStorageRepo_removeDeviceTokens_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockallTokensStorageRepo creates a new instance of MockallTokensStorageRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockallTokensStorageRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockallTokensStorageRepo {
	mock := &MockallTokensStorageRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
