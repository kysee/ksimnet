package p2p

import (
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

type UpperPeer interface {
	OnAccepted(types.NetConn) error
	OnConnected(types.NetConn) error
	OnReceived(types.NetConn, []byte) error
	OnClosed(types.NetConn) error
}

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
	handledMsgIDs map[types.MsgID]interface{}
	rxCnt         int
	txCnt         int

	seedAddr *net.TCPAddr

	upperPeer UpperPeer
}

var _ types.Peer = (*SimPeer)(nil)

func NewSimPeer(hostIP net.IP, minOthers, maxOthers int, seedAddr *net.TCPAddr) *SimPeer {
	peer := &SimPeer{
		id:     types.NewPeerID(hostIP),
		hostIP: hostIP,
		others: make(map[types.PeerID]types.NetConn),

		stopCh:    make(chan interface{}),
		minOthers: minOthers,
		maxOthers: maxOthers,
		addrBook:  NewAddrBook(),

		//recvBufs:    make(map[uint64][]byte),
		handledMsgIDs: make(map[types.MsgID]interface{}),

		seedAddr: seedAddr,
	}

	peer.broadcastCh = make(chan []byte, peer.maxOthers*2)
	return peer
}

func (peer *SimPeer) SetUpperPeer(upper UpperPeer) {
	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	peer.upperPeer = upper
}
func (peer *SimPeer) SetSeed(addr *net.TCPAddr) {
	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	peer.seedAddr = addr
}

func (peer *SimPeer) GetSeedAddr() *net.TCPAddr {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.seedAddr
}

func (peer *SimPeer) IsMsgHandled(msgId types.MsgID) bool {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	_, ok := peer.handledMsgIDs[msgId]
	return ok
}

func (peer *SimPeer) RxCnt() int {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.rxCnt
	//ret := 0
	//for _, p := range peer.others {
	//	ret += p.(*simnet.NetPoint).RxCnt()
	//}
	//return ret
}

func (peer *SimPeer) TxCnt() int {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.txCnt
	//ret := 0
	//for _, p := range peer.others {
	//	ret += p.(*simnet.NetPoint).TxCnt()
	//}
	//return ret
}

func (peer *SimPeer) Start(listenPort int) error {
	// start listening
	peer.listener = simnet.NewListener(peer)
	if err := peer.listener.Listen(listenPort); err != nil {
		return err
	}

	go pexRoutine(peer)
	go connectPeerRoutine(peer)
	go broadcastRoutine(peer)

	return nil
}

func (peer *SimPeer) ID() types.PeerID {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.id //(int64)(binary.BigEndian.Uint64(peer.hostIP.To4()[:]))
}

func (peer *SimPeer) Send(mb types.MsgBody) (int, error) {
	return peer.SendTo(types.NewZeroPeerID(), mb)
}

func (peer *SimPeer) SendTo(toId types.PeerID, mb types.MsgBody) (int, error) {
	peer.mtx.Lock()
	defer peer.mtx.Unlock()

	peer.txCnt++

	m := NewSimMsg(peer.id, toId, mb)

	pack, err := m.Encode()
	if err != nil {
		return 0, err
	}

	peer.broadcastCh <- pack

	return len(pack), nil
}

func (peer *SimPeer) Stop() {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	peer.listener.Shutdown()

	close(peer.stopCh)
}

func (peer *SimPeer) PeerCnt() int {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return len(peer.others)
}

func (peer *SimPeer) PeerIDs() []types.PeerID {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	var ret []types.PeerID
	for k, _ := range peer.others {
		ret = append(ret, k)
	}
	return ret
}

func (peer *SimPeer) GetPeerConn(id types.PeerID) types.NetConn {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.others[id]
}

func (peer *SimPeer) HasPeer(id types.PeerID) bool {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	_, ok := peer.others[id]
	return ok
}

func (peer *SimPeer) HostIP() net.IP {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return peer.hostIP
}

func (peer *SimPeer) HandledMsgIDs() map[types.MsgID]interface{} {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	ret := make(map[types.MsgID]interface{})
	for k, v := range peer.handledMsgIDs {
		ret[k] = v
	}
	return ret
}

func (peer *SimPeer) HandledMsgCnt() int {
	peer.mtx.RLock()
	defer peer.mtx.RUnlock()

	return len(peer.handledMsgIDs)
}

