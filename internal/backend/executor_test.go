package backend_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ju4n97/relic/internal/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRunner is a simple test double.
type mockRunner struct {
	runFunc   func(ctx context.Context, name string, args []string, stdin io.Reader) ([]byte, []byte, error)
	startFunc func(ctx context.Context, name string, args []string, stdin io.Reader) (io.ReadCloser, io.ReadCloser, func() error, error)
}

func (m *mockRunner) Run(ctx context.Context, name string, args []string, stdin io.Reader) (stdout, stderr []byte, err error) {
	return m.runFunc(ctx, name, args, stdin)
}

func (m *mockRunner) Start(ctx context.Context, name string, args []string, stdin io.Reader) (stdout, stderr io.ReadCloser, cancel func() error, err error) {
	return m.startFunc(ctx, name, args, stdin)
}

// nopCloser wraps a reader.
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestNewExecutor(t *testing.T) {
	t.Run("nonexistent binary", func(t *testing.T) {
		ex, err := backend.NewExecutor("/nonexistent/binary", time.Second)
		require.Error(t, err)
		assert.Nil(t, ex)
		assert.Contains(t, err.Error(), "binary not found")
	})
}

func TestExecute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		runner := &mockRunner{
			runFunc: func(_ context.Context, name string, args []string, _ io.Reader) ([]byte, []byte, error) {
				assert.Equal(t, "/bin/test", name)
				assert.Equal(t, []string{"arg1", "arg2"}, args)
				return []byte("output"), []byte(""), nil
			},
		}

		ex := backend.NewExecutorWithRunner("/bin/test", time.Second, runner)
		stdout, stderr, err := ex.Execute(context.Background(), []string{"arg1", "arg2"}, nil)

		require.NoError(t, err)
		assert.Equal(t, []byte("output"), stdout)
		assert.Equal(t, []byte(""), stderr)
	})

	t.Run("with stdin", func(t *testing.T) {
		runner := &mockRunner{
			runFunc: func(_ context.Context, _ string, _ []string, stdin io.Reader) ([]byte, []byte, error) {
				data, _ := io.ReadAll(stdin)
				assert.Equal(t, "input", string(data))
				return []byte("output"), []byte(""), nil
			},
		}

		ex := backend.NewExecutorWithRunner("/bin/test", time.Second, runner)
		stdout, _, err := ex.Execute(context.Background(), []string{}, strings.NewReader("input"))

		require.NoError(t, err)
		assert.Equal(t, []byte("output"), stdout)
	})

	t.Run("command fails", func(t *testing.T) {
		runner := &mockRunner{
			runFunc: func(_ context.Context, _ string, _ []string, _ io.Reader) ([]byte, []byte, error) {
				return []byte(""), []byte("error"), errors.New("exit status 1")
			},
		}

		ex := backend.NewExecutorWithRunner("/bin/test", time.Second, runner)
		stdout, stderr, err := ex.Execute(context.Background(), []string{}, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "exit status 1")
		assert.Equal(t, []byte(""), stdout)
		assert.Equal(t, []byte("error"), stderr)
	})

	t.Run("timeout", func(t *testing.T) {
		runner := &mockRunner{
			runFunc: func(ctx context.Context, _ string, _ []string, _ io.Reader) ([]byte, []byte, error) {
				<-ctx.Done()
				return nil, nil, ctx.Err()
			},
		}

		ex := backend.NewExecutorWithRunner("/bin/test", 10*time.Millisecond, runner)
		_, _, err := ex.Execute(context.Background(), []string{}, nil)

		require.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestStream(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		runner := &mockRunner{
			startFunc: func(_ context.Context, _ string, _ []string, _ io.Reader) (io.ReadCloser, io.ReadCloser, func() error, error) {
				stdout := nopCloser{strings.NewReader("line1\nline2\nline3\n")}
				stderr := nopCloser{strings.NewReader("")}
				wait := func() error { return nil }
				return stdout, stderr, wait, nil
			},
		}

		ex := backend.NewExecutorWithRunner("/bin/test", time.Second, runner)
		ch, err := ex.Stream(context.Background(), []string{}, nil)

		require.NoError(t, err)

		var chunks []backend.StreamChunk
		for chunk := range ch {
			chunks = append(chunks, chunk)
		}

		require.Len(t, chunks, 4)
		assert.Equal(t, []byte("line1\n"), chunks[0].Data)
		assert.Equal(t, []byte("line2\n"), chunks[1].Data)
		assert.Equal(t, []byte("line3\n"), chunks[2].Data)
		assert.True(t, chunks[3].Done)
		assert.NoError(t, chunks[3].Error)
	})

	t.Run("command fails", func(t *testing.T) {
		runner := &mockRunner{
			startFunc: func(_ context.Context, _ string, _ []string, _ io.Reader) (io.ReadCloser, io.ReadCloser, func() error, error) {
				stdout := nopCloser{strings.NewReader("output\n")}
				stderr := nopCloser{strings.NewReader("error message")}
				wait := func() error { return errors.New("exit status 1") }
				return stdout, stderr, wait, nil
			},
		}

		ex := backend.NewExecutorWithRunner("/bin/test", time.Second, runner)
		ch, err := ex.Stream(context.Background(), []string{}, nil)

		require.NoError(t, err)

		var chunks []backend.StreamChunk
		for chunk := range ch {
			chunks = append(chunks, chunk)
		}

		require.Len(t, chunks, 2)
		assert.Equal(t, []byte("output\n"), chunks[0].Data)
		assert.True(t, chunks[1].Done)
		assert.Error(t, chunks[1].Error)
		assert.Contains(t, chunks[1].Error.Error(), "exit status 1")
		assert.Contains(t, chunks[1].Error.Error(), "error message")
	})

	t.Run("start fails", func(t *testing.T) {
		runner := &mockRunner{
			startFunc: func(_ context.Context, _ string, _ []string, _ io.Reader) (io.ReadCloser, io.ReadCloser, func() error, error) {
				return nil, nil, nil, errors.New("cannot start")
			},
		}

		ex := backend.NewExecutorWithRunner("/bin/test", time.Second, runner)
		ch, err := ex.Stream(context.Background(), []string{}, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "executor: failed to start command")
		assert.Nil(t, ch)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		runner := &mockRunner{
			startFunc: func(_ context.Context, _ string, _ []string, _ io.Reader) (io.ReadCloser, io.ReadCloser, func() error, error) {
				// Simulate slow output
				r := &slowReader{
					data:   "line1\nline2\nline3\n",
					delay:  50 * time.Millisecond,
					cancel: cancel,
				}
				stdout := nopCloser{r}
				stderr := nopCloser{strings.NewReader("")}
				wait := func() error { return nil }
				return stdout, stderr, wait, nil
			},
		}

		ex := backend.NewExecutorWithRunner("/bin/test", 5*time.Second, runner)
		ch, err := ex.Stream(ctx, []string{}, nil)

		require.NoError(t, err)

		var chunks []backend.StreamChunk
		for chunk := range ch {
			chunks = append(chunks, chunk)
		}

		// Should get some chunks then cancellation
		require.NotEmpty(t, chunks)
		last := chunks[len(chunks)-1]
		assert.True(t, last.Done)
		if last.Error != nil {
			assert.ErrorIs(t, last.Error, context.Canceled)
		}
	})

	t.Run("empty output", func(t *testing.T) {
		runner := &mockRunner{
			startFunc: func(_ context.Context, _ string, _ []string, _ io.Reader) (io.ReadCloser, io.ReadCloser, func() error, error) {
				stdout := nopCloser{strings.NewReader("")}
				stderr := nopCloser{strings.NewReader("")}
				wait := func() error { return nil }
				return stdout, stderr, wait, nil
			},
		}

		ex := backend.NewExecutorWithRunner("/bin/test", time.Second, runner)
		ch, err := ex.Stream(context.Background(), []string{}, nil)

		require.NoError(t, err)

		var chunks []backend.StreamChunk
		for chunk := range ch {
			chunks = append(chunks, chunk)
		}

		require.Len(t, chunks, 1)
		assert.True(t, chunks[0].Done)
		assert.NoError(t, chunks[0].Error)
	})
}

// slowReader simulates slow I/O and cancels context after first read.
type slowReader struct {
	cancel context.CancelFunc
	data   string
	pos    int
	delay  time.Duration
}

func (s *slowReader) Read(p []byte) (int, error) {
	if s.pos == 0 && s.cancel != nil {
		// Cancel after first read starts
		go func() {
			time.Sleep(s.delay)
			s.cancel()
		}()
	}

	if s.pos >= len(s.data) {
		return 0, io.EOF
	}

	time.Sleep(s.delay / 10) // Small delay per read
	n := copy(p, s.data[s.pos:])
	s.pos += n
	return n, nil
}
