package p2p

import (
	"encoding/binary"
	"errors"
	"github.com/ksimnet/simnet"
	"github.com/ksimnet/types"
	"net"
	"sync"
	"time"
)

var MaxOthers = 5
var AddrBook []string
var WG = sync.WaitGroup{}

type Peer struct {
	mtx sync.RWMutex

	server *simnet.Listener
	others map[string]types.NetConn

	recvMsgs map[uint64][]byte

	stopCh chan interface{}
}

func NewPeer(laddr string) *Peer {
	peer := &Peer{
		others:   make(map[string]types.NetConn),
		recvMsgs: make(map[uint64][]byte),
		stopCh:   make(chan interface{}),
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

	peer.mtx.Lock()
	defer peer.mtx.Unlock()
	AddrBook = append(AddrBook, peer.server.ListenAddr().String())

	go peer.connectRoutine()
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

func (peer *Peer) LocalAddr() *net.TCPAddr {
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

func (peer *Peer) Broadcast(n uint64, d []byte) {

	// encoding a packet
	pack := make([]byte, 8+len(d))
	copy(pack[8:], d)
	binary.BigEndian.PutUint64(pack[:8], n)

	go peer.broadcastRoutine(pack, nil)
}

func (peer *Peer) broadcastRoutine(d []byte, srcConn types.NetConn) {
	srcKey := ""
	if srcConn != nil {
		srcKey = srcConn.Key()
	}

	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	for k, p := range peer.others {
		if srcKey == "" || k != srcKey {
			if _, err := p.Write(d); err != nil {
				panic("Wrting to peer(" + k + ") is failed: " + err.Error())
			}
		}
	}
}

func (peer *Peer) connectRoutine() {

	ticker := time.NewTicker(time.Millisecond * 1000)

Loop:
	for {
		select {
		case <-ticker.C:
			if len(peer.others) < MaxOthers {
				peer.mtx.RLock()
				addrs := make([]string, len(AddrBook))
				copy(addrs, AddrBook)
				peer.mtx.RUnlock()

				for _, addr := range addrs {
					tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
					if err != nil {
						panic(err)
					}
					ip := tcpAddr.IP.String()
					if !peer.HasPeer(ip) && ip != peer.LocalAddr().IP.String() {
						if _, err := simnet.Connect(peer, peer.LocalAddr().IP.String(), addr); err == nil {
							break
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
	go peer.broadcastRoutine(d, conn)

	return size, nil
}

func (peer *Peer) OnClose(conn types.NetConn) {
	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	peer.delPeerConn(conn)
}
