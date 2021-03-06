package p2p

import (
	"encoding/binary"
	"errors"
	"github.com/kysee/ksimnet/simnet"
	"github.com/kysee/ksimnet/types"
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

type TPeer struct {
	mtx sync.RWMutex

	hostIP   net.IP
	listener *simnet.Listener
	others   map[string]types.NetConn

	recvMsgs map[uint64][]byte

	broadcastCh chan *BroadcastPack
	stopCh      chan interface{}
}

func NewPeer(hostIp net.IP) *TPeer {
	peer := &TPeer{
		hostIP:      hostIp,
		others:      make(map[string]types.NetConn),
		recvMsgs:    make(map[uint64][]byte),
		broadcastCh: make(chan *BroadcastPack, MaxOthers*2),
		stopCh:      make(chan interface{}),
	}
	return peer
}

func (peer *TPeer) Start(port int) {
	peer.listener = simnet.NewListener(peer)
	if err := peer.listener.Listen(port); err != nil {
		panic(err)
	}

	addrBookMtx.Lock()
	AddrBook = append(AddrBook, peer.listener.ListenAddr().String())
	addrBookMtx.Unlock()

	go peer.connectRoutine()
	go peer.broadcastRoutine()
}

func (peer *TPeer) addPeerConn(conn types.NetConn) {
	peer.others[conn.RemoteIP().String()] = conn
}

func (peer *TPeer) delPeerConn(conn types.NetConn) {
	delete(peer.others, conn.RemoteIP().String())
}

func (peer *TPeer) hasPeerConn(conn types.NetConn) bool {
	return peer.hasPeer(conn.RemoteIP().String())
}

func (peer *TPeer) hasPeer(ip string) bool {
	if _, ok := peer.others[ip]; ok {
		return true
	}
	return false
}

func (peer *TPeer) HasPeer(ip string) bool {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.hasPeer(ip)
}

func (peer *TPeer) PeerCnt() int {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return len(peer.others)
}

func (peer *TPeer) HostIP() net.IP {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.hostIP
}

func (peer *TPeer) RecvMsgCnt() int {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return len(peer.recvMsgs)
}

func (peer *TPeer) RecvMsg(k uint64) []byte {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	if m, ok := peer.recvMsgs[k]; ok {
		return m
	}
	return nil
}

func (peer *TPeer) Broadcast(n uint64, d []byte, srcConn types.NetConn) {

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

func (peer *TPeer) broadcastRoutine() {
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

func (peer *TPeer) connectRoutine() {

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
					if !peer.HasPeer(ip) && ip != peer.HostIP().String() {
						np := simnet.NewNetPoint(peer, 0 /*bind unused port*/, true)
						if err := np.Connect(addr); err == nil {
							break
						} else {
							log.Printf("the peer(%s) can not connect to the peer(%s): %s\n",
								peer.HostIP(), addr, err)
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

var _ types.ServerWorker = (*TPeer)(nil)
var _ types.ClientWorker = (*TPeer)(nil)

func (peer *TPeer) OnConnect(conn types.NetConn) error {
	//log.Printf("OnConnect - peer(%s) <> peer(%s)\n", conn.LocalAddr(), conn.RemoteAddr())
	peer.mtx.Lock()
	defer peer.mtx.Unlock()
	peer.addPeerConn(conn)

	return nil
}

func (peer *TPeer) OnAccept(conn types.NetConn) error {
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

func (peer *TPeer) OnRecv(conn types.NetConn, d []byte, size int) error {
	bn := binary.BigEndian.Uint64(d[:8])

	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	if _, ok := peer.recvMsgs[bn]; ok {
		return errors.New("this msg is already received")
	}

	peer.recvMsgs[bn] = d[8:]

	WG.Done()
	//log.Printf("TPeer(%s) received msg[%d] from peer(%s)\n", conn.LocalAddr(), bn, conn.RemoteAddr())

	// broadcast to others
	peer.broadcastCh <- &BroadcastPack{
		pack: d,
		src:  conn,
	}

	return nil
}

func (peer *TPeer) OnClose(conn types.NetConn) error {
	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	peer.delPeerConn(conn)

	return nil
}
