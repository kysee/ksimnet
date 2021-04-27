package simnet

import (
	"errors"
	"github.com/kysee/ksimnet/types"
	"net"
	"sync"
	"time"
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
	//rxCh       chan int
	rxCnt int
	txCnt int

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
		//rxCh:       make(chan int, 4084),
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

	//Loop:
	//	for {
	//		select {
	//		case <-nc.rxCh:
	//			nc.AddRxCnt()
	//
	//			//fmt.Println("[",nc.LocalAddr().String(),"] receiveRoutine notified:", seq, "front:", nc.rxSeqFront, "end:", nc.rxSeqEnd, "ch length:", len(nc.rxCh))
	//
	//			if d := nc.pickRX(); d != nil {
	//				if err := nc.Worker().OnRecv(nc, d, len(d)); err != nil {
	//					log.Printf("[receiveRoutine] %v\n", err)
	//				}
	//			} else {
	//				log.Printf("[receiveRoutine] there is no data\n")
	//			}
	//		case <-nc.done:
	//			break Loop
	//		}
	//	}

	ticker := time.NewTicker(time.Microsecond * 500)

Loop:
	for {
		select {
		case <-ticker.C:

			cnt := nc.RxBufCnt()
			//log.Printf("[NetPoint:%s] rx %d\n", nc.LocalIP(), cnt)

			for i := 0; i < cnt; i++ {
				if d := nc.pickRX(); d != nil {
					if err := nc.Worker().OnRecv(nc, d, len(d)); err != nil {
						//log.Printf("[NetPoint:%s receiveRoutine] RxCnt=%d, err=%v\n", nc.LocalIP(), nc.RxCnt(), err)
					}
				}
			}
		case <-nc.done:
			break Loop
		}
		time.Sleep(time.Millisecond)
	}

	gMtx.Lock()
	goroutineNum--
	//log.Printf("[%s] Goodbye... the number of receiveRoutine = %d", nc.Key(), goroutineNum)
	gMtx.Unlock()
}

func (nc *NetPoint) AddRxCnt() {
	nc.mtx.Lock()
	defer nc.mtx.Unlock()

	nc.rxCnt++
}

func (nc *NetPoint) RxCnt() int {
	nc.mtx.RLock()
	defer nc.mtx.RUnlock()

	return nc.rxCnt
}

func (nc *NetPoint) AddTxCnt() {
	nc.mtx.Lock()
	defer nc.mtx.Unlock()

	nc.txCnt++
}

func (nc *NetPoint) TxCnt() int {
	nc.mtx.RLock()
	defer nc.mtx.RUnlock()

	return nc.txCnt
}

func (nc *NetPoint) Close() {
	go func() {
		nc.close()
		nc.remotePoint.close()

		nc.worker.OnClose(nc)
		nc.remotePoint.worker.OnClose(nc.remotePoint)
	}()

	RemoveSession(nc.session)
}

func (nc *NetPoint) close() {
	close(nc.done)
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

	b := make([]byte, len(d))
	n := copy(b, d)

	nc.rxMtx.Lock()

	seq := nc.rxSeqEnd
	nc.rxBuf[seq] = b
	nc.rxSeqEnd++

	nc.rxMtx.Unlock()

	//log.Printf("[NetPoint:%s] writeRX: rxBuf count=%d\n", nc.LocalIP(), nc.RxBufCnt())
	//nc.rxCh <- seq

	return n, nil
}

func (nc *NetPoint) RxBufCnt() int {
	nc.rxMtx.RLock()
	defer nc.rxMtx.RUnlock()

	return nc.rxSeqEnd - nc.rxSeqFront
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

func (nc *NetPoint) pickRX() []byte {
	nc.rxMtx.Lock()
	defer nc.rxMtx.Unlock()

	p, ok := nc.rxBuf[nc.rxSeqFront]
	if !ok {
		//fmt.Println(nc.localAddr, "end pickRX")
		return nil
	}
	nc.rxBuf[nc.rxSeqFront] = nil
	nc.rxSeqFront++

	nc.rxCnt++

	//log.Printf("[NetPoint:%s] pickRX: rxBuf count=%d\n", nc.LocalIP(), nc.RxBufCnt())
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
	return nc.LocalAddr().String() + "-" + nc.RemoteAddr().String()
}

func (nc *NetPoint) Write(d []byte) (int, error) {
	if d == nil || len(d) == 0 {
		return 0, errors.New("invalid buffer")
	}

	ret, err := nc.RemotePoint().putRX(d)

	nc.AddTxCnt()

	return ret, err
}

func (nc *NetPoint) Read(d []byte) (int, error) {
	//if d == nil || len(d) == 0 {
	//	return 0, errors.New("invalid buffer")
	//}
	//
	//return np.getRX(d), nil
	return 0, nil
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

func (nc *NetPoint) LocalIP() net.IP {
	nc.mtx.RLock()
	defer nc.mtx.RUnlock()

	return nc.localAddr.IP
}

func (nc *NetPoint) LocalPort() int {
	nc.mtx.RLock()
	defer nc.mtx.RUnlock()

	return nc.localAddr.Port
}

func (nc *NetPoint) RemoteIP() net.IP {
	//nc.mtx.RLock()
	//defer nc.mtx.RUnlock()

	return nc.RemotePoint().LocalIP()
}

func (nc *NetPoint) RemotePort() int {
	return nc.RemotePoint().LocalPort()
}

var addrMtx sync.Mutex
var ipv4s []net.IP
var ports map[string]int

func ResetIPPorts() {
	addrMtx.Lock()
	defer addrMtx.Unlock()

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

func (nc *NetPoint) Connect(toAddr string) error {
	sess, err := BuildSession(nc, toAddr)
	if err != nil {
		nc.close()
		return err
	}

	err = nc.worker.(types.ClientWorker).OnConnect(nc)
	if err != nil {
		sess.AckCh <- nil
		nc.Close()
		return err
	}

	sess.AckCh <- nil
	return nil
}
