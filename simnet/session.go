package simnet

import (
	"github.com/kysee/ksimnet/types"
	"sync"
)

type HostType int

const (
	CLIENT HostType = 0 + iota
	SERVER
)

type Session struct {
	mtx sync.RWMutex

	RemoteRetCh chan error
	conns       []types.NetConn
	listener    *Listener
}

func NewSession(c1, c2 types.NetConn) *Session {
	return &Session{
		RemoteRetCh: make(chan error),
		conns:       []types.NetConn{c1, c2},
	}
}

func (sess *Session) SetNetConn(i HostType, c types.NetConn) {
	sess.mtx.Lock()
	defer sess.mtx.Unlock()

	sess.conns[i] = c
}

func (sess *Session) GetNetConn(i HostType) types.NetConn {
	sess.mtx.RLock()
	defer sess.mtx.RUnlock()

	return sess.conns[i]
}

func (sess *Session) Key() string {
	sess.mtx.RLock()
	defer sess.mtx.RUnlock()

	c1 := sess.conns[CLIENT]
	c2 := sess.conns[SERVER]
	if c1 == nil || c2 == nil {
		panic("")
	}
	return c1.LocalAddr().String() + "-" + c2.RemoteAddr().String()
}

func (sess *Session) String() string {
	sess.mtx.RLock()
	defer sess.mtx.RUnlock()

	r := ""
	for _, c := range sess.conns {
		if c == nil {
			continue
		}
		if r != "" {
			r += "<->"
		}
		r += c.LocalAddr().String()
	}
	return r
}
