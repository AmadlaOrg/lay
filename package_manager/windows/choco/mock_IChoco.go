// Code generated by mockery v2.43.2. DO NOT EDIT.

package choco

import mock "github.com/stretchr/testify/mock"

// MockPackageManagerWindowsChoco is an autogenerated mock type for the IChoco type
type MockPackageManagerWindowsChoco struct {
	mock.Mock
}

type MockPackageManagerWindowsChoco_Expecter struct {
	mock *mock.Mock
}

func (_m *MockPackageManagerWindowsChoco) EXPECT() *MockPackageManagerWindowsChoco_Expecter {
	return &MockPackageManagerWindowsChoco_Expecter{mock: &_m.Mock}
}

// NewMockPackageManagerWindowsChoco creates a new instance of MockPackageManagerWindowsChoco. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPackageManagerWindowsChoco(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPackageManagerWindowsChoco {
	mock := &MockPackageManagerWindowsChoco{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
