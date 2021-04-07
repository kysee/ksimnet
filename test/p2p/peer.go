package p2p

import (
	"github.com/ksimnet/simnet"
	"github.com/ksimnet/types"
)

var PeerCnt = 100
var AddrBook []string

type Peer struct {
	server     *simnet.Listener
	others     []types.NetConn
	stopListen chan<- struct{}
}

func NewPeer() *Peer {
	peer := &Peer{
		others: make([]types.NetConn, 50),
	}
	return peer
}

func (peer *Peer) Start(laddr string) {
	s, err := simnet.NewListener(peer, laddr)
	if err != nil {
		panic(err)
	}

	err = s.Listen()
	if err != nil {
		panic(err)
	}

	peer.server = s

	// todo: add the peer.server.Key() to AddrBook[]
}

var _ types.ServerWorker = (*Peer)(nil)
var _ types.ClientWorker = (*Peer)(nil)

func (peer *Peer) OnConnect(conn types.NetConn) {
	panic("implement me")
}

func (peer *Peer) OnAccept(conn types.NetConn) error {
	panic("implement me")
}

func (peer *Peer) OnRecv(conn types.NetConn, bytes []byte, i int) (int, error) {
	panic("implement me")
}

func (peer *Peer) OnClose(conn types.NetConn) {
	panic("implement me")
}
