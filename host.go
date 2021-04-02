package knetsim

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

type Host struct {
	mtx sync.Mutex
	
	Ip   net.IP
	Port int
}

func NewHost(laddr string, port int) (*Host, error) {
	ip := net.ParseIP(laddr)
	if ip == nil {
		ip = net.IPv4(127, 0, 0, 1)
	}

	h := &Host{
		Ip:   ip,
		Port: port,
	}

	return h, nil
}


func (h *Host) Key() string {
	return addr + ":" + strconv.Itoa(port)
}

func (h *Host) Close() error {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	
	return RemoveServer(h)	
}
