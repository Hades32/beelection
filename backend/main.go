package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"beelection-backend/gologger"

	"github.com/rs/zerolog"
)

var logger = gologger.NewLogger()

func main() {
	logger := gologger.NewLogger()
	zerolog.DefaultContextLogger = &logger

	s, err := StartHTTPServer(EnvPort)
	if err != nil {
		logger.Error().Err(err).Msg("error creating tcp listener")
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	select {
	case <-c:
		logger.Warn().Msg("received shutdown signal!")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.Shutdown(ctx)
	case <-s.IdleChan:
		logger.Warn().Msg("idle - shutting down")
	}
}
