package pingpong

import (
	"github.com/ksimnet/netconn"
	"sync"
)

var clientCnt = 200
var testMsgCnt = 10000
var WaitGrp *sync.WaitGroup

type PingClientApp struct {
	conn netconn.NetConn

	recvSeq int
	recvBuf []string
	sendSeq int
	sendBuf []string
}

var _ netconn.ClientWorker = (*PingClientApp)(nil)

func (c *PingClientApp) OnConnect(conn *netconn.NetPoint) error {
	c.conn = conn
	return nil
}

func (c *PingClientApp) OnRecv(conn *netconn.NetPoint, d []byte, l int) (int, error) {
	c.recvBuf[c.recvSeq] = string(d)
	c.recvSeq++

	if c.recvSeq == testMsgCnt {
		WaitGrp.Done()
	}
	return l, nil
}

func (c *PingClientApp) OnClose(conn *netconn.NetPoint) error {
	panic("implement me")
}
