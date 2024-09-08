package testfunc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/chriskuehl/fluffy/server"
	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/logging"
)

type ConfigOption func(*config.Config) error

func WithStorageBackend(backend config.StorageBackend) ConfigOption {
	return func(c *config.Config) error {
		c.StorageBackend = backend
		return nil
	}
}

func NewConfig(opt ...ConfigOption) *config.Config {
	c, err := server.NewConfig()
	if err != nil {
		panic(fmt.Sprintf("unexpected error: %v", err))
	}
	c.Version = "(test)"
	c.StorageBackend = NewMemoryStorageBackend()

	for _, o := range opt {
		if err := o(c); err != nil {
			panic(fmt.Sprintf("unexpected error: %v", err))
		}
	}

	return c
}

type TestServer struct {
	Cleanup func()
	Port    int
}

func RunningServer(t *testing.T, config *config.Config) TestServer {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	port, done, err := run(t, ctx, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := WaitForReady(ctx, 5*time.Second, port); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return TestServer{
		Cleanup: sync.OnceFunc(func() {
			cancel()
			<-done
		}),
		Port: port,
	}
}

type serverState struct {
	port int
	err  error
}

func run(t *testing.T, ctx context.Context, config *config.Config) (int, chan struct{}, error) {
	done := make(chan struct{})
	logger := logging.NewSlogLogger(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	handler, err := server.NewServer(logger, config)
	if err != nil {
		return 0, nil, fmt.Errorf("creating server: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, nil, fmt.Errorf("listening: %w", err)
	}
	httpServer := &http.Server{Handler: handler}
	go func() {
		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Errorf("listening and serving: %v", err)
		}
	}()
	go func() {
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			t.Errorf("shutting down http server: %v", err)
		}
		close(done)
	}()

	return listener.Addr().(*net.TCPAddr).Port, done, nil
}

func WaitForReady(ctx context.Context, timeout time.Duration, port int) error {
	client := http.Client{}
	startTime := time.Now()
	for {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("http://localhost:%d/healthz", port),
			nil,
		)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %s\n", err.Error())
			continue
		}
		if resp.StatusCode == http.StatusOK {
			fmt.Println("/healthz is ready!")
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) >= timeout {
				return fmt.Errorf("timeout reached while waiting for endpoint")
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}
