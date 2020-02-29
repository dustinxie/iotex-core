// Code generated by MockGen. DO NOT EDIT.
// Source: ./blockchain/blockcreationsubscriber.go

// Package mock_blockcreationsubscriber is a generated GoMock package.
package mock_blockcreationsubscriber

import (
	gomock "github.com/golang/mock/gomock"
	block "github.com/iotexproject/iotex-core/blockchain/block"
	reflect "reflect"
)

// MockBlockCreationSubscriber is a mock of BlockCreationSubscriber interface
type MockBlockCreationSubscriber struct {
	ctrl     *gomock.Controller
	recorder *MockBlockCreationSubscriberMockRecorder
}

// MockBlockCreationSubscriberMockRecorder is the mock recorder for MockBlockCreationSubscriber
type MockBlockCreationSubscriberMockRecorder struct {
	mock *MockBlockCreationSubscriber
}

// NewMockBlockCreationSubscriber creates a new mock instance
func NewMockBlockCreationSubscriber(ctrl *gomock.Controller) *MockBlockCreationSubscriber {
	mock := &MockBlockCreationSubscriber{ctrl: ctrl}
	mock.recorder = &MockBlockCreationSubscriberMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBlockCreationSubscriber) EXPECT() *MockBlockCreationSubscriberMockRecorder {
	return m.recorder
}

// ReceiveBlock mocks base method
func (m *MockBlockCreationSubscriber) ReceiveBlock(arg0 *block.Block) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReceiveBlock", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReceiveBlock indicates an expected call of ReceiveBlock
func (mr *MockBlockCreationSubscriberMockRecorder) ReceiveBlock(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReceiveBlock", reflect.TypeOf((*MockBlockCreationSubscriber)(nil).ReceiveBlock), arg0)
}