package p2p

import "github.com/ksimnet/netconn"

type Peer struct {

}

func (p Peer) OnConnect(point *netconn.NetPoint) error {
	panic("implement me")
}

func (p Peer) OnAccept(point *netconn.NetPoint) error {
	panic("implement me")
}

func (p Peer) OnRecv(point *netconn.NetPoint, bytes []byte, i int) (int, error) {
	panic("implement me")
}

func (p Peer) OnClose(point *netconn.NetPoint) error {
	panic("implement me")
}

var _ netconn.ServerWorker = (*Peer)(nil)
var _ netconn.ClientWorker = (*Peer)(nil)

