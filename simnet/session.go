package simnet

import (
	"github.com/ksimnet/netconn"
	"sync"
)

type Session struct {
	mtx sync.Mutex

	evtCh chan struct{}
	conns []*netconn.NetConn
}

func NewSession(c1, c2 *netconn.NetConn) *Session {
	return &Session{
		evtCh: make(chan struct{}),
		conns: []*netconn.NetConn{c1, c2},
	}
}

func (sess *Session) SetNetConn(i int, c *netconn.NetConn) {
	sess.conns[i] = c
}

func (sess *Session) GetNetConn(i int) *netconn.NetConn {
	return sess.conns[i]
}

func (sess *Session) String() string {
	r := ""
	for _, c := range sess.conns {
		if c == nil {
			continue
		}
		if r != "" {
			r += " <--> "
		}
		r += c.LocalAddr().String()
	}
	return r
}
