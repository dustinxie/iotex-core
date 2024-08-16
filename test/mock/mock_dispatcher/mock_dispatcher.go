// Code generated by MockGen. DO NOT EDIT.
// Source: ./dispatcher/dispatcher.go

// Package mock_dispatcher is a generated GoMock package.
package mock_dispatcher

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	dispatcher "github.com/iotexproject/iotex-core/dispatcher"
	peer "github.com/libp2p/go-libp2p/core/peer"
	proto "google.golang.org/protobuf/proto"
)

// MockDispatcher is a mock of Dispatcher interface.
type MockDispatcher struct {
	ctrl     *gomock.Controller
	recorder *MockDispatcherMockRecorder
}

// MockDispatcherMockRecorder is the mock recorder for MockDispatcher.
type MockDispatcherMockRecorder struct {
	mock *MockDispatcher
}

// NewMockDispatcher creates a new mock instance.
func NewMockDispatcher(ctrl *gomock.Controller) *MockDispatcher {
	mock := &MockDispatcher{ctrl: ctrl}
	mock.recorder = &MockDispatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDispatcher) EXPECT() *MockDispatcherMockRecorder {
	return m.recorder
}

// AddSubscriber mocks base method.
func (m *MockDispatcher) AddSubscriber(arg0 uint32, arg1 dispatcher.Subscriber) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddSubscriber", arg0, arg1)
}

// AddSubscriber indicates an expected call of AddSubscriber.
func (mr *MockDispatcherMockRecorder) AddSubscriber(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSubscriber", reflect.TypeOf((*MockDispatcher)(nil).AddSubscriber), arg0, arg1)
}

// HandleBroadcast mocks base method.
func (m *MockDispatcher) HandleBroadcast(arg0 context.Context, arg1 uint32, arg2 string, arg3 proto.Message) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandleBroadcast", arg0, arg1, arg2, arg3)
}

// HandleBroadcast indicates an expected call of HandleBroadcast.
func (mr *MockDispatcherMockRecorder) HandleBroadcast(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleBroadcast", reflect.TypeOf((*MockDispatcher)(nil).HandleBroadcast), arg0, arg1, arg2, arg3)
}

// HandleTell mocks base method.
func (m *MockDispatcher) HandleTell(arg0 context.Context, arg1 uint32, arg2 peer.AddrInfo, arg3 proto.Message) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandleTell", arg0, arg1, arg2, arg3)
}

// HandleTell indicates an expected call of HandleTell.
func (mr *MockDispatcherMockRecorder) HandleTell(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleTell", reflect.TypeOf((*MockDispatcher)(nil).HandleTell), arg0, arg1, arg2, arg3)
}

// Start mocks base method.
func (m *MockDispatcher) Start(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockDispatcherMockRecorder) Start(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockDispatcher)(nil).Start), arg0)
}

// Stop mocks base method.
func (m *MockDispatcher) Stop(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockDispatcherMockRecorder) Stop(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockDispatcher)(nil).Stop), arg0)
}
