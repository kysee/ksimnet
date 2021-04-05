package netapp

import (
	"github.com/ksimnet/netconn"
)

type NetApp interface {
	OnRecv(*netconn.NetConn, []byte, int) error
	OnClose(*netconn.NetConn) error
}

type ClientApp interface {
	OnConnect(*netconn.NetConn) error
	NetApp
}

type ServerApp interface {
	OnAccept(*netconn.NetConn) error
	NetApp
}
