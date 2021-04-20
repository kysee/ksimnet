package types

import "net"

type NetWorker interface {
	HostIP() net.IP
	OnRecv(NetConn, []byte, int) error
	OnClose(NetConn) error
}

type ClientWorker interface {
	OnConnect(NetConn) error
	NetWorker
}

type ServerWorker interface {
	OnAccept(NetConn) error
	NetWorker
}
