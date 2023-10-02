// Code generated by mockery v2.32.0. DO NOT EDIT.

package storage

import (
	context "context"

	pgtype "github.com/jackc/pgx/v5/pgtype"
	mock "github.com/stretchr/testify/mock"
)

// Mockrepository is an autogenerated mock type for the repository type
type Mockrepository struct {
	mock.Mock
}

type Mockrepository_Expecter struct {
	mock *mock.Mock
}

func (_m *Mockrepository) EXPECT() *Mockrepository_Expecter {
	return &Mockrepository_Expecter{mock: &_m.Mock}
}

// Ping provides a mock function with given fields: ctx
func (_m *Mockrepository) Ping(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Mockrepository_Ping_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Ping'
type Mockrepository_Ping_Call struct {
	*mock.Call
}

// Ping is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Mockrepository_Expecter) Ping(ctx interface{}) *Mockrepository_Ping_Call {
	return &Mockrepository_Ping_Call{Call: _e.mock.On("Ping", ctx)}
}

func (_c *Mockrepository_Ping_Call) Run(run func(ctx context.Context)) *Mockrepository_Ping_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Mockrepository_Ping_Call) Return(_a0 error) *Mockrepository_Ping_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Mockrepository_Ping_Call) RunAndReturn(run func(context.Context) error) *Mockrepository_Ping_Call {
	_c.Call.Return(run)
	return _c
}

// add provides a mock function with given fields: ctx, toAdd, ownerID
func (_m *Mockrepository) add(ctx context.Context, toAdd Storage, ownerID pgtype.UUID) (bool, *Storage, error) {
	ret := _m.Called(ctx, toAdd, ownerID)

	var r0 bool
	var r1 *Storage
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, Storage, pgtype.UUID) (bool, *Storage, error)); ok {
		return rf(ctx, toAdd, ownerID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, Storage, pgtype.UUID) bool); ok {
		r0 = rf(ctx, toAdd, ownerID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, Storage, pgtype.UUID) *Storage); ok {
		r1 = rf(ctx, toAdd, ownerID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*Storage)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, Storage, pgtype.UUID) error); ok {
		r2 = rf(ctx, toAdd, ownerID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Mockrepository_add_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'add'
type Mockrepository_add_Call struct {
	*mock.Call
}

// add is a helper method to define mock.On call
//   - ctx context.Context
//   - toAdd Storage
//   - ownerID pgtype.UUID
func (_e *Mockrepository_Expecter) add(ctx interface{}, toAdd interface{}, ownerID interface{}) *Mockrepository_add_Call {
	return &Mockrepository_add_Call{Call: _e.mock.On("add", ctx, toAdd, ownerID)}
}

func (_c *Mockrepository_add_Call) Run(run func(ctx context.Context, toAdd Storage, ownerID pgtype.UUID)) *Mockrepository_add_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(Storage), args[2].(pgtype.UUID))
	})
	return _c
}

func (_c *Mockrepository_add_Call) Return(_a0 bool, _a1 *Storage, _a2 error) *Mockrepository_add_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *Mockrepository_add_Call) RunAndReturn(run func(context.Context, Storage, pgtype.UUID) (bool, *Storage, error)) *Mockrepository_add_Call {
	_c.Call.Return(run)
	return _c
}

// allByOwnerID provides a mock function with given fields: ctx, ownerID, defaultNames
func (_m *Mockrepository) allByOwnerID(ctx context.Context, ownerID pgtype.UUID, defaultNames [3]string) ([]*Storage, error) {
	ret := _m.Called(ctx, ownerID, defaultNames)

	var r0 []*Storage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, pgtype.UUID, [3]string) ([]*Storage, error)); ok {
		return rf(ctx, ownerID, defaultNames)
	}
	if rf, ok := ret.Get(0).(func(context.Context, pgtype.UUID, [3]string) []*Storage); ok {
		r0 = rf(ctx, ownerID, defaultNames)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*Storage)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, pgtype.UUID, [3]string) error); ok {
		r1 = rf(ctx, ownerID, defaultNames)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Mockrepository_allByOwnerID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'allByOwnerID'
type Mockrepository_allByOwnerID_Call struct {
	*mock.Call
}

// allByOwnerID is a helper method to define mock.On call
//   - ctx context.Context
//   - ownerID pgtype.UUID
//   - defaultNames [3]string
func (_e *Mockrepository_Expecter) allByOwnerID(ctx interface{}, ownerID interface{}, defaultNames interface{}) *Mockrepository_allByOwnerID_Call {
	return &Mockrepository_allByOwnerID_Call{Call: _e.mock.On("allByOwnerID", ctx, ownerID, defaultNames)}
}

func (_c *Mockrepository_allByOwnerID_Call) Run(run func(ctx context.Context, ownerID pgtype.UUID, defaultNames [3]string)) *Mockrepository_allByOwnerID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(pgtype.UUID), args[2].([3]string))
	})
	return _c
}

