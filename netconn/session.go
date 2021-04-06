package netconn

import (
	"sync"
)

type Session struct {
	mtx sync.Mutex

	EvtCh chan struct{}
	conns []*NetPoint
}

func NewSession(c1, c2 *NetPoint) *Session {
	return &Session{
		EvtCh: make(chan struct{}),
		conns: []*NetPoint{c1, c2},
	}
}

func (sess *Session) SetNetPoint(i int, c *NetPoint) {
	sess.conns[i] = c
}

func (sess *Session) GetNetPoint(i int) *NetPoint {
	return sess.conns[i]
}

func (sess *Session) Key() string {
	return sess.String()
}

func (sess *Session) String() string {
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
