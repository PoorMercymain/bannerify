// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/PoorMercymain/bannerify/internal/bannerify/domain (interfaces: AuthorizationRepository)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockAuthorizationRepository is a mock of AuthorizationRepository interface.
type MockAuthorizationRepository struct {
	ctrl     *gomock.Controller
	recorder *MockAuthorizationRepositoryMockRecorder
}

// MockAuthorizationRepositoryMockRecorder is the mock recorder for MockAuthorizationRepository.
type MockAuthorizationRepositoryMockRecorder struct {
	mock *MockAuthorizationRepository
}

// NewMockAuthorizationRepository creates a new mock instance.
func NewMockAuthorizationRepository(ctrl *gomock.Controller) *MockAuthorizationRepository {
	mock := &MockAuthorizationRepository{ctrl: ctrl}
	mock.recorder = &MockAuthorizationRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuthorizationRepository) EXPECT() *MockAuthorizationRepositoryMockRecorder {
	return m.recorder
}

// GetPasswordHash mocks base method.
func (m *MockAuthorizationRepository) GetPasswordHash(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPasswordHash", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPasswordHash indicates an expected call of GetPasswordHash.
func (mr *MockAuthorizationRepositoryMockRecorder) GetPasswordHash(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPasswordHash", reflect.TypeOf((*MockAuthorizationRepository)(nil).GetPasswordHash), arg0, arg1)
}

// IsAdmin mocks base method.
func (m *MockAuthorizationRepository) IsAdmin(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAdmin", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsAdmin indicates an expected call of IsAdmin.
func (mr *MockAuthorizationRepositoryMockRecorder) IsAdmin(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAdmin", reflect.TypeOf((*MockAuthorizationRepository)(nil).IsAdmin), arg0, arg1)
}

// Register mocks base method.
func (m *MockAuthorizationRepository) Register(arg0 context.Context, arg1, arg2 string, arg3 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register.
func (mr *MockAuthorizationRepositoryMockRecorder) Register(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockAuthorizationRepository)(nil).Register), arg0, arg1, arg2, arg3)
}
