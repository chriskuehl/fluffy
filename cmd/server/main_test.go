package main

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chriskuehl/fluffy/testfunc"
)

const port = 14921

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
	if err := testfunc.WaitForReady(ctx, 5*time.Second, port); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cancel()
	<-done

	want := "req.method=GET req.path=/healthz req.content_length=0"
	if got := buf.String(); !strings.Contains(got, want) {
		t.Errorf("log output did not contain expected string: got %q, want %q", got, want)
	}
}
