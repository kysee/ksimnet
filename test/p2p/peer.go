package p2p

import (
	"github.com/ksimnet/netconn"
	"github.com/ksimnet/simnet"
)

var PeerCnt = 100
var AddrBook []string = make([]string, PeerCnt)

type Peer struct {
	server     *simnet.Server
	others     []*netconn.NetPoint
	stopListen chan<- struct{}
}

func NewPeer() *Peer {
	peer := &Peer{
		others: make([]*netconn.NetPoint, 50),
	}
	return peer
}

func (peer *Peer) Start(laddr string) {
	s, err := simnet.NewServer(peer, laddr)
	if err != nil {
		panic(err)
	}

	stopListen, err := s.Listen()
	if err != nil {
		panic(err)
	}

	peer.server = s
	peer.stopListen = stopListen

	// todo: add the peer.server.Key() to AddrBook[]
}

func (peer *Peer) OnConnect(point *netconn.NetPoint) error {
	panic("implement me")
}

func (peer *Peer) OnAccept(point *netconn.NetPoint) error {
	panic("implement me")
}

func (peer *Peer) OnRecv(point *netconn.NetPoint, bytes []byte, i int) (int, error) {
	panic("implement me")
}

func (peer *Peer) OnClose(point *netconn.NetPoint) error {
	panic("implement me")
}

var _ netconn.ServerWorker = (*Peer)(nil)
var _ netconn.ClientWorker = (*Peer)(nil)
