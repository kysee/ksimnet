package simnet

import (
	"errors"
	"github.com/ksimnet/netapp"
	"log"
	"net"
)

type Server struct {
	app netapp.ServerApp

	listenAddr *net.TCPAddr
	clients    map[string]*ServerConn
	listenCh   chan *Session
}

func NewServer(app netapp.ServerApp, laddr string) (*Server, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", laddr)
	if err != nil {
		return nil, err
	}
	return &Server{
		app:        app,
		listenAddr: tcpAddr,
		clients:    make(map[string]*ServerConn),
		listenCh:   make(chan *Session, 256), // backlog = 256
	}, nil
}

func (s *Server) Key() string {
	return s.listenAddr.String()
}

func (s *Server) Listen() (<-chan struct{}, error) {
	go func() {

		for {
			if sess, ok := <-s.listenCh; ok {
				log.Printf("Received %v\n", sess)
				s.accept(sess)

				sess.GetNetConn(0).SetRemoteConn(sess.GetNetConn(1))
				sess.evtCh <- struct{}{}
				log.Printf("Constructed %v\n", sess)
			} else {
				panic("error in receiving from listenCh")
				break
			}
		}
	}()

	AddServer(s)

	return make(<-chan struct{}), nil
}

func (s *Server) accept(sess *Session) error {
	c1 := sess.GetNetConn(0)
	if c1 == nil {
		return errors.New("a session has no NetConn")
	}

	serverConn, err := NewServerConn(s.app, s.listenAddr)
	if err != nil {
		return err
	}

	serverConn.SetRemoteConn(c1)
	serverConn.SetSession(sess)
	sess.SetNetConn(1, serverConn.NetConn)

	s.clients[serverConn.Key()] = serverConn

	return s.app.OnAccept(serverConn.NetConn)
}
