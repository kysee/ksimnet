package p2p

import (
	"errors"
	"github.com/kysee/ksimnet/simnet"
	"github.com/kysee/ksimnet/types"
	"net"
	"sync"
)

type SimSeedPeer struct {
	mtx sync.RWMutex

	id       types.PeerID
	hostIP   net.IP
	listener *simnet.Listener
	stopCh   chan interface{}

	minOthers int //var MinOthers = 3
	maxOthers int //var MaxOthers = 100
	addrBook  *AddrBook
}

func NewSimSeedPeer(hostIP net.IP, minOthers, maxOthers int) *SimSeedPeer {
	peer := &SimSeedPeer{
		id:     types.NewRandPeerID(),
		hostIP: hostIP,

		stopCh:    make(chan interface{}),
		minOthers: minOthers,
		maxOthers: maxOthers,
		addrBook:  NewAddrBook(),
	}

	return peer
}

func (seed *SimSeedPeer) ID() types.PeerID {
	seed.mtx.RLock()
	defer seed.mtx.RUnlock()

	return seed.id
}

func (seed *SimSeedPeer) Start(listenPort int) error {
	// start listening
	seed.listener = simnet.NewListener(seed)
	if err := seed.listener.Listen(listenPort); err != nil {
		return err
	}

	copy(seed.id[:], seed.hostIP.String())
	return nil
}

func (seed *SimSeedPeer) Send(d []byte) (int, error) {
	panic("implement me")
}

func (seed *SimSeedPeer) Stop() {
	panic("implement me")
}

func (seed *SimSeedPeer) PeerCnt() int {
	panic("implement me")
}

func (seed *SimSeedPeer) HasPeer(id types.PeerID) bool {
	panic("implement me")
}

func (seed *SimSeedPeer) OnConnect(conn types.NetConn) {
	panic("implement me")
}

func (seed *SimSeedPeer) HostIP() net.IP {
	seed.mtx.RLock()
	defer seed.mtx.RUnlock()

	return seed.hostIP
}

func (seed *SimSeedPeer) HostAddr() net.TCPAddr {
	seed.mtx.RLock()
	defer seed.mtx.RUnlock()

	return *seed.listener.ListenAddr()
}

func (seed *SimSeedPeer) OnRecv(conn types.NetConn, pack []byte, i int) (int, error) {
	seed.mtx.Lock()
	defer func() {
		seed.mtx.Unlock()
		conn.Close()
	}()

	//log.Printf("[SimSeed::OnReceive] Peer(%s) receives a pack\n", seed.hostIP)

	header := &Header{}
	if err := header.Decode(pack); err != nil {
		return 0, err
	}
	if header.MsgType != REQ_PEERS {
		return 0, errors.New("this message type is not REQ_PEERS")
	}

	reqPeers := &ReqPeers{}
	if err := reqPeers.Decode(pack[HeaderSize:]); err != nil {
		return 0, err
	}

	alreadyAdded := false
	ackPeersBody := NewAckPeers()
	addrs := seed.addrBook.Addrs()
	for _, addr := range addrs {
		if !addr.IP.Equal(conn.RemoteIP()) {
			ackPeersBody.AddPeerAddr(addr)
		}
		if alreadyAdded == false && EqualAddr(addr, reqPeers.ExportAddr) {
			alreadyAdded = true
		}
	}
	ackPeersMsg := NewAnonySimMsg(ackPeersBody)
	ackPeersMsg.SetSrc(seed.id)
	if pack, err := ackPeersMsg.Encode(); err != nil {
		return 0, err
	} else {
		conn.Write(pack)
		//log.Printf("%d addresses are sent to the peer(%s)\n", len(ackPeersBody.Addrs), conn.Key())
	}

	if !alreadyAdded {
		seed.addrBook.Add(reqPeers.ExportAddr)
		//log.Printf("Add address(%s) to AddrBook, address count: %d\n", reqPeers.ExportAddr.String(), len(seed.addrBook.addrs))
	}

	return 0, nil
}

func (seed *SimSeedPeer) OnClose(conn types.NetConn) {

}

func (seed *SimSeedPeer) OnAccept(conn types.NetConn) error {
	return nil
}

var _ types.Peer = (*SimSeedPeer)(nil)
