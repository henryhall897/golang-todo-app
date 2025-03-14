// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package usersmock

import (
	"context"
	"github.com/google/uuid"
	"github.com/henryhall897/golang-todo-app/internal/users/domain"
	"sync"
)

// Ensure, that ServiceMock does implement domain.Service.
// If this is not the case, regenerate this file with moq.
var _ domain.Service = &ServiceMock{}

// ServiceMock is a mock implementation of domain.Service.
//
//	func TestSomethingThatUsesService(t *testing.T) {
//
//		// make and configure a mocked domain.Service
//		mockedService := &ServiceMock{
//			CreateUserFunc: func(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
//				panic("mock out the CreateUser method")
//			},
//			DeleteUserFunc: func(ctx context.Context, id uuid.UUID) error {
//				panic("mock out the DeleteUser method")
//			},
//			GetUserByEmailFunc: func(ctx context.Context, email string) (domain.User, error) {
//				panic("mock out the GetUserByEmail method")
//			},
//			GetUserByIDFunc: func(ctx context.Context, id uuid.UUID) (domain.User, error) {
//				panic("mock out the GetUserByID method")
//			},
//			GetUsersFunc: func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
//				panic("mock out the GetUsers method")
//			},
//			UpdateUserFunc: func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
//				panic("mock out the UpdateUser method")
//			},
//		}
//
//		// use mockedService in code that requires domain.Service
//		// and then make assertions.
//
//	}
type ServiceMock struct {
	// CreateUserFunc mocks the CreateUser method.
	CreateUserFunc func(ctx context.Context, params domain.CreateUserParams) (domain.User, error)

	// DeleteUserFunc mocks the DeleteUser method.
	DeleteUserFunc func(ctx context.Context, id uuid.UUID) error

	// GetUserByEmailFunc mocks the GetUserByEmail method.
	GetUserByEmailFunc func(ctx context.Context, email string) (domain.User, error)

	// GetUserByIDFunc mocks the GetUserByID method.
	GetUserByIDFunc func(ctx context.Context, id uuid.UUID) (domain.User, error)

	// GetUsersFunc mocks the GetUsers method.
	GetUsersFunc func(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error)

	// UpdateUserFunc mocks the UpdateUser method.
	UpdateUserFunc func(ctx context.Context, params domain.UpdateUserParams) (domain.User, error)

	// calls tracks calls to the methods.
	calls struct {
		// CreateUser holds details about calls to the CreateUser method.
		CreateUser []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params domain.CreateUserParams
		}
		// DeleteUser holds details about calls to the DeleteUser method.
		DeleteUser []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ID is the id argument value.
			ID uuid.UUID
		}
		// GetUserByEmail holds details about calls to the GetUserByEmail method.
		GetUserByEmail []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Email is the email argument value.
			Email string
		}
		// GetUserByID holds details about calls to the GetUserByID method.
		GetUserByID []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ID is the id argument value.
			ID uuid.UUID
		}
		// GetUsers holds details about calls to the GetUsers method.
		GetUsers []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params domain.GetUsersParams
		}
		// UpdateUser holds details about calls to the UpdateUser method.
		UpdateUser []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Params is the params argument value.
			Params domain.UpdateUserParams
		}
	}
	lockCreateUser     sync.RWMutex
	lockDeleteUser     sync.RWMutex
	lockGetUserByEmail sync.RWMutex
	lockGetUserByID    sync.RWMutex
	lockGetUsers       sync.RWMutex
	lockUpdateUser     sync.RWMutex
}

// CreateUser calls CreateUserFunc.
func (mock *ServiceMock) CreateUser(ctx context.Context, params domain.CreateUserParams) (domain.User, error) {
	if mock.CreateUserFunc == nil {
		panic("ServiceMock.CreateUserFunc: method is nil but Service.CreateUser was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params domain.CreateUserParams
	}{
		Ctx:    ctx,
		Params: params,
	}
	mock.lockCreateUser.Lock()
	mock.calls.CreateUser = append(mock.calls.CreateUser, callInfo)
	mock.lockCreateUser.Unlock()
	return mock.CreateUserFunc(ctx, params)
}

// CreateUserCalls gets all the calls that were made to CreateUser.
// Check the length with:
//
//	len(mockedService.CreateUserCalls())
func (mock *ServiceMock) CreateUserCalls() []struct {
	Ctx    context.Context
	Params domain.CreateUserParams
} {
	var calls []struct {
		Ctx    context.Context
		Params domain.CreateUserParams
	}
	mock.lockCreateUser.RLock()
	calls = mock.calls.CreateUser
	mock.lockCreateUser.RUnlock()
	return calls
}

// DeleteUser calls DeleteUserFunc.
func (mock *ServiceMock) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if mock.DeleteUserFunc == nil {
		panic("ServiceMock.DeleteUserFunc: method is nil but Service.DeleteUser was just called")
	}
	callInfo := struct {
		Ctx context.Context
		ID  uuid.UUID
	}{
		Ctx: ctx,
		ID:  id,
	}
	mock.lockDeleteUser.Lock()
	mock.calls.DeleteUser = append(mock.calls.DeleteUser, callInfo)
	mock.lockDeleteUser.Unlock()
	return mock.DeleteUserFunc(ctx, id)
}

