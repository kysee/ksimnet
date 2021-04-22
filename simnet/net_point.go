package simnet

import (
	"errors"
	"fmt"
	"github.com/kysee/ksimnet/types"
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
	rxCnt      int
	txCnt      int

	//txMtx sync.RWMutex
	//txBuf map[int][]byte
	//txSeqFirst int
	//txSeqLast int

	done chan interface{}

	asyncMode bool
}

func NewNetPoint(worker types.NetWorker, hostPort int, asyncMode bool) *NetPoint {
	hostIp := worker.HostIP()
	if hostPort == 0 {
		hostPort = PickPort(hostIp)
	}

	ret := &NetPoint{
		worker:     worker,
		localAddr:  &net.TCPAddr{IP: worker.HostIP(), Port: hostPort},
		rxBuf:      make(map[int][]byte),
		rxSeqFront: 0,
		rxSeqEnd:   0,
		rxCh:       make(chan int, 4084),
		//txBuf: make(map[int][]byte),
		done:      make(chan interface{}),
		asyncMode: asyncMode,
	}

	if asyncMode {
		go receiveRoutine(ret)
	}

	return ret
}

var gMtx sync.Mutex
var goroutineNum = 0

func receiveRoutine(nc *NetPoint) {
	gMtx.Lock()
	goroutineNum++
	//log.Printf("the number of receiveRoutine is %d\n", goroutineNum)
	gMtx.Unlock()

Loop:
	for {
		select {
		case <-nc.rxCh:
			nc.AddRxCnt()

			//fmt.Println("[",nc.LocalAddr().String(),"] receiveRoutine notified:", seq, "front:", nc.rxSeqFront, "end:", nc.rxSeqEnd, "ch length:", len(nc.rxCh))

			if d := nc.pickRX(); d != nil {
				if err := nc.Worker().OnRecv(nc, d, len(d)); err != nil {
					//log.Printf("[receiveRoutine] %v\n", err)
				}

			}
		case <-nc.done:
			break Loop
		}
	}

	gMtx.Lock()
	goroutineNum--
	//log.Printf("[%s] Goodbye... the number of receiveRoutine = %d", nc.Key(), goroutineNum)
	gMtx.Unlock()
}

func (np *NetPoint) AddRxCnt() {
	np.mtx.Lock()
	defer np.mtx.Unlock()

	np.rxCnt++
}

func (np *NetPoint) RxCnt() int {
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.rxCnt
}

func (np *NetPoint) AddTxCnt() {
	np.mtx.Lock()
	defer np.mtx.Unlock()

	np.txCnt++
}

func (np *NetPoint) TxCnt() int {
	np.mtx.RLock()
	defer np.mtx.RUnlock()

	return np.txCnt
}

func (np *NetPoint) Close() {
	go func() {
		np.close()
		np.remotePoint.close()

		np.worker.OnClose(np)
		np.remotePoint.worker.OnClose(np.remotePoint)
	}()

	RemoveSession(np.session)
}

func (np *NetPoint) close() {
	np.done <- struct{}{}
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
	if len(np.rxCh) == cap(np.rxCh) {
		log.Printf("[NetPoint:%s] rxCh=%d/%d\n", np.LocalIP(), len(np.rxCh), cap(np.rxCh))
	}

	return n, nil
}

//func (np *NetPoint) getRX(d []byte) int {
//	np.rxMtx.Lock()
//	defer np.rxMtx.Unlock()
//
//	copied := 0
//	size := cap(d)
//
//	for copied < size {
//		p, ok := np.rxBuf[np.rxSeqFront]
//		if !ok {
//			break
//		} else if size < len(p) {
//			copy(d[copied:], p[:size])
//			copied = copied + size
//			np.rxBuf[np.rxSeqFront] = p[size:]
//		} else {
//			copy(d[copied:], p)
//			copied = copied + len(p)
//			np.rxBuf[np.rxSeqFront] = nil
//			np.rxSeqFront++
//		}
//	}
//	return copied
//}

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

	np.AddTxCnt()

	return ret, err
}

func (np *NetPoint) Read(d []byte) (int, error) {
	//if d == nil || len(d) == 0 {
	//	return 0, errors.New("invalid buffer")
	//}
	//
	//return np.getRX(d), nil
	return 0, nil
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

var addrMtx sync.Mutex
var ipv4s []net.IP
var ports map[string]int

func ResetIPPorts() {
	ipv4s = nil
	ports = make(map[string]int)
}

func IsUsedIp4(ip net.IP) bool {
	addrMtx.Lock()
	defer addrMtx.Unlock()

	for _, _ip := range ipv4s {
		if ip.Equal(_ip) {
			return true
		}
	}
	return false
}

func PickIP() net.IP {
	for i := 0; i < 256*256*256*256; i++ {
		a := byte(1)     //byte(rand.Intn(255) + 1)
		b := byte(0)     //byte(rand.Intn(256))
		c := byte(0)     //byte(rand.Intn(256))
		d := byte(i + 1) //byte(rand.Intn(255) + 1)

		ret := net.IPv4(a, b, c, d)
		if !IsUsedIp4(ret) {
			addrMtx.Lock()
			defer addrMtx.Unlock()

			ipv4s = append(ipv4s, ret)
			return ret
		}
	}
	return nil
}

func PickPort(hostIp net.IP) int {
	addrMtx.Lock()
	defer addrMtx.Unlock()

	host := hostIp.String()
	if p, ok := ports[host]; ok {
		ports[host] = p + 1
		return p + 1
	}
	ports[host] = 1
	return 1
}

func BindPort(hostIp net.IP) *net.TCPAddr {
	tcpAddr := &net.TCPAddr{
		IP:   hostIp,
		Port: PickPort(hostIp),
	}
	return tcpAddr
}

func (np *NetPoint) Connect(toAddr string) error {
	if _, err := BuildSession(np, toAddr); err != nil {
		return err
	}

	if err := np.worker.(types.ClientWorker).OnConnect(np); err != nil {
		np.Close()
		return err
	}
	return nil
}
