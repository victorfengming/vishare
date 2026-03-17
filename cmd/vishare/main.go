package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/victorfengming/vishare/internal/client"
	"github.com/victorfengming/vishare/internal/config"
	"github.com/victorfengming/vishare/internal/server"
	"github.com/victorfengming/vishare/internal/singleinstance"
	"github.com/victorfengming/vishare/internal/status"
	"github.com/victorfengming/vishare/internal/tray"
)

func main() {
	cfgPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Prevent multiple instances from running simultaneously
	if err := singleinstance.Acquire("vishare"); err != nil {
		log.Fatal().Err(err).Msg("startup failed")
	}
	defer singleinstance.Release()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// OS signal cancellation
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		log.Info().Msg("shutting down")
		cancel()
	}()

	statusCh := make(chan status.Msg, 8)

	switch cfg.Role {
	case config.RoleServer:
		srv := server.New(cfg, statusCh)
		go func() {
			if err := srv.Run(ctx); err != nil {
				log.Error().Err(err).Msg("server error")
				cancel()
			}
			close(statusCh)
		}()

	case config.RoleClient:
		cli := client.New(cfg, statusCh)
		go func() {
			if err := cli.Run(ctx); err != nil {
				log.Error().Err(err).Msg("client error")
				cancel()
			}
			close(statusCh)
		}()
	}

	// Platform-specific setup (macOS: initialise NSApp before systray.Run)
	platformSetup()

	// systray.Run must be on the main goroutine
	tray.Run(statusCh, iconConnected, iconDisconnected, cancel)
}
