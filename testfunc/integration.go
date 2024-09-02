package testfunc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/chriskuehl/fluffy/server"
	"github.com/chriskuehl/fluffy/server/logging"
)

func NewConfig() *server.Config {
	c := server.NewConfig()
	c.Version = "(test)"
	return c
}

type TestServer struct {
	Cleanup func()
	Logs    *bytes.Buffer
	Port    int
}

func RunningServer(t *testing.T, config *server.Config) TestServer {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	var buf bytes.Buffer
	port, done, err := run(t, ctx, &buf, config)
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
		Logs: &buf,
		Port: port,
	}
}

type serverState struct {
	port int
	err  error
}

func run(t *testing.T, ctx context.Context, w io.Writer, config *server.Config) (int, chan struct{}, error) {
	done := make(chan struct{})
	logger := logging.NewSlogLogger(slog.New(slog.NewTextHandler(w, nil)))

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