func (_c *Mockrepository_allByOwnerID_Call) Return(_a0 []*Storage, _a1 error) *Mockrepository_allByOwnerID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Mockrepository_allByOwnerID_Call) RunAndReturn(run func(context.Context, pgtype.UUID, [3]string) ([]*Storage, error)) *Mockrepository_allByOwnerID_Call {
	_c.Call.Return(run)
	return _c
}

// clear provides a mock function with given fields: ctx, storageID, ownerID
func (_m *Mockrepository) clear(ctx context.Context, storageID pgtype.UUID, ownerID pgtype.UUID) (bool, error) {
	ret := _m.Called(ctx, storageID, ownerID)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, pgtype.UUID, pgtype.UUID) (bool, error)); ok {
		return rf(ctx, storageID, ownerID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, pgtype.UUID, pgtype.UUID) bool); ok {
		r0 = rf(ctx, storageID, ownerID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, pgtype.UUID, pgtype.UUID) error); ok {
		r1 = rf(ctx, storageID, ownerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Mockrepository_clear_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'clear'
type Mockrepository_clear_Call struct {
	*mock.Call
}

// clear is a helper method to define mock.On call
//   - ctx context.Context
//   - storageID pgtype.UUID
//   - ownerID pgtype.UUID
func (_e *Mockrepository_Expecter) clear(ctx interface{}, storageID interface{}, ownerID interface{}) *Mockrepository_clear_Call {
	return &Mockrepository_clear_Call{Call: _e.mock.On("clear", ctx, storageID, ownerID)}
}

func (_c *Mockrepository_clear_Call) Run(run func(ctx context.Context, storageID pgtype.UUID, ownerID pgtype.UUID)) *Mockrepository_clear_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(pgtype.UUID), args[2].(pgtype.UUID))
	})
	return _c
}

func (_c *Mockrepository_clear_Call) Return(_a0 bool, _a1 error) *Mockrepository_clear_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Mockrepository_clear_Call) RunAndReturn(run func(context.Context, pgtype.UUID, pgtype.UUID) (bool, error)) *Mockrepository_clear_Call {
	_c.Call.Return(run)
	return _c
}

// delete provides a mock function with given fields: ctx, storageID, ownerID
func (_m *Mockrepository) delete(ctx context.Context, storageID pgtype.UUID, ownerID pgtype.UUID) (bool, error) {
	ret := _m.Called(ctx, storageID, ownerID)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, pgtype.UUID, pgtype.UUID) (bool, error)); ok {
		return rf(ctx, storageID, ownerID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, pgtype.UUID, pgtype.UUID) bool); ok {
		r0 = rf(ctx, storageID, ownerID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, pgtype.UUID, pgtype.UUID) error); ok {
		r1 = rf(ctx, storageID, ownerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Mockrepository_delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'delete'
type Mockrepository_delete_Call struct {
	*mock.Call
}

// delete is a helper method to define mock.On call
//   - ctx context.Context
//   - storageID pgtype.UUID
//   - ownerID pgtype.UUID
func (_e *Mockrepository_Expecter) delete(ctx interface{}, storageID interface{}, ownerID interface{}) *Mockrepository_delete_Call {
	return &Mockrepository_delete_Call{Call: _e.mock.On("delete", ctx, storageID, ownerID)}
}

func (_c *Mockrepository_delete_Call) Run(run func(ctx context.Context, storageID pgtype.UUID, ownerID pgtype.UUID)) *Mockrepository_delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(pgtype.UUID), args[2].(pgtype.UUID))
	})
	return _c
}

