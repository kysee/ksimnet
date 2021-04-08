package simnet2

import (
	"github.com/ksimnet/simnet2/types"
	"sync"
)

type Host struct {
	mtx sync.RWMutex

	ifset    []*types.NetInf
	tcpPorts map[int]int
}

func NewHost(ips ...string) *Host {
	return nil
}
