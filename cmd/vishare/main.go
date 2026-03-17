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
	"github.com/victorfengming/vishare/internal/tray"
)

func main() {
	cfgPath := flag.String("config", "config.toml", "path to config file")
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

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

	statusCh := make(chan tray.StatusMsg, 8)

	switch cfg.Role {
	case config.RoleServer:
		srvStatusCh := make(chan server.StatusMsg, 8)
		srv := server.New(cfg, srvStatusCh)

		// Bridge server status → tray status
		go func() {
			for msg := range srvStatusCh {
				statusCh <- tray.StatusMsg{Connected: msg.Connected, ClientName: msg.ClientName}
			}
		}()

		go func() {
			if err := srv.Run(ctx); err != nil {
				log.Error().Err(err).Msg("server error")
				cancel()
			}
		}()

	case config.RoleClient:
		cliStatusCh := make(chan client.StatusMsg, 8)
		cli := client.New(cfg, cliStatusCh)

		// Bridge client status → tray status
		go func() {
			for msg := range cliStatusCh {
				statusCh <- tray.StatusMsg{Connected: msg.Connected}
			}
		}()

		go func() {
			if err := cli.Run(ctx); err != nil {
				log.Error().Err(err).Msg("client error")
				cancel()
			}
		}()
	}

	// systray.Run must be on the main goroutine
	tray.Run(statusCh, iconConnected, iconDisconnected, cancel)
}