func (_c *Mockrepository_delete_Call) Return(_a0 bool, _a1 error) *Mockrepository_delete_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Mockrepository_delete_Call) RunAndReturn(run func(context.Context, pgtype.UUID, pgtype.UUID) (bool, error)) *Mockrepository_delete_Call {
	_c.Call.Return(run)
	return _c
}

// isForbiddenToDelete provides a mock function with given fields: ctx, storageID, ownerID
func (_m *Mockrepository) isForbiddenToDelete(ctx context.Context, storageID pgtype.UUID, ownerID pgtype.UUID) (bool, error) {
	ret := _m.Called(ctx, storageID, ownerID)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, pgtype.UUID, pgtype.UUID) (bool, error)); ok {
		return rf(ctx, storageID, ownerID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, pgtype.UUID, pgtype.UUID) bool); ok {
		r0 = rf(ctx, storageID, ownerID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, pgtype.UUID, pgtype.UUID) error); ok {
		r1 = rf(ctx, storageID, ownerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Mockrepository_isForbiddenToDelete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'isForbiddenToDelete'
type Mockrepository_isForbiddenToDelete_Call struct {
	*mock.Call
}

// isForbiddenToDelete is a helper method to define mock.On call
//   - ctx context.Context
//   - storageID pgtype.UUID
//   - ownerID pgtype.UUID
func (_e *Mockrepository_Expecter) isForbiddenToDelete(ctx interface{}, storageID interface{}, ownerID interface{}) *Mockrepository_isForbiddenToDelete_Call {
	return &Mockrepository_isForbiddenToDelete_Call{Call: _e.mock.On("isForbiddenToDelete", ctx, storageID, ownerID)}
}

func (_c *Mockrepository_isForbiddenToDelete_Call) Run(run func(ctx context.Context, storageID pgtype.UUID, ownerID pgtype.UUID)) *Mockrepository_isForbiddenToDelete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(pgtype.UUID), args[2].(pgtype.UUID))
	})
	return _c
}

func (_c *Mockrepository_isForbiddenToDelete_Call) Return(_a0 bool, _a1 error) *Mockrepository_isForbiddenToDelete_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Mockrepository_isForbiddenToDelete_Call) RunAndReturn(run func(context.Context, pgtype.UUID, pgtype.UUID) (bool, error)) *Mockrepository_isForbiddenToDelete_Call {
	_c.Call.Return(run)
	return _c
}

// update provides a mock function with given fields: ctx, updated, ownerID
func (_m *Mockrepository) update(ctx context.Context, updated Storage, ownerID pgtype.UUID) (bool, *Storage, error) {
	ret := _m.Called(ctx, updated, ownerID)

	var r0 bool
	var r1 *Storage
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, Storage, pgtype.UUID) (bool, *Storage, error)); ok {
		return rf(ctx, updated, ownerID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, Storage, pgtype.UUID) bool); ok {
		r0 = rf(ctx, updated, ownerID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, Storage, pgtype.UUID) *Storage); ok {
		r1 = rf(ctx, updated, ownerID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*Storage)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, Storage, pgtype.UUID) error); ok {
		r2 = rf(ctx, updated, ownerID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Mockrepository_update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'update'
type Mockrepository_update_Call struct {
	*mock.Call
}

// update is a helper method to define mock.On call
//   - ctx context.Context
//   - updated Storage
//   - ownerID pgtype.UUID
func (_e *Mockrepository_Expecter) update(ctx interface{}, updated interface{}, ownerID interface{}) *Mockrepository_update_Call {
	return &Mockrepository_update_Call{Call: _e.mock.On("update", ctx, updated, ownerID)}
}

func (_c *Mockrepository_update_Call) Run(run func(ctx context.Context, updated Storage, ownerID pgtype.UUID)) *Mockrepository_update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(Storage), args[2].(pgtype.UUID))
	})
	return _c
}

func (_c *Mockrepository_update_Call) Return(_a0 bool, _a1 *Storage, _a2 error) *Mockrepository_update_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *Mockrepository_update_Call) RunAndReturn(run func(context.Context, Storage, pgtype.UUID) (bool, *Storage, error)) *Mockrepository_update_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockrepository creates a new instance of Mockrepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockrepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *Mockrepository {
	mock := &Mockrepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}