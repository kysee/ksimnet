package netconn

import "net"

type NetOps interface {
	Key() string
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Close()
	LocalAddr() *net.TCPAddr
	RemoteAddr() *net.TCPAddr
}
