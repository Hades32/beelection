package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"golang.org/x/net/http2"
)

var logger = zerolog.New(os.Stdout)

func main() {
	s, err := StartHTTPServer(GetEnv("PORT", "8080"))
	if err != nil {
		logger.Error().Err(err).Msg("error creating tcp listener")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Warn().Msg("received shutdown signal!")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.Shutdown(ctx)
}

type HTTPServer struct {
	server *echo.Echo
}

func StartHTTPServer(port string) (*HTTPServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, err
	}
	s := &HTTPServer{
		server: echo.New(),
	}
	s.server.HideBanner = true
	s.server.HidePort = true
	s.server.JSONSerializer = &NoEscapeJSONSerializer{}
	logConfig := middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/hc"
		},
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out},"proto":"${protocol}","userID":"${header:loguserid}","reqID":"${header:reqID}"}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
		Output:           os.Stdout, // logger or os.Stdout
	}
	// s.server.Use(CreateReqContext)
	// s.server.Validator = &CustomValidator{validator: validator.New()}
	s.server.Use(middleware.LoggerWithConfig(logConfig))
	s.server.Use(middleware.CORS())

	s.server.Static("", "../frontend/build")
	s.server.GET("/hc", s.HealthCheck)

	s.server.Listener = listener
	go func() {
		logger.Info().Msg("starting h2c server on " + listener.Addr().String())
		err := s.server.StartH2CServer("", &http2.Server{})
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error().Err(err).Msg("failed to start h2c server, exiting")
			os.Exit(1)
		}
	}()

	return s, nil
}

func (s *HTTPServer) HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
