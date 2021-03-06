package types

import (
	"net"
)

type NetConn interface {
	Key() string
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Close()

	LocalAddr() *net.TCPAddr
	RemoteAddr() *net.TCPAddr

	LocalIP() net.IP
	LocalPort() int
	RemoteIP() net.IP
	RemotePort() int
}
