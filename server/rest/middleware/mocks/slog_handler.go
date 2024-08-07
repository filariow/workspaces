// Code generated by MockGen. DO NOT EDIT.
// Source: interfaces_test.go
//
// Generated by this command:
//
//	mockgen -source=interfaces_test.go -destination=mocks/slog_handler.go -package=mocks -exclude_interfaces=FakeCRCache,FakeHTTPHandler
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	slog "log/slog"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockFakeSlogHandler is a mock of FakeSlogHandler interface.
type MockFakeSlogHandler struct {
	ctrl     *gomock.Controller
	recorder *MockFakeSlogHandlerMockRecorder
}

// MockFakeSlogHandlerMockRecorder is the mock recorder for MockFakeSlogHandler.
type MockFakeSlogHandlerMockRecorder struct {
	mock *MockFakeSlogHandler
}

// NewMockFakeSlogHandler creates a new mock instance.
func NewMockFakeSlogHandler(ctrl *gomock.Controller) *MockFakeSlogHandler {
	mock := &MockFakeSlogHandler{ctrl: ctrl}
	mock.recorder = &MockFakeSlogHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFakeSlogHandler) EXPECT() *MockFakeSlogHandlerMockRecorder {
	return m.recorder
}

// Enabled mocks base method.
func (m *MockFakeSlogHandler) Enabled(arg0 context.Context, arg1 slog.Level) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enabled", arg0, arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Enabled indicates an expected call of Enabled.
func (mr *MockFakeSlogHandlerMockRecorder) Enabled(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enabled", reflect.TypeOf((*MockFakeSlogHandler)(nil).Enabled), arg0, arg1)
}

// Handle mocks base method.
func (m *MockFakeSlogHandler) Handle(arg0 context.Context, arg1 slog.Record) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Handle", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Handle indicates an expected call of Handle.
func (mr *MockFakeSlogHandlerMockRecorder) Handle(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Handle", reflect.TypeOf((*MockFakeSlogHandler)(nil).Handle), arg0, arg1)
}

// WithAttrs mocks base method.
func (m *MockFakeSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithAttrs", attrs)
	ret0, _ := ret[0].(slog.Handler)
	return ret0
}

// WithAttrs indicates an expected call of WithAttrs.
func (mr *MockFakeSlogHandlerMockRecorder) WithAttrs(attrs any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithAttrs", reflect.TypeOf((*MockFakeSlogHandler)(nil).WithAttrs), attrs)
}

// WithGroup mocks base method.
func (m *MockFakeSlogHandler) WithGroup(name string) slog.Handler {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithGroup", name)
	ret0, _ := ret[0].(slog.Handler)
	return ret0
}

// WithGroup indicates an expected call of WithGroup.
func (mr *MockFakeSlogHandlerMockRecorder) WithGroup(name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithGroup", reflect.TypeOf((*MockFakeSlogHandler)(nil).WithGroup), name)
}
