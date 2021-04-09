package p2p

import (
	"net"
	"sync"
)

type AddrBook struct {
	mtx   sync.RWMutex
	addrs []*net.TCPAddr
}

func NewAddrBook() *AddrBook {
	return &AddrBook{}
}

func (ab *AddrBook) Add(addr *net.TCPAddr) {
	ab.mtx.Lock()
	defer ab.mtx.Unlock()

	ab.addrs = append(ab.addrs, addr)
}

func (ab *AddrBook) Find(addr *net.TCPAddr) *net.TCPAddr {
	ab.mtx.RLock()
	defer ab.mtx.RUnlock()

	for _, a := range ab.addrs {
		if EqualAddr(a, addr) {
			return a
		}
	}
	return nil
}

func (ab *AddrBook) Addrs() []net.TCPAddr {
	ab.mtx.RLock()
	defer ab.mtx.RUnlock()

	ret := make([]net.TCPAddr, len(ab.addrs))
	for i, v := range ab.addrs {
		ret[i] = *v
	}
	return ret
}

func EqualAddr(a1, a2 *net.TCPAddr) bool {
	return (a1.IP.Equal(a2.IP) && a1.Port == a2.Port && a1.Zone == a2.Zone)
}
