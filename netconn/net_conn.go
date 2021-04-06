package netconn

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
}
