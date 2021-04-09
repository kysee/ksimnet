package p2p

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/kysee/ksimnet/simnet"
	"github.com/kysee/ksimnet/types"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type SimPeer struct {
	mtx sync.RWMutex

	id       types.PeerID
	hostIP   net.IP
	listener *simnet.Listener
	stopCh   chan interface{}

	others    map[types.PeerID]types.NetConn
	minOthers int //var MinOthers = 3
	maxOthers int //var MaxOthers = 100
	addrBook  *AddrBook

	broadcastCh chan []byte
	//recvBufs map[uint64][]byte
	handledMsgIDs map[[sha256.Size]byte]interface{}
}

func NewSimPeer(hostIP net.IP, minOthers, maxOthers int) *SimPeer {
	peer := &SimPeer{
		id:     types.NewPeerID(),
		hostIP: hostIP,
		others: make(map[types.PeerID]types.NetConn),

		stopCh:    make(chan interface{}),
		minOthers: minOthers,
		maxOthers: maxOthers,
		addrBook:  NewAddrBook(),

		//recvBufs:    make(map[uint64][]byte),
		handledMsgIDs: make(map[[sha256.Size]byte]interface{}),
	}

	peer.broadcastCh = make(chan []byte, peer.maxOthers*2)
	return peer
}

func (peer *SimPeer) Start(listenPort int) error {
	// start listening
	peer.listener = simnet.NewListener(peer)
	if err := peer.listener.Listen(listenPort); err != nil {
		return err
	}

	copy(peer.id[:], peer.hostIP.String())

	// start discovery routine
	go discoverPeersRoutine(peer)
	go broadcastRoutine(peer)

	return nil
}

func (peer *SimPeer) ID() types.PeerID {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.id //(int64)(binary.BigEndian.Uint64(peer.hostIP.To4()[:]))
}

func (peer *SimPeer) Send(d []byte) (int, error) {
	srcId := peer.ID()
	mHash := sha256.Sum256(d)

	pack := make([]byte, types.PeerIDSize+sha256.Size+len(d))
	n1 := copy(pack, srcId[:])
	n2 := copy(pack[n1:], mHash[:])
	copy(pack[n1+n2:], d)

	// broadcast
	peer.broadcastCh <- pack

	return len(pack), nil
}

func (peer *SimPeer) Stop() {
	peer.stopCh <- struct{}{}
}

func (peer *SimPeer) PeerCnt() int {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return len(peer.others)
}

func (peer *SimPeer) HasPeer(id types.PeerID) bool {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	_, ok := peer.others[id]
	return ok
}

func (peer *SimPeer) OnConnect(conn types.NetConn) {
	var peerId types.PeerID
	copy(peerId[:], []byte(conn.RemoteIP().String()))

	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	if _, ok := peer.others[peerId]; !ok {
		peer.others[peerId] = conn
	}
}

func (peer *SimPeer) HostIP() net.IP {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.hostIP
}

func (peer *SimPeer) OnRecv(conn types.NetConn, pack []byte, sz int) (int, error) {
	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	// Has this msg been already received ?
	var msgId [sha256.Size]byte
	copy(msgId[:], pack[types.PeerIDSize:types.PeerIDSize+sha256.Size])
	if _, ok := peer.handledMsgIDs[msgId]; ok {
		return 0, errors.New("this message is already handled")
	}

	peer.handledMsgIDs[msgId] = struct{}{}

	return sz, nil
}

func (peer *SimPeer) OnClose(conn types.NetConn) {
	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	var peerId types.PeerID
	copy(peerId[:], []byte(conn.RemoteIP().String()))
	delete(peer.others, peerId)
}

func (peer *SimPeer) OnAccept(conn types.NetConn) error {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	var peerId types.PeerID
	copy(peerId[:], []byte(conn.RemoteIP().String()))

	if _, ok := peer.others[peerId]; ok {
		return errors.New(fmt.Sprintf("the peer(%s) is already connected", conn.RemoteIP()))
	}
	peer.others[peerId] = conn
	return nil
}

var _ types.Peer = (*SimPeer)(nil)

func discoverPeersRoutine(me *SimPeer) {
	ticker := time.NewTicker(time.Millisecond * 500)

Loop:
	for {
		select {
		case <-ticker.C:
			if me.PeerCnt() < me.minOthers {

				addrs := me.addrBook.Addrs()

				for {
					rdx := rand.Intn(len(addrs))
					toAddr := addrs[rdx]

					var peerId types.PeerID
					copy(peerId[:], []byte(toAddr.IP.String()))

					if !me.HasPeer(peerId) && !toAddr.IP.Equal(me.HostIP()) {
						np := simnet.NewNetPoint(me, 0 /*bind unused port*/)
						if err := np.Connect(toAddr.String()); err == nil {
							break
						} else {
							log.Printf("The local peer(%s) can not connect to the peer(%s): %s\n",
								me.HostIP(), toAddr, err)
						}
					}
				}
			} else {
				break Loop
			}

		case <-me.stopCh:
			break Loop
		}

	}
}

func broadcastRoutine(me *SimPeer) {
Loop:
	for {
		select {
		case brdPack := <-me.broadcastCh:

			me.mtx.RLock()
			for _, conn := range me.others {
				if _, err := conn.Write(brdPack); err != nil {
					panic("writing to peer(" + conn.Key() + ") is failed : " + err.Error())
				}
			}
			me.mtx.RUnlock()

		case <-me.stopCh:
			break Loop
		}
	}
}
