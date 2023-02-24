package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/http2"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		portlessHost, _, _ := strings.Cut(r.Host, ":")
		if portlessHost == "127.0.0.1" || portlessHost == "localhost" {
			return true
		}
		oUrl, err := url.Parse(r.Header.Get("origin"))
		if err != nil {
			return false
		}
		return portlessHost == oUrl.Hostname()
	},
}

type HTTPServer struct {
	server      *echo.Echo
	IdleChan    chan struct{}
	sessionsMgr *SessionManager
}

func StartHTTPServer(port string) (*HTTPServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, err
	}
	s := &HTTPServer{
		server:      echo.New(),
		IdleChan:    make(chan struct{}, 1),
		sessionsMgr: NewSessionManager(),
	}
	go s.sessionsMgr.Run()
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
	s.server.POST("/api/session", s.NewSession)
	s.server.Any("/api/session/:sessionID/ws", s.SessionWS)

	s.server.Any("/debug/pprof/:profile", func(c echo.Context) error {
		return echo.WrapHandler(pprof.Handler(c.Param(("profile"))))(c)
	})

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

func (s *HTTPServer) NewSession(c echo.Context) error {
	type (
		resp struct {
			Address string
		}
	)
	// fake it for now
	return c.JSON(http.StatusOK, resp{
		Address: withoutPort(c.Request().Host) + ":" + EnvPort + "/api/session/" + EnvPort + "/ws",
	})
}

func withoutPort(s string) string {
	host, _, _ := strings.Cut(s, ":")
	return host
}

func (s *HTTPServer) SessionWS(c echo.Context) error {
	sessionID := c.Param("sessionID")
	clientID := c.QueryParam("clientID")
	logger := logger.With().Str("clientID", clientID).Str("sessionID", sessionID).Logger()
	logger.Info().Msg("starting ws session")
	ws, err := wsUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()
	ws.SetReadLimit(100_000)
	session := s.sessionsMgr.Join(sessionID, clientID)
	doneChan := make(chan struct{}, 2)

	// TODO rewrite to differentiate between clientID and connectionID

	// read loop
	go func() {
		defer func() {
			close(session.ClientMsgs)
			doneChan <- struct{}{}
		}()
		for {
			logger.Debug().Msg("waiting for client message")
			msgType, msg, err := ws.ReadMessage()
			if err != nil {
				logger.Error().Err(err).Int("msgType", msgType).Msg("failed to read message")
				return
			}
			logger.Debug().Msg("got ws message")
			var cm ClientMessage
			err = json.Unmarshal(msg, &cm)
			if err != nil {
				logger.Error().Err(err).Int("msgType", msgType).Msg("failed to read message")
				return
			}
			cm.ClientID = clientID
			session.ClientMsgs <- cm
		}
	}()
	// write loop
	go func() {
		defer func() {
			doneChan <- struct{}{}
		}()
		for msg := range session.ServerMsgs {
			b, _ := json.Marshal(msg)
			err := ws.WriteMessage(websocket.TextMessage, b)
			if err != nil {
				logger.Error().Err(err).Msg("failed to write message")
				return
			}
		}
	}()

	<-doneChan
	logger.Info().Msg("closing ws session")
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
