// Code generated by MockGen. DO NOT EDIT.
// Source: ./state/factory/factory.go

// Package mock_factory is a generated GoMock package.
package mock_factory

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	address "github.com/iotexproject/iotex-address/address"
	action "github.com/iotexproject/iotex-core/v2/action"
	protocol "github.com/iotexproject/iotex-core/v2/action/protocol"
	actpool "github.com/iotexproject/iotex-core/v2/actpool"
	block "github.com/iotexproject/iotex-core/v2/blockchain/block"
	state "github.com/iotexproject/iotex-core/v2/state"
)

// MockFactory is a mock of Factory interface.
type MockFactory struct {
	ctrl     *gomock.Controller
	recorder *MockFactoryMockRecorder
}

// MockFactoryMockRecorder is the mock recorder for MockFactory.
type MockFactoryMockRecorder struct {
	mock *MockFactory
}

// NewMockFactory creates a new mock instance.
func NewMockFactory(ctrl *gomock.Controller) *MockFactory {
	mock := &MockFactory{ctrl: ctrl}
	mock.recorder = &MockFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFactory) EXPECT() *MockFactoryMockRecorder {
	return m.recorder
}

// DeleteTipBlock mocks base method.
func (m *MockFactory) DeleteTipBlock(arg0 context.Context, arg1 *block.Block) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteTipBlock", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteTipBlock indicates an expected call of DeleteTipBlock.
func (mr *MockFactoryMockRecorder) DeleteTipBlock(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteTipBlock", reflect.TypeOf((*MockFactory)(nil).DeleteTipBlock), arg0, arg1)
}

// Height mocks base method.
func (m *MockFactory) Height() (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Height")
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Height indicates an expected call of Height.
func (mr *MockFactoryMockRecorder) Height() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Height", reflect.TypeOf((*MockFactory)(nil).Height))
}

// NewBlockBuilder mocks base method.
func (m *MockFactory) NewBlockBuilder(arg0 context.Context, arg1 actpool.ActPool, arg2 func(action.Envelope) (*action.SealedEnvelope, error)) (*block.Builder, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewBlockBuilder", arg0, arg1, arg2)
	ret0, _ := ret[0].(*block.Builder)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewBlockBuilder indicates an expected call of NewBlockBuilder.
func (mr *MockFactoryMockRecorder) NewBlockBuilder(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewBlockBuilder", reflect.TypeOf((*MockFactory)(nil).NewBlockBuilder), arg0, arg1, arg2)
}

// PutBlock mocks base method.
func (m *MockFactory) PutBlock(arg0 context.Context, arg1 *block.Block) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutBlock", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutBlock indicates an expected call of PutBlock.
func (mr *MockFactoryMockRecorder) PutBlock(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutBlock", reflect.TypeOf((*MockFactory)(nil).PutBlock), arg0, arg1)
}

// ReadContractStorage mocks base method.
func (m *MockFactory) ReadContractStorage(arg0 context.Context, arg1 address.Address, arg2 []byte) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadContractStorage", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadContractStorage indicates an expected call of ReadContractStorage.
func (mr *MockFactoryMockRecorder) ReadContractStorage(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadContractStorage", reflect.TypeOf((*MockFactory)(nil).ReadContractStorage), arg0, arg1, arg2)
}

// ReadView mocks base method.
func (m *MockFactory) ReadView(arg0 string) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadView", arg0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadView indicates an expected call of ReadView.
func (mr *MockFactoryMockRecorder) ReadView(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadView", reflect.TypeOf((*MockFactory)(nil).ReadView), arg0)
}

// Register mocks base method.
func (m *MockFactory) Register(arg0 protocol.Protocol) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Register indicates an expected call of Register.
func (mr *MockFactoryMockRecorder) Register(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockFactory)(nil).Register), arg0)
}

// SimulateExecution mocks base method.
func (m *MockFactory) SimulateExecution(arg0 context.Context, arg1 address.Address, arg2 action.Envelope, arg3 ...protocol.SimulateOption) ([]byte, *action.Receipt, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SimulateExecution", varargs...)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(*action.Receipt)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// SimulateExecution indicates an expected call of SimulateExecution.
func (mr *MockFactoryMockRecorder) SimulateExecution(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SimulateExecution", reflect.TypeOf((*MockFactory)(nil).SimulateExecution), varargs...)
}

// Start mocks base method.
func (m *MockFactory) Start(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockFactoryMockRecorder) Start(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockFactory)(nil).Start), arg0)
}

// State mocks base method.
func (m *MockFactory) State(arg0 interface{}, arg1 ...protocol.StateOption) (uint64, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "State", varargs...)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// State indicates an expected call of State.
func (mr *MockFactoryMockRecorder) State(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "State", reflect.TypeOf((*MockFactory)(nil).State), varargs...)
}

// StateAtHeight mocks base method.
func (m *MockFactory) StateAtHeight(arg0 uint64, arg1 interface{}, arg2 ...protocol.StateOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StateAtHeight", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// StateAtHeight indicates an expected call of StateAtHeight.
func (mr *MockFactoryMockRecorder) StateAtHeight(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StateAtHeight", reflect.TypeOf((*MockFactory)(nil).StateAtHeight), varargs...)
}

// States mocks base method.
func (m *MockFactory) States(arg0 ...protocol.StateOption) (uint64, state.Iterator, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "States", varargs...)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(state.Iterator)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// States indicates an expected call of States.
func (mr *MockFactoryMockRecorder) States(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "States", reflect.TypeOf((*MockFactory)(nil).States), arg0...)
}

// StatesAtHeight mocks base method.
func (m *MockFactory) StatesAtHeight(arg0 uint64, arg1 ...protocol.StateOption) (state.Iterator, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StatesAtHeight", varargs...)
	ret0, _ := ret[0].(state.Iterator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StatesAtHeight indicates an expected call of StatesAtHeight.
func (mr *MockFactoryMockRecorder) StatesAtHeight(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StatesAtHeight", reflect.TypeOf((*MockFactory)(nil).StatesAtHeight), varargs...)
}

// Stop mocks base method.
func (m *MockFactory) Stop(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockFactoryMockRecorder) Stop(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockFactory)(nil).Stop), arg0)
}

// Validate mocks base method.
func (m *MockFactory) Validate(arg0 context.Context, arg1 *block.Block) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validate", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Validate indicates an expected call of Validate.
func (mr *MockFactoryMockRecorder) Validate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockFactory)(nil).Validate), arg0, arg1)
}
