package simnet

import (
	"errors"
	"github.com/ksimnet/netconn"
	"net"
)

type Server struct {
	worker netconn.ServerWorker
	//clients map[string]*netconn.NetPoint

	listenAddr *net.TCPAddr
	listenCh   chan *netconn.Session
}

func NewServer(app netconn.ServerWorker, laddr string) (*Server, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", laddr)
	if err != nil {
		return nil, err
	}
	return &Server{
		worker: app,
		//clients:    make(map[string]*netconn.NetPoint),
		listenAddr: tcpAddr,
		listenCh:   make(chan *netconn.Session, 256), // backlog = 256
	}, nil
}

func (s *Server) Key() string {
	return s.listenAddr.String()
}

func (s *Server) Listen() (chan<- struct{}, error) {
	done := make(chan struct{})

	go func() {
	Loop:
		for {
			select {
			case sess, ok := <-s.listenCh:
				if ok {
					s.accept(sess)

					sess.EvtCh <- struct{}{}

					//log.Printf("Constructed session: %v\n", sess)
				} else {
					panic("error in receiving from listenCh")
					break Loop
				}
			case <-done:
				break Loop
			}
		}
	}()

	AddServer(s)

	return done, nil
}

func (s *Server) accept(sess *netconn.Session) error {
	c1 := sess.GetNetPoint(0)
	if c1 == nil {
		return errors.New("a session has no NetConn")
	}

	c2 := netconn.NewNetPoint(s.worker, s.listenAddr)
	sess.SetNetPoint(1, c2)
	c2.SetSession(sess)
	c2.SetRemotePoint(c1)

	//s.clients[c2.RemoteAddr().String()] = c2

	return s.worker.OnAccept(c2)
}
