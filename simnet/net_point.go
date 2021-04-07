package simnet

import (
	"errors"
	"fmt"
	"github.com/ksimnet/types"
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
		rxCh:       make(chan int, 10),
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

func (nc *NetPoint) Worker() types.NetWorker {
	nc.mtx.RLock()
	defer nc.mtx.RUnlock()

	return nc.worker
}

func (nc *NetPoint) SetSession(sess *Session) {
	nc.mtx.Lock()
	defer nc.mtx.Unlock()

	nc.session = sess
}

func (nc *NetPoint) GetSession() *Session {
	nc.mtx.RLock()
	defer nc.mtx.RUnlock()

	return nc.session
}

func (nc *NetPoint) SetRemotePoint(r *NetPoint) {
	nc.mtx.Lock()
	defer nc.mtx.Unlock()

	nc.remotePoint = r
}

func (nc *NetPoint) RemotePoint() *NetPoint {
	nc.mtx.RLock()
	defer nc.mtx.RUnlock()

	return nc.remotePoint
}

func (nc *NetPoint) putRX(d []byte) (int, error) {
	nc.rxMtx.Lock()

	b := make([]byte, len(d))
	n := copy(b, d)

	seq := nc.rxSeqEnd
	nc.rxBuf[seq] = b
	nc.rxSeqEnd++

	nc.rxMtx.Unlock()

	nc.rxCh <- seq

	return n, nil
}

func (nc *NetPoint) getRX(d []byte) int {
	nc.rxMtx.Lock()
	defer nc.rxMtx.Unlock()

	copied := 0
	size := cap(d)

	for copied < size {
		p, ok := nc.rxBuf[nc.rxSeqFront]
		if !ok {
			break
		} else if size < len(p) {
			copy(d[copied:], p[:size])
			copied = copied + size
			nc.rxBuf[nc.rxSeqFront] = p[size:]
		} else {
			copy(d[copied:], p)
			copied = copied + len(p)
			nc.rxBuf[nc.rxSeqFront] = nil
			nc.rxSeqFront++
		}
	}
	return copied
}

func (nc *NetPoint) pickRX() []byte {
	nc.rxMtx.Lock()
	defer nc.rxMtx.Unlock()

	p, ok := nc.rxBuf[nc.rxSeqFront]
	if !ok {
		fmt.Println(nc.localAddr, "end pickRX")
		return nil
	}
	nc.rxBuf[nc.rxSeqFront] = nil
	nc.rxSeqFront++

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

func (nc *NetPoint) Key() string {
	nc.mtx.RLock()
	defer nc.mtx.RUnlock()

	return nc.localAddr.String() + "-" + nc.RemoteAddr().String()
}

func (nc *NetPoint) Write(d []byte) (int, error) {
	if d == nil || len(d) == 0 {
		return 0, errors.New("invalid buffer")
	}

	remotePoint := nc.RemotePoint()
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

func (nc *NetPoint) Read(d []byte) (int, error) {
	if d == nil || len(d) == 0 {
		return 0, errors.New("invalid buffer")
	}

	return nc.getRX(d), nil
}

func (nc *NetPoint) Close() {
	RemoveSession(nc.session)

	nc.worker.OnClose(nc)
	go func(npo *NetPoint) {
		npo.worker.OnClose(npo)
	}(nc.remotePoint)
}

func (nc *NetPoint) LocalAddr() *net.TCPAddr {
	nc.mtx.RLock()
	defer nc.mtx.RUnlock()

	return nc.localAddr
}

func (nc *NetPoint) RemoteAddr() *net.TCPAddr {
	nc.mtx.RLock()
	defer nc.mtx.RUnlock()

	return nc.remotePoint.LocalAddr()
}
