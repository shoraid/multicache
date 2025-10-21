package testutil

import (
	"reflect"
	"sync"
	"testing"
)

type MockHelper struct {
	t         *testing.T
	mu        sync.Mutex
	expectMap map[string][]mockCall
	calls     []mockCall
}

type mockCall struct {
	Method string
	Args   []any
	Return []any
}

// NewMockHelper creates a new lightweight mock helper.
func NewMockHelper(t *testing.T) *MockHelper {
	return &MockHelper{
		t:         t,
		expectMap: make(map[string][]mockCall),
	}
}

// On registers an expected call and its return values.
func (m *MockHelper) On(method string, args ...any) *mockCallBuilder {
	m.mu.Lock()
	defer m.mu.Unlock()

	call := mockCall{Method: method, Args: args}
	m.expectMap[method] = append(m.expectMap[method], call)
	return &mockCallBuilder{m: m, call: &m.expectMap[method][len(m.expectMap[method])-1]}
}

type mockCallBuilder struct {
	m    *MockHelper
	call *mockCall
}

func (b *mockCallBuilder) Return(values ...any) {
	b.call.Return = values
}

// Called records a method call and returns its mocked result if exists.
func (m *MockHelper) Called(method string, args ...any) []any {
	m.mu.Lock()
	defer m.mu.Unlock()

	call := mockCall{Method: method, Args: args}
	m.calls = append(m.calls, call)

	// Find matching expected call
	if exps, ok := m.expectMap[method]; ok {
		for _, exp := range exps {
			if reflect.DeepEqual(exp.Args, args) {
				return exp.Return
			}
		}
	}

	m.t.Fatalf("unexpected call to %s with args %+v", method, args)
	return nil
}

// AssertCalled checks if method was called at least once with given args.
func (m *MockHelper) AssertCalled(t *testing.T, method string, args ...any) {
	t.Helper()

	for _, c := range m.calls {
		if c.Method == method && reflect.DeepEqual(c.Args, args) {
			return
		}
	}

	t.Fatalf("expected %s to be called with %+v, but it was not", method, args)
}

// AssertNotCalled checks that the given method was never called,
// optionally with specific arguments.
func (m *MockHelper) AssertNotCalled(t *testing.T, method string, args ...any) {
	t.Helper()

	for _, c := range m.calls {
		if c.Method == method {
			// If specific args are provided, check them
			if len(args) > 0 && reflect.DeepEqual(c.Args, args) {
				t.Fatalf("expected %s not to be called with %+v, but it was", method, args)
			}

			// If no args provided, fail on any call to the method
			if len(args) == 0 {
				t.Fatalf("expected %s not to be called, but it was", method)
			}
		}
	}
}

// AssertCalledCount checks how many times a method was called (optionally with specific arguments).
func (m *MockHelper) AssertCalledCount(t *testing.T, method string, expectedCount int, args ...any) {
	t.Helper()

	count := 0
	for _, c := range m.calls {
		if c.Method != method {
			continue
		}
		if len(args) > 0 && !reflect.DeepEqual(c.Args, args) {
			continue
		}
		count++
	}

	if count != expectedCount {
		if len(args) > 0 {
			t.Fatalf("expected %s to be called %d time(s) with %+v, but was called %d time(s)", method, expectedCount, args, count)
		} else {
			t.Fatalf("expected %s to be called %d time(s), but was called %d time(s)", method, expectedCount, count)
		}
	}
}

// AssertExpectations ensures all expected calls were made.
func (m *MockHelper) AssertExpectations(t *testing.T) {
	t.Helper()

	for method, exps := range m.expectMap {
		for _, exp := range exps {
			found := false
			for _, c := range m.calls {
				if c.Method == method && reflect.DeepEqual(exp.Args, c.Args) {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected call %s(%+v) was not made", method, exp.Args)
			}
		}
	}
}
