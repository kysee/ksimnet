package simnet2

import (
	"errors"
	"github.com/ksimnet/simnet2/types"
	"net"
	"sync"
)

type IPNetInf struct {
	mtx sync.RWMutex

	ip      net.IP
	remotes map[string]*IPNetInf
}

func NewIPNetInf(ip net.IP) *IPNetInf {
	return &IPNetInf{}
}

func (ipnf *IPNetInf) Key() string {
	return ipnf.ip.String()
}

func (ipnf *IPNetInf) OnRx(from types.NetInf, d []byte) error {
	panic("implement me")
}

func (ipnf *IPNetInf) DoTx(to types.NetInf, d []byte) error {
	if rif, ok := ipnf.remotes[to.Key()]; ok {
		return rif.OnRx(ipnf, d)
	}
	return errors.New("not found connection: " + to.Key())
}

var _ types.NetInf = (*IPNetInf)(nil)
