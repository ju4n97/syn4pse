package backend

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock types ---

type MockBackend struct {
	mock.Mock
}

func (m *MockBackend) Provider() string {
	args := m.Called()
	return string(args.String(0))
}

func (m *MockBackend) Infer(ctx context.Context, req *Request) (*Response, error) {
	args := m.Called(ctx, req)
	if resp, ok := args.Get(0).(*Response); ok {
		return resp, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockBackend) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockStreamingBackend struct {
	MockBackend
}

func (m *MockStreamingBackend) InferStream(ctx context.Context, req *Request) (<-chan StreamChunk, error) {
	args := m.Called(ctx, req)
	if ch, ok := args.Get(0).(<-chan StreamChunk); ok {
		return ch, args.Error(1)
	}
	return nil, args.Error(1)
}

// --- Tests ---

func TestRegistry_RegisterAndGet(t *testing.T) {
	reg := NewRegistry()
	mockBackend := new(MockBackend)
	mockBackend.On("Provider").Return("test-backend")

	_ = reg.Register(mockBackend)

	got, ok := reg.Get("test-backend")
	assert.True(t, ok)
	assert.Equal(t, mockBackend, got)

	// Ensure a missing backend returns false
	_, ok = reg.Get("missing")
	assert.False(t, ok)

	mockBackend.AssertExpectations(t)
}

func TestRegistry_Close(t *testing.T) {
	reg := NewRegistry()

	b1 := new(MockBackend)
	b2 := new(MockBackend)
	b1.On("Provider").Return("b1")
	b2.On("Provider").Return("b2")

	// Normal close
	b1.On("Close").Return(nil).Once()
	b2.On("Close").Return(nil).Once()

	_ = reg.Register(b1)
	_ = reg.Register(b2)

	err := reg.Close()
	assert.NoError(t, err)

	b1.AssertExpectations(t)
	b2.AssertExpectations(t)
}

func TestRegistry_CloseErrorPropagation(t *testing.T) {
	reg := NewRegistry()

	b1 := new(MockBackend)
	b2 := new(MockBackend)

	b1.On("Provider").Return("b1")
	b2.On("Provider").Return("b2")

	b1.On("Close").Return(errors.New("close failed")).Once()
	b2.On("Close").Return(nil).Maybe()

	_ = reg.Register(b1)
	_ = reg.Register(b2)

	err := reg.Close()
	assert.EqualError(t, err, "close failed")

	b1.AssertExpectations(t)
	b2.AssertExpectations(t)
}
