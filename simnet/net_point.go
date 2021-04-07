package simnet

import (
	"errors"
	"fmt"
	"github.com/ksimnet/types"
	"log"
	"net"
	"sync"
)

type NetPoint struct {
	mtx sync.RWMutex

	worker  types.NetWorker
	session *Session

	localAddr   *net.TCPAddr
	remotePoint *NetPoint

	rxMtx      sync.RWMutex
	rxBuf      map[int][]byte
	rxSeqFront int
	rxSeqEnd   int
	rxCh       chan int

	//txMtx sync.RWMutex
	//txBuf map[int][]byte
	//txSeqFirst int
	//txSeqLast int

	done chan struct{}
}

func NewNetPoint(worker types.NetWorker, localAddr *net.TCPAddr) *NetPoint {
	ret := &NetPoint{
		worker:     worker,
		localAddr:  localAddr,
		rxBuf:      make(map[int][]byte),
		rxSeqFront: 0,
		rxSeqEnd:   0,
		rxCh:       make(chan int, 1024),
		//txBuf: make(map[int][]byte),
		done: make(chan struct{}),
	}

	go receiveRoutine(ret)

	return ret
}

func receiveRoutine(nc *NetPoint) {
Loop:
	for {
		select {
		case <-nc.rxCh:
			//fmt.Println("[",nc.LocalAddr().String(),"] receiveRoutine notified:", seq, "front:", nc.rxSeqFront, "end:", nc.rxSeqEnd, "ch length:", len(nc.rxCh))

			if d := nc.pickRX(); d != nil {
				nc.Worker().OnRecv(nc, d, len(d))
			}
		case <-nc.done:
			break Loop
		}
	}

	fmt.Println("[", nc.LocalAddr().String(), "] Goodbye...")
}

func (np *NetPoint) Worker() types.NetWorker {
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.worker
}

func (np *NetPoint) SetSession(sess *Session) {
	np.mtx.Lock()
	defer np.mtx.Unlock()

	np.session = sess
}

func (np *NetPoint) GetSession() *Session {
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.session
}

func (np *NetPoint) SetRemotePoint(r *NetPoint) {
	np.mtx.Lock()
	defer np.mtx.Unlock()

	np.remotePoint = r
}

func (np *NetPoint) RemotePoint() *NetPoint {
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.remotePoint
}

func (np *NetPoint) putRX(d []byte) (int, error) {
	np.rxMtx.Lock()

	b := make([]byte, len(d))
	n := copy(b, d)

	seq := np.rxSeqEnd
	np.rxBuf[seq] = b
	np.rxSeqEnd++

	np.rxMtx.Unlock()

	np.rxCh <- seq

	return n, nil
}

func (np *NetPoint) getRX(d []byte) int {
	np.rxMtx.Lock()
	defer np.rxMtx.Unlock()

	copied := 0
	size := cap(d)

	for copied < size {
		p, ok := np.rxBuf[np.rxSeqFront]
		if !ok {
			break
		} else if size < len(p) {
			copy(d[copied:], p[:size])
			copied = copied + size
			np.rxBuf[np.rxSeqFront] = p[size:]
		} else {
			copy(d[copied:], p)
			copied = copied + len(p)
			np.rxBuf[np.rxSeqFront] = nil
			np.rxSeqFront++
		}
	}
	return copied
}

func (np *NetPoint) pickRX() []byte {
	np.rxMtx.Lock()
	defer np.rxMtx.Unlock()

	p, ok := np.rxBuf[np.rxSeqFront]
	if !ok {
		fmt.Println(np.localAddr, "end pickRX")
		return nil
	}
	np.rxBuf[np.rxSeqFront] = nil
	np.rxSeqFront++

	return p
}

//func (nc *NetPoint) TX_Put(d []byte) int {
//	nc.txMtx.Lock()
//	defer nc.txMtx.Unlock()
//
//	b := make([]byte, len(d))
//	n := copy(b, d)
//
//	seq := len(nc.txBuf)
//	nc.txBuf[seq] = b
//	nc.txSeqLast = seq
//
//	return n
//}
//
//func (nc *NetPoint) TX_Get(d []byte) int {
//	nc.txMtx.Lock()
//	defer nc.txMtx.Unlock()
//
//	copied := 0
//	size := cap(d)
//
//	for copied < size {
//		p, ok := nc.txBuf[nc.txSeqFirst]
//		if !ok {
//			break
//		} else if size < len(p) {
//			copy(d[copied:], p[:size])
//			nc.txBuf[nc.txSeqFirst] = p[size:]
//		} else {
//			copy(d[copied:], p)
//			nc.txBuf[nc.txSeqFirst] = nil
//			nc.txSeqFirst++
//		}
//	}
//	return copied
//}

//
// implement the NetConn interface
//

var _ types.NetConn = (*NetPoint)(nil)

func (np *NetPoint) Key() string {
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.localAddr.String() + "-" + np.RemoteAddr().String()
}

func (np *NetPoint) Write(d []byte) (int, error) {
	if d == nil || len(d) == 0 {
		return 0, errors.New("invalid buffer")
	}

	remotePoint := np.RemotePoint()
	ret, err := remotePoint.putRX(d)

	//if err == nil {
	//	rb := make([]byte, ret)
	//	go func() {
	//		rlen := remotePoint.getRX(rb)
	//		remotePoint.Worker().OnRecv(remotePoint, rb, rlen)
	//	}()
	//}

	return ret, err
}

func (np *NetPoint) Read(d []byte) (int, error) {
	if d == nil || len(d) == 0 {
		return 0, errors.New("invalid buffer")
	}

	return np.getRX(d), nil
}

func (np *NetPoint) Close() {
	RemoveSession(np.session)

	np.worker.OnClose(np)
	go func(npo *NetPoint) {
		npo.worker.OnClose(npo)
	}(np.remotePoint)
}

func (np *NetPoint) LocalAddr() *net.TCPAddr {
	if np == nil {
		log.Println("break")
	}
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.localAddr
}

func (np *NetPoint) RemoteAddr() *net.TCPAddr {
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.remotePoint.LocalAddr()
}

func (np *NetPoint) LocalIP() net.IP {
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.localAddr.IP
}

func (np *NetPoint) LocalPort() int {
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.localAddr.Port
}

func (np *NetPoint) RemoteIP() net.IP {
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.remotePoint.LocalIP()
}

func (np *NetPoint) RemotePort() int {
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.remotePoint.LocalPort()
}
