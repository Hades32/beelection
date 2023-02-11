package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/http2"
)

type HTTPServer struct {
	server   *echo.Echo
	IdleChan chan struct{}
}

func StartHTTPServer(port string) (*HTTPServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, err
	}
	s := &HTTPServer{
		server:   echo.New(),
		IdleChan: make(chan struct{}, 1),
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
	s.server.Use(idleDetector(s.IdleChan, EnvIdleTimeout))

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

func idleDetector(idleChan chan struct{}, idleTime time.Duration) func(next echo.HandlerFunc) echo.HandlerFunc {
	watchdog := make(chan struct{})
	go func() {
		for {
			select {
			case <-watchdog:
				continue
			case <-time.After(idleTime):
				select {
				case idleChan <- struct{}{}:
				default:
				}
			}
		}
	}()
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			select {
			case watchdog <- struct{}{}:
			default:
			}
			return next(c)
		}
	}
}

func (s *HTTPServer) HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