// DeleteUserCalls gets all the calls that were made to DeleteUser.
// Check the length with:
//
//	len(mockedService.DeleteUserCalls())
func (mock *ServiceMock) DeleteUserCalls() []struct {
	Ctx context.Context
	ID  uuid.UUID
} {
	var calls []struct {
		Ctx context.Context
		ID  uuid.UUID
	}
	mock.lockDeleteUser.RLock()
	calls = mock.calls.DeleteUser
	mock.lockDeleteUser.RUnlock()
	return calls
}

// GetUserByEmail calls GetUserByEmailFunc.
func (mock *ServiceMock) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	if mock.GetUserByEmailFunc == nil {
		panic("ServiceMock.GetUserByEmailFunc: method is nil but Service.GetUserByEmail was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		Email string
	}{
		Ctx:   ctx,
		Email: email,
	}
	mock.lockGetUserByEmail.Lock()
	mock.calls.GetUserByEmail = append(mock.calls.GetUserByEmail, callInfo)
	mock.lockGetUserByEmail.Unlock()
	return mock.GetUserByEmailFunc(ctx, email)
}

// GetUserByEmailCalls gets all the calls that were made to GetUserByEmail.
// Check the length with:
//
//	len(mockedService.GetUserByEmailCalls())
func (mock *ServiceMock) GetUserByEmailCalls() []struct {
	Ctx   context.Context
	Email string
} {
	var calls []struct {
		Ctx   context.Context
		Email string
	}
	mock.lockGetUserByEmail.RLock()
	calls = mock.calls.GetUserByEmail
	mock.lockGetUserByEmail.RUnlock()
	return calls
}

// GetUserByID calls GetUserByIDFunc.
func (mock *ServiceMock) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	if mock.GetUserByIDFunc == nil {
		panic("ServiceMock.GetUserByIDFunc: method is nil but Service.GetUserByID was just called")
	}
	callInfo := struct {
		Ctx context.Context
		ID  uuid.UUID
	}{
		Ctx: ctx,
		ID:  id,
	}
	mock.lockGetUserByID.Lock()
	mock.calls.GetUserByID = append(mock.calls.GetUserByID, callInfo)
	mock.lockGetUserByID.Unlock()
	return mock.GetUserByIDFunc(ctx, id)
}

// GetUserByIDCalls gets all the calls that were made to GetUserByID.
// Check the length with:
//
//	len(mockedService.GetUserByIDCalls())
func (mock *ServiceMock) GetUserByIDCalls() []struct {
	Ctx context.Context
	ID  uuid.UUID
} {
	var calls []struct {
		Ctx context.Context
		ID  uuid.UUID
	}
	mock.lockGetUserByID.RLock()
	calls = mock.calls.GetUserByID
	mock.lockGetUserByID.RUnlock()
	return calls
}

// GetUsers calls GetUsersFunc.
func (mock *ServiceMock) GetUsers(ctx context.Context, params domain.GetUsersParams) ([]domain.User, error) {
	if mock.GetUsersFunc == nil {
		panic("ServiceMock.GetUsersFunc: method is nil but Service.GetUsers was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params domain.GetUsersParams
	}{
		Ctx:    ctx,
		Params: params,
	}
	mock.lockGetUsers.Lock()
	mock.calls.GetUsers = append(mock.calls.GetUsers, callInfo)
	mock.lockGetUsers.Unlock()
	return mock.GetUsersFunc(ctx, params)
}

// GetUsersCalls gets all the calls that were made to GetUsers.
// Check the length with:
//
//	len(mockedService.GetUsersCalls())
func (mock *ServiceMock) GetUsersCalls() []struct {
	Ctx    context.Context
	Params domain.GetUsersParams
} {
	var calls []struct {
		Ctx    context.Context
		Params domain.GetUsersParams
	}
	mock.lockGetUsers.RLock()
	calls = mock.calls.GetUsers
	mock.lockGetUsers.RUnlock()
	return calls
}

// UpdateUser calls UpdateUserFunc.
func (mock *ServiceMock) UpdateUser(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
	if mock.UpdateUserFunc == nil {
		panic("ServiceMock.UpdateUserFunc: method is nil but Service.UpdateUser was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Params domain.UpdateUserParams
	}{
		Ctx:    ctx,
		Params: params,
	}
	mock.lockUpdateUser.Lock()
	mock.calls.UpdateUser = append(mock.calls.UpdateUser, callInfo)
	mock.lockUpdateUser.Unlock()
	return mock.UpdateUserFunc(ctx, params)
}

// UpdateUserCalls gets all the calls that were made to UpdateUser.
// Check the length with:
//
//	len(mockedService.UpdateUserCalls())
func (mock *ServiceMock) UpdateUserCalls() []struct {
	Ctx    context.Context
	Params domain.UpdateUserParams
} {
	var calls []struct {
		Ctx    context.Context
		Params domain.UpdateUserParams
	}
	mock.lockUpdateUser.RLock()
	calls = mock.calls.UpdateUser
	mock.lockUpdateUser.RUnlock()
	return calls
}
