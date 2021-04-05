package simnet

import (
	"errors"
	"github.com/ksimnet/netapp"
	"github.com/ksimnet/netconn"
	"net"
	"strconv"
)

type ClientConn struct {
	*netconn.NetConn

	app     netapp.ClientApp
	session *Session
}

var _ netconn.NetOps = (*ClientConn)(nil)

func NewClientConn(a netapp.ClientApp, hostIp string) (*ClientConn, error) {
	// bind
	port := NewPort(hostIp)
	tcpAddr, err := net.ResolveTCPAddr("tcp", hostIp+":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	netConn, err := netconn.NewNetConn(tcpAddr)
	if err != nil {
		return nil, err
	}
	return &ClientConn{
		NetConn: netConn,
		app:     a,
	}, nil
}

func (c *ClientConn) Key() string {
	return c.LocalAddr().String()
}

func (c *ClientConn) Write(d []byte) (int, error) {
	if d == nil || len(d) == 0 {
		return 0, errors.New("invalid buffer")
	}
	return c.RemoteConn().RX_Put(d)
}

func (c *ClientConn) Read(d []byte) (int, error) {
	if d == nil || len(d) == 0 {
		return 0, errors.New("invalid buffer")
	}
	return c.RX_Get(d), nil
}

func (c *ClientConn) Close() {
	panic("implement close")
}

//func (c *ClientConn) LocalAddr() *net.TCPAddr {
//	return c.LocalAddr()
//}
//
//func (c *ClientConn) RemoteAddr() *net.TCPAddr {
//	return c.RemoteAddr()
//}

func (c *ClientConn) SetSession(sess *Session) {
	c.session = sess
}

func (c *ClientConn) GetSession() *Session {
	return c.session
}
