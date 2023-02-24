package main

import (
	"time"
)

type (
	SessionManager struct {
		stopChan            chan struct{}
		joinChan            chan JoinReq
		sessionShutdownChan chan string
		sessions            map[string]*session
	}
	JoinReq struct {
		SessionID string
		ClientID  string
		RespChan  chan JoinResp
	}
	JoinResp struct {
		ServerMsgs <-chan any
		ClientMsgs chan<- ClientMessage
	}
	ClientMessage struct {
		ClientID string `json:"-"`
		State    int
	}
	session struct {
		id             string
		stopChan       chan struct{}
		clientMsgsChan chan ClientMessage
		joinChan       chan JoinReq
		leaveChan      chan string
		clients        map[string]*client
		mgr            *SessionManager
	}
	client struct {
		ClientID   string
		ServerMsgs chan<- any
		ClientMsgs <-chan ClientMessage
		State      int
	}
)

func NewSessionManager() *SessionManager {
	return &SessionManager{
		stopChan:            make(chan struct{}, 1),
		joinChan:            make(chan JoinReq),
		sessionShutdownChan: make(chan string, 1),
		sessions:            map[string]*session{},
	}
}

func (s *session) Join(clientID string) JoinResp {
	rc := make(chan JoinResp)
	s.joinChan <- JoinReq{
		ClientID: clientID,
		RespChan: rc,
	}
	r := <-rc
	return r
}

func (s *session) Run() {
	logger := logger.With().Str("sessionID", s.id).Logger()
	logger.Info().Msg("starting session")
	timer := time.NewTicker(250 * time.Millisecond)
	defer timer.Stop()
	idleTimer := time.NewTicker(2 * time.Minute)
	defer idleTimer.Stop()
	for {
		select {

		case cid := <-s.leaveChan:
			client, ok := s.clients[cid]
			if !ok {
				logger.Warn().Str("clientID", cid).Msg("client double left")
				continue
			}
			close(client.ServerMsgs)
			delete(s.clients, cid)
			if len(s.clients) == 0 {
				idleTimer.Reset(10 * time.Second)
			}
			// tell other clients?
			logger.Info().Str("clientID", cid).Msg("client left")

		case <-s.stopChan:
			logger.Warn().Msg("session was stopped")
			s.shutdown()
			return
		case <-idleTimer.C:
			logger.Warn().Msg("session idle - shutting down")
			s.shutdown()
			return

		case msg := <-s.joinChan:
			idleTimer.Stop()
			sChan := make(chan any, 1)
			cChan := make(chan ClientMessage)
			s.clients[msg.ClientID] = &client{
				ClientID:   msg.ClientID,
				ServerMsgs: sChan,
				ClientMsgs: cChan,
			}
			msg.RespChan <- JoinResp{
				ServerMsgs: sChan,
				ClientMsgs: cChan,
			}
			go func(clientID string) {
				for cm := range cChan {
					s.clientMsgsChan <- cm
				}
				s.leaveChan <- clientID
			}(msg.ClientID)
			// tell other clients?
			logger.Info().Str("clientID", msg.ClientID).Msg("client joined")

		case msg := <-s.clientMsgsChan:
			logger.Info().Msg("client message")
			s.handleMsg(msg)

		case <-timer.C:
			// do logic
			sum := 0
			for _, c := range s.clients {
				sum += c.State
			}
			for _, c := range s.clients {
				select {
				case c.ServerMsgs <- map[string]any{
					"test": sum,
					"cnt":  len(s.clients),
				}:
				default:
				}
			}
		}
	}
}

func (s *session) handleMsg(msg ClientMessage) {
	s.clients[msg.ClientID].State = msg.State
}

func (s *session) shutdown() {
	s.mgr.sessionShutdownChan <- s.id
	for _, c := range s.clients {
		close(c.ServerMsgs)
	}
}

func (sm *SessionManager) shutdown() {
}

func (sm *SessionManager) Run() {
	defer sm.shutdown()
	for {
		select {
		case <-sm.stopChan:
			for _, s := range sm.sessions {
				s.stopChan <- struct{}{}
			}
			return
		case sid := <-sm.sessionShutdownChan:
			delete(sm.sessions, sid)
		case msg := <-sm.joinChan:
			s := sm.sessions[msg.SessionID]
			if s == nil {
				s = &session{
					id:             msg.SessionID,
					stopChan:       make(chan struct{}, 1),
					clientMsgsChan: make(chan ClientMessage, 10),
					joinChan:       make(chan JoinReq),
					leaveChan:      make(chan string),
					clients:        map[string]*client{},
					mgr:            sm,
				}
				go s.Run()
				sm.sessions[msg.SessionID] = s
			}
			msg.RespChan <- s.Join(msg.ClientID)
		}
	}
}

func (sm *SessionManager) Join(id string, clientID string) JoinResp {
	respChan := make(chan JoinResp)
	sm.joinChan <- JoinReq{SessionID: id, ClientID: clientID, RespChan: respChan}
	return <-respChan
}
