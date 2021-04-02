package knetsim

import (
	"net"
	"sync"
)

type Conn struct {
	sendMtx sync.Mutex
	recvMtx sync.Mutex

	local  net.TCPAddr
	remote net.TCPAddr

	sendQ map[int][]byte
	recvQ map[int][]byte
}

func Listen(ip string, port int) {

}

func Dial(ip string, port int) *Conn {
	// find backlog
	return nil
}

func (c *Conn) Send(d []byte, size int) int {
	return 0
}

func (c *Conn) Recv(d []byte, size int) int {
	return 0
}
