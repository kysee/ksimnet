package types

import "net"

type NetWorker interface {
	HostIP() net.IP
	OnRecv(NetConn, []byte, int) (int, error)
	OnClose(NetConn)
}

type ClientWorker interface {
	OnConnect(NetConn)
	NetWorker
}

type ServerWorker interface {
	OnAccept(NetConn) error
	NetWorker
}