func (peer *SimPeer) OnConnect(conn types.NetConn) error {
	peerId := types.NewPeerID(conn.RemoteIP())

	peer.mtx.Lock()

	if _, ok := peer.others[peerId]; !ok {
		peer.others[peerId] = conn
		//log.Printf("OnConnect(%s), peer count: %d\n", conn.Key(), len(peer.others))
	}

	if peer.upperPeer != nil {
		peer.mtx.Unlock()
		return peer.upperPeer.OnConnected(conn)
	}
	peer.mtx.Unlock()
	return nil
}

func (peer *SimPeer) OnAccept(conn types.NetConn) error {
	peerId := types.NewPeerID(conn.RemoteIP())

	peer.mtx.Lock()

	if len(peer.others) >= peer.maxOthers {
		peer.mtx.Unlock()
		return errors.New(fmt.Sprintf("the peer(%s) has too many peers(%d)", peer.HostIP(), peer.PeerCnt()))
	}

	if _, ok := peer.others[peerId]; ok {
		peer.mtx.Unlock()
		return errors.New(fmt.Sprintf("the peer(%s) is already connected", conn.RemoteIP()))
	}
	peer.others[peerId] = conn
	//log.Printf("OnAccept(%s), peer count: %d\n", conn.Key(), len(peer.others))

	if peer.upperPeer != nil {
		peer.mtx.Unlock()
		return peer.upperPeer.OnAccepted(conn)
	}

	peer.mtx.Unlock()
	return nil
}

func (peer *SimPeer) OnRecv(conn types.NetConn, pack []byte, sz int) error {
	//log.Printf("[SimPeer::OnReceive] Peer(%s) receives a pack\n", peer.hostIP)
	header := SimMsgHeader{}
	if err := header.Decode(pack); err != nil {
		return err
	}

	peer.mtx.Lock()

	peer.rxCnt++

	// Has this msg been already received ?
	if _, ok := peer.handledMsgIDs[header.MsgID]; ok && header.Dst() == types.NewZeroPeerID() {
		peer.mtx.Unlock()
		//log.Printf("this message(%x) is already handled\n", header.MsgID)
		return fmt.Errorf("peer(%s) handled the message(%d,%s) already", peer.hostIP, header.MsgType, &header.MsgID)
	}

	if header.MsgType == ACK_PEERS {
		msg := &AckPeers{}
		if err := msg.Decode(pack[HeaderSize:]); err == nil {
			//log.Printf("the peer(%s) recieves %d addresses\n", peer.hostIP, len(msg.Addrs))
			for _, addr := range msg.Addrs {
				//log.Printf("Address(%s) will be processed by the peer(%s)\n", addr, peer.hostIP)
				if peer.addrBook.Find(addr) == nil {
					peer.addrBook.Add(addr)
					//log.Printf("The peer(%s) get an peer address(%s), address count: %d\n", peer.hostIP, addr, len(peer.addrBook.addrs))
				}
			}
		}
		peer.mtx.Unlock()
		return nil
	} else if header.Dst() == types.NewZeroPeerID() {
		// broadcast...
		peer.handledMsgIDs[header.MsgID] = struct{}{}
		peer.broadcastCh <- pack
		//log.Printf("Peer(%s) has handled the message(%d,%s)\n", peer.hostIP, header.MsgType, &header.MsgID)
	}

	if peer.upperPeer != nil {
		peer.mtx.Unlock()
		return peer.upperPeer.OnReceived(conn, pack)
	}
	peer.mtx.Unlock()
	return nil
}

func (peer *SimPeer) OnClose(conn types.NetConn) error {
	peer.mtx.Lock()

	peerId := types.NewPeerID(conn.RemoteIP())
	delete(peer.others, peerId)

	//log.Printf("OnClose(%s), peer count: %d\n", conn.Key(), len(peer.others))

	if peer.upperPeer != nil {
		peer.mtx.Unlock()
		return peer.upperPeer.OnClosed(conn)
	}
	peer.mtx.Unlock()
	return nil
}

