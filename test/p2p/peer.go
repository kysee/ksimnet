package p2p

import (
	"encoding/binary"
	"errors"
	"github.com/ksimnet/simnet"
	"github.com/ksimnet/types"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

var MinOthers = 3
var MaxOthers = 100
var AddrBook []string
var addrBookMtx sync.RWMutex
var WG = sync.WaitGroup{}

type BroadcastPack struct {
	src  types.NetConn
	pack []byte
}

type Peer struct {
	mtx sync.RWMutex

	server *simnet.Listener
	others map[string]types.NetConn

	recvMsgs map[uint64][]byte

	broadcastCh chan *BroadcastPack
	stopCh      chan interface{}
}

func NewPeer(laddr string) *Peer {
	peer := &Peer{
		others:      make(map[string]types.NetConn),
		recvMsgs:    make(map[uint64][]byte),
		broadcastCh: make(chan *BroadcastPack, MaxOthers*2),
		stopCh:      make(chan interface{}),
	}

	s, err := simnet.NewListener(peer, laddr)
	if err != nil {
		panic(err)
	}
	peer.server = s
	return peer
}

func (peer *Peer) Start() {
	err := peer.server.Listen()
	if err != nil {
		panic(err)
	}

	addrBookMtx.Lock()
	AddrBook = append(AddrBook, peer.server.ListenAddr().String())
	addrBookMtx.Unlock()

	go peer.connectRoutine()
	go peer.broadcastRoutine()
}

func (peer *Peer) addPeerConn(conn types.NetConn) {
	peer.others[conn.RemoteIP().String()] = conn
}

func (peer *Peer) delPeerConn(conn types.NetConn) {
	delete(peer.others, conn.RemoteIP().String())
}

func (peer *Peer) hasPeerConn(conn types.NetConn) bool {
	return peer.hasPeer(conn.RemoteIP().String())
}

func (peer *Peer) hasPeer(ip string) bool {
	if _, ok := peer.others[ip]; ok {
		return true
	}
	return false
}

func (peer *Peer) HasPeer(ip string) bool {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.hasPeer(ip)
}

func (peer *Peer) PeerCnt() int {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return len(peer.others)
}

func (peer *Peer) LocalListenAddr() *net.TCPAddr {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.server.ListenAddr()
}

func (peer *Peer) LocalIP() net.IP {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.server.ListenAddr().IP
}

func (peer *Peer) RecvMsgCnt() int {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return len(peer.recvMsgs)
}

func (peer *Peer) RecvMsg(k uint64) []byte {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	if m, ok := peer.recvMsgs[k]; ok {
		return m
	}
	return nil
}

func (peer *Peer) Broadcast(n uint64, d []byte, srcConn types.NetConn) {

	// encoding a packet
	_pack := make([]byte, 8+len(d))
	copy(_pack[8:], d)
	binary.BigEndian.PutUint64(_pack[:8], n)

	brdPack := &BroadcastPack{
		pack: _pack,
		src:  srcConn,
	}

	peer.broadcastCh <- brdPack

}

func (peer *Peer) broadcastRoutine() {
Loop:
	for {
		select {
		case brdPack := <-peer.broadcastCh:
			srcKey := ""
			if brdPack.src != nil {
				srcKey = brdPack.src.Key()
			}

			peer.mtx.RLock()
			for k, p := range peer.others {
				if srcKey == "" || k != srcKey {
					if _, err := p.Write(brdPack.pack); err != nil {
						panic("Wrting to peer(" + k + ") is failed: " + err.Error())
					}
				}
			}
			peer.mtx.RUnlock()

		case <-peer.stopCh:
			break Loop
		}
	}
}

func (peer *Peer) connectRoutine() {

	ticker := time.NewTicker(time.Millisecond * 500)

Loop:
	for {
		select {
		case <-ticker.C:
			if peer.PeerCnt() < MinOthers {
				addrBookMtx.RLock()
				addrs := make([]string, len(AddrBook))
				copy(addrs, AddrBook)
				addrBookMtx.RUnlock()

				for {
					rdx := rand.Intn(len(addrs))
					addr := addrs[rdx]

					tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
					if err != nil {
						panic(err)
					}
					ip := tcpAddr.IP.String()
					if !peer.HasPeer(ip) && ip != peer.LocalListenAddr().IP.String() {
						np := simnet.NewNetPoint(peer, simnet.BindPort(peer.LocalIP().String()))
						if err := np.Connect(addr); err == nil {
							//if _, err := simnet.Connect(peer, peer.LocalListenAddr().IP.String(), addr); err == nil {
							break
						} else {
							log.Printf("the peer(%s) can not connect to the peer(%s): %s\n",
								peer.LocalIP(), addr, err)
						}
					}
				}
			} else {
				break Loop
			}

		case <-peer.stopCh:
			break Loop
		}

	}

	WG.Done()
}

var _ types.ServerWorker = (*Peer)(nil)
var _ types.ClientWorker = (*Peer)(nil)

func (peer *Peer) OnConnect(conn types.NetConn) {
	//log.Printf("OnConnect - peer(%s) <> peer(%s)\n", conn.LocalAddr(), conn.RemoteAddr())
	peer.mtx.Lock()
	defer peer.mtx.Unlock()
	peer.addPeerConn(conn)
}

func (peer *Peer) OnAccept(conn types.NetConn) error {
	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	if len(peer.others) >= MaxOthers {
		return errors.New("too many clients")
	}

	if peer.hasPeerConn(conn) {
		return errors.New("the peer(" + conn.Key() + ") is already connected")
	}
	peer.addPeerConn(conn)
	return nil
}

func (peer *Peer) OnRecv(conn types.NetConn, d []byte, size int) (int, error) {
	bn := binary.BigEndian.Uint64(d[:8])

	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	if _, ok := peer.recvMsgs[bn]; ok {
		return 0, errors.New("this msg is already received")
	}

	peer.recvMsgs[bn] = d[8:]

	WG.Done()
	//log.Printf("Peer(%s) received msg[%d] from peer(%s)\n", conn.LocalAddr(), bn, conn.RemoteAddr())

	// broadcast to others
	peer.broadcastCh <- &BroadcastPack{
		pack: d,
		src:  conn,
	}

	return size, nil
}

func (peer *Peer) OnClose(conn types.NetConn) {
	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	peer.delPeerConn(conn)
}
