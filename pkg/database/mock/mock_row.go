// Package mock provides mock implementations for database interfaces.
package mock

import (
	"reflect"

	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

// MockRow is a mock implementation of pgx.Row interface.
type MockRow struct {
	ctrl     *gomock.Controller
	recorder *MockRowMockRecorder
}

// MockRowMockRecorder is the mock recorder for MockRow.
type MockRowMockRecorder struct {
	mock *MockRow
}

// NewMockRow creates a new mock instance.
func NewMockRow(ctrl *gomock.Controller) *MockRow {
	mock := &MockRow{ctrl: ctrl}
	mock.recorder = &MockRowMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRow) EXPECT() *MockRowMockRecorder {
	return m.recorder
}

// Scan mocks base method.
func (m *MockRow) Scan(dest ...any) error {
	m.ctrl.T.Helper()
	varargs := make([]any, len(dest))
	for i, d := range dest {
		varargs[i] = d
	}
	ret := m.ctrl.Call(m, "Scan", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Scan indicates an expected call of Scan.
func (mr *MockRowMockRecorder) Scan(dest ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scan", reflect.TypeOf((*MockRow)(nil).Scan), dest...)
}

// Compile-time check to ensure MockRow implements pgx.Row
var _ pgx.Row = (*MockRow)(nil)

// ScanRowFunc is a helper function type for custom Scan behavior.
// Use this with DoAndReturn to populate scan destinations with test data.
type ScanRowFunc func(dest ...any) error