func pexRoutine(me *SimPeer) {
	c := 1
	du := time.Millisecond * 1000
	ticker := time.NewTicker(du)

Loop:
	for {
		select {
		case <-ticker.C:

			if me.PeerCnt() > me.minOthers {
				c++
				du += du * time.Duration(c)
				ticker.Reset(du)
				//log.Printf("pexRoutine: Peer Count = %d\n", me.PeerCnt())
				continue
			} else if c > 1 {
				c = 1
				du = time.Millisecond * 1000
				ticker.Reset(du)
			}

			me.mtx.RLock()
			seedAddr := me.seedAddr
			listenAddr := me.listener.ListenAddr()
			me.mtx.RUnlock()

			np := simnet.NewNetPoint(me, 0 /*bind unused port*/, true)
			if err := np.Connect(seedAddr.String()); err == nil {

				reqPeersMsg := NewAnonySimMsg(NewReqPeers(listenAddr))
				reqPeersMsg.Header.SetSrc(me.ID())
				pack, err := reqPeersMsg.Encode()
				if err != nil {
					break
				}

				np.Write(pack)

			} else {
				log.Printf("The local peer(%s) can not connect to the peer(%v): %s\n",
					me.HostIP(), seedAddr.String(), err)
			}
			//if me.PeerCnt() >= me.maxOthers {
			//	break Loop
			//}
		case <-me.stopCh:
			break Loop
		}
	}

	//log.Printf("pexRoutine of peer(%s) is over.\n", me.hostIP)
}

func connectPeerRoutine(me *SimPeer) {
	c := 1
	du := time.Millisecond * 1200
	ticker := time.NewTicker(du)

Loop:
	for {
		select {
		case <-ticker.C:
			if me.PeerCnt() > me.minOthers {
				c++
				du += du * time.Duration(c)
				ticker.Reset(du)
				//log.Printf("connectPeerRoutine: Peer:%s has %d peers\n", me.HostIP(), me.PeerCnt())
				continue
			} else if c > 1 {
				c = 1
				du = time.Millisecond * 1200
				ticker.Reset(du)
			}

			me.mtx.RLock()
			addrs := me.addrBook.Addrs()
			me.mtx.RUnlock()

			//log.Printf("The peer(%s) has %d others\n", me.hostIP, me.PeerCnt())

			tryCnt := 0
			for tryCnt < len(addrs) {
				r := rand.Intn(len(addrs))
				toAddr := addrs[r]
				toPeerID := types.NewPeerID(toAddr.IP)

				if !me.HasPeer(toPeerID) && !toAddr.IP.Equal(me.HostIP()) {
					np := simnet.NewNetPoint(types.NetWorker(me), 0 /*bind unused port*/, true)
					if err := np.Connect(toAddr.String()); err != nil {
						//log.Printf("The local peer(%s) can not connect to the peer(%v): %s\n",
						//	me.HostIP(), toAddr, err)
					} else if me.PeerCnt() >= me.maxOthers {
						//log.Printf("discoverPeersRoutine: Peer:%s has %d peers over max\n", me.HostIP(), me.PeerCnt())
						break
					}
				}
				tryCnt++
			}

			//if len(addrs) > 0 {
			//
			//	for _, toAddr := range addrs {
			//
			//		peerId := types.NewPeerID(toAddr.IP)
			//
			//		if !me.HasPeer(peerId) && !toAddr.IP.Equal(me.HostIP()) {
			//			np := simnet.NewNetPoint(types.NetWorker(me), 0 /*bind unused port*/, true)
			//			if err := np.Connect(toAddr.String()); err != nil {
			//				//log.Printf("The local peer(%s) can not connect to the peer(%v): %s\n",
			//				//	me.HostIP(), toAddr, err)
			//			} else if me.PeerCnt() >= me.maxOthers {
			//				//log.Printf("discoverPeersRoutine: Peer:%s has %d peers over max\n", me.HostIP(), me.PeerCnt())
			//				break
			//			}
			//		}
			//	}
			//}

		case <-me.stopCh:
			break Loop
		}

	}
	//log.Printf("connectPeerRoutine of peer(%s) is over.\n", me.hostIP)
}

func broadcastRoutine(me *SimPeer) {
Loop:
	for {
		select {
		case brdPack := <-me.broadcastCh:
			h := &SimMsgHeader{}
			if err := h.Decode(brdPack); err != nil {
				panic(err)
			}

			dstId := h.Dst()
			if dstId == types.NewZeroPeerID() {
				me.mtx.RLock()
				for _, conn := range me.others {
					if _, err := conn.Write(brdPack); err != nil {
						panic("writing to peer(" + conn.Key() + ") is failed : " + err.Error())
					}
				}
				me.mtx.RUnlock()
			} else {
				if conn := me.GetPeerConn(dstId); conn != nil {
					if _, err := conn.Write(brdPack); err != nil {
						panic("writing to peer(" + conn.Key() + ") is failed : " + err.Error())
					}
				} else {
					//panic("not found peer id(" + dstId.String() + ")")
				}
			}

		case <-me.stopCh:
			break Loop
		}
	}
	//log.Printf("broadcastRoutine of peer(%s) is over.\n", me.hostIP)
}
