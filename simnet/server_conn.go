package simnet

import (
	"errors"
	"github.com/ksimnet/netapp"
	"github.com/ksimnet/netconn"
	"net"
)

type ServerConn struct {
	*netconn.NetConn

	app     netapp.ServerApp
	session *Session
}

var _ netconn.NetOps = (*ServerConn)(nil)

func NewServerConn(a netapp.ServerApp, hostAddr *net.TCPAddr) (*ServerConn, error) {
	netConn, err := netconn.NewNetConn(hostAddr)
	if err != nil {
		return nil, err
	}

	return &ServerConn{
		NetConn: netConn,
		app:     a,
	}, nil
}

func (s *ServerConn) Key() string {
	return s.RemoteAddr().String()
}

func (s *ServerConn) Write(d []byte) (int, error) {
	if d == nil || len(d) == 0 {
		return 0, errors.New("invalid buffer")
	}
	return s.RemoteConn().RX_Put(d)
}

func (s *ServerConn) Read(d []byte) (int, error) {
	if d == nil || len(d) == 0 {
		return 0, errors.New("invalid buffer")
	}
	return s.RX_Get(d), nil
}

func (s *ServerConn) Close() {
	panic("implement me")
}

func (s *ServerConn) LocalAddr() *net.TCPAddr {
	return s.LocalAddr()
}

func (s *ServerConn) RemoteAddr() *net.TCPAddr {
	return s.RemoteConn().LocalAddr()
}

func (s *ServerConn) SetSession(sess *Session) {
	s.session = sess
}

func (s *ServerConn) GetSession() *Session {
	return s.session
}
