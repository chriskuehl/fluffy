package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/chriskuehl/fluffy/server"
	"github.com/chriskuehl/fluffy/server/logging"
)

type config struct {
	*server.Config

	host string
	port string
}

func newConfigFromArgs(args []string) (*config, error) {
	c := config{}
	fs := flag.NewFlagSet("fluffy", flag.ExitOnError)
	fs.StringVar(&c.host, "host", "localhost", "host to listen on")
	fs.StringVar(&c.port, "port", "8080", "port to listen on")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return &c, nil
}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config, err := newConfigFromArgs(args)
	if err != nil {
		return fmt.Errorf("parsing args: %w", err)
	}

	logger := logging.NewSlogLogger(slog.New(slog.NewTextHandler(w, nil)))

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.host, config.port),
		Handler: server.NewServer(logger, config.Config),
	}
	go func() {
		logger.Info(ctx, "listening", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()

	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
