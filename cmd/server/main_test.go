package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

const port = 17491

func waitForReady(ctx context.Context, timeout time.Duration) error {
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

func TestIntegration(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		if err := run(ctx, &buf, []string{"-port", strconv.Itoa(port)}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		close(done)
	}()
	if err := waitForReady(ctx, 5*time.Second); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	cancel()
	<-done

	want := "req.method=GET req.path=/healthz req.content_length=0"
	if got := buf.String(); !strings.Contains(got, want) {
		t.Errorf("log output did not contain expected string: got %q, want %q", got, want)
	}
}
