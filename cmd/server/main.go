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
	"strconv"
	"time"

	"github.com/chriskuehl/fluffy/server"
	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/config/loader"
	"github.com/chriskuehl/fluffy/server/logging"
)

var Version = "(dev)"

type cmdConfig struct {
	conf *config.Config

	printConfig bool
	configPath  string
}

func newConfigFromArgs(args []string) (*cmdConfig, error) {
	c := &cmdConfig{
		conf: server.NewConfig(),
	}
	fs := flag.NewFlagSet("fluffy", flag.ExitOnError)
	fs.BoolVar(&c.printConfig, "print-config", false, "print the config and exit")
	fs.StringVar(&c.configPath, "config", "", "path to config file")
	fs.StringVar(&c.conf.Host, "host", "localhost", "host to listen on")
	fs.UintVar(&c.conf.Port, "port", 8080, "port to listen on")
	fs.BoolVar(&c.conf.DevMode, "dev", false, "enable dev mode")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if c.configPath != "" {
		if err := loader.LoadConfigTOML(c.conf, c.configPath); err != nil {
			return nil, fmt.Errorf("loading config: %w", err)
		}
	}
	c.conf.Version = Version
	return c, nil
}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	conf, err := newConfigFromArgs(args)
	if err != nil {
		return fmt.Errorf("parsing args: %w", err)
	}

	if conf.printConfig {
		c, err := loader.DumpConfigTOML(conf.conf)
		if err != nil {
			return fmt.Errorf("dumping config: %w", err)
		}

		fmt.Println(c)
		return nil
	}

	logger := logging.NewSlogLogger(slog.New(slog.NewTextHandler(w, nil)))

	handler, err := server.NewServer(logger, conf.conf)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(conf.conf.Host, strconv.FormatUint(uint64(conf.conf.Port), 10)),
		Handler: handler,
	}
	go func() {
		logger.Info(ctx, "listening", "addr", httpServer.Addr)
		if conf.conf.DevMode {
			logger.Warn(ctx, "dev mode enabled! do not use in production!")
		}
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx := context.Background()
	shutdownCtx, cancel = context.WithTimeout(shutdownCtx, 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutting down: %w", err)
	}
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
