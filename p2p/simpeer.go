package p2p

import (
	"encoding/binary"
	"github.com/kysee/ksimnet/simnet"
	net_types "github.com/kysee/ksimnet/types"
	"github.com/kysee/snowball/types"
	"net"
	"sync"
)

type SimPeer struct {
	mtx sync.RWMutex

	hostIP   net.IP
	listener *simnet.Listener
}

func NewSimPeer(ip net.IP) *SimPeer {
	return &SimPeer{
		hostIP: ip,
	}
}

func (sp *SimPeer) ID() int64 {
	sp.mtx.RLock()
	defer sp.mtx.RUnlock()

	return (int64)(binary.BigEndian.Uint64(sp.hostIP[:]))
}

func (sp *SimPeer) Start(listenPort int) error {
	// start listening
	sp.listener = simnet.NewListener(sp)
	if err := sp.listener.Listen(listenPort); err != nil {
		return err
	}

	// start discovery routine
	go discoverPeersRoutine(sp)

	return nil
}

func (sp *SimPeer) Connect() {
	panic("implement me")
}

func (sp *SimPeer) Send() {
	panic("implement me")
}

func (sp *SimPeer) Stop() {
	panic("implement me")
}

func (sp *SimPeer) PeerCnt() int {
	panic("implement me")
}

func (sp *SimPeer) FindPeer(id int64) types.Peer {
	panic("implement me")
}

func (sp *SimPeer) OnConnect(conn net_types.NetConn) {
	panic("implement me")
}

func (sp *SimPeer) HostIP() net.IP {
	sp.mtx.RLock()
	defer sp.mtx.RUnlock()

	return sp.hostIP
}

func (sp *SimPeer) OnRecv(conn net_types.NetConn, d []byte, sz int) (int, error) {
	panic("implement me")
}

func (sp *SimPeer) OnClose(conn net_types.NetConn) {
	panic("implement me")
}

func (sp *SimPeer) OnAccept(conn net_types.NetConn) error {
	panic("implement me")
}

func discoverPeersRoutine(me types.Peer) {
	// connect to seed

	// get ip address of peers

	// add address book
}
