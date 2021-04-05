package netconn

import (
	"net"
	"sync"
)

type NetConn struct {
	localAddr  *net.TCPAddr
	remoteConn *NetConn

	rxMtx      sync.RWMutex
	rxBuf      map[int][]byte
	rxSeqFirst int
	rxSeqLast  int
	rxCh       chan int

	//txMtx sync.RWMutex
	//txBuf map[int][]byte
	//txSeqFirst int
	//txSeqLast int

}

func NewNetConn(localAddr *net.TCPAddr) (*NetConn, error) {
	return &NetConn{
		localAddr: localAddr,
		rxBuf:     make(map[int][]byte),
		rxCh:      make(chan int),
		//txBuf: make(map[int][]byte),
	}, nil
}

func (nc *NetConn) LocalAddr() *net.TCPAddr {
	return nc.localAddr
}

func (nc *NetConn) RemoteAddr() *net.TCPAddr {
	return nc.remoteConn.localAddr
}

func (nc *NetConn) SetRemoteConn(r *NetConn) {
	nc.remoteConn = r
}

func (nc *NetConn) RemoteConn() *NetConn {
	return nc.remoteConn
}

func (nc *NetConn) RX_Put(d []byte) (int, error) {
	nc.rxMtx.Lock()
	defer nc.rxMtx.Unlock()

	b := make([]byte, len(d))
	n := copy(b, d)

	seq := len(nc.rxBuf)
	nc.rxBuf[seq] = b
	nc.rxSeqLast = seq

	nc.rxCh <- seq
	return n, nil
}

func (nc *NetConn) RX_Get(d []byte) int {
	nc.rxMtx.Lock()
	defer nc.rxMtx.Unlock()

	copied := 0
	size := cap(d)

	for copied < size {
		p, ok := nc.rxBuf[nc.rxSeqFirst]
		if !ok {
			<-nc.rxCh
		} else if size < len(p) {
			copy(d[copied:], p[:size])
			copied = copied + size
			nc.rxBuf[nc.rxSeqFirst] = p[size:]
		} else {
			copy(d[copied:], p)
			copied = copied + len(p)
			nc.rxBuf[nc.rxSeqFirst] = nil
			nc.rxSeqFirst++
		}
	}
	return copied
}

//func (nc *NetConn) TX_Put(d []byte) int {
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
//func (nc *NetConn) TX_Get(d []byte) int {
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
