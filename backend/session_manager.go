package main

import (
	"time"
)

type (
	SessionManager struct {
		stopChan chan struct{}
		joinChan chan JoinReq
		sessions map[string]*session
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
		stopChan      chan struct{}
		clientMsgChan chan ClientMessage
		joinChan      chan JoinReq
		leaveChan     chan string
		clients       map[string]*client
		mgr           *SessionManager
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
		stopChan: make(chan struct{}, 1),
		joinChan: make(chan JoinReq),
		sessions: map[string]*session{},
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
	timer := time.NewTicker(250 * time.Millisecond)
	defer timer.Stop()
	idleTimer := time.NewTicker(10 * time.Second)
	defer idleTimer.Stop()
	for {
		select {

		case cid := <-s.leaveChan:
			delete(s.clients, cid)
			if len(s.clients) == 0 {
				idleTimer.Reset(10 * time.Second)
			}
			// tell other clients?

		case <-s.stopChan:
			s.shutdown()
			return
		case <-idleTimer.C:
			s.shutdown()
			return

		case msg := <-s.joinChan:
			idleTimer.Stop()
			sChan := make(chan any)
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
					s.clientMsgChan <- cm
				}
				s.leaveChan <- clientID
			}(msg.ClientID)
			// tell other clients?

		case msg := <-s.clientMsgChan:
			s.handleMsg(msg)

		case <-timer.C:
			// do logic
			sum := 0
			for _, c := range s.clients {
				sum += c.State
			}
			for _, c := range s.clients {
				c.ServerMsgs <- map[string]any{
					"test": sum,
				}
			}
		}
	}
}

func (s *session) handleMsg(msg ClientMessage) {
	s.clients[msg.ClientID].State = msg.State
}

func (s *session) shutdown() {
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
		case msg := <-sm.joinChan:
			s := sm.sessions[msg.SessionID]
			if s == nil {
				s = &session{
					stopChan:  make(chan struct{}, 1),
					joinChan:  make(chan JoinReq),
					leaveChan: make(chan string),
					clients:   map[string]*client{},
					mgr:       sm,
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
