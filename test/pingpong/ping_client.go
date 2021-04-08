package pingpong

import (
	"github.com/kysee/ksimnet/types"
	"net"
)

var clientCnt = 200
var testMsgCnt = 10000

type PingClientApp struct {
	hostIp net.IP
	conn   types.NetConn

	recvSeq int
	recvBuf []string
	sendSeq int
	sendBuf []string
}

var _ types.ClientWorker = (*PingClientApp)(nil)

func (c *PingClientApp) HostIP() net.IP {
	return c.hostIp
}

func (c *PingClientApp) OnConnect(conn types.NetConn) {
	c.conn = conn
}

func (c *PingClientApp) OnRecv(conn types.NetConn, d []byte, l int) (int, error) {
	//log.Printf("Client %s received '%s' from %s\n", conn.LocalAddr(), string(d), conn.RemoteAddr())

	c.recvBuf[c.recvSeq] = string(d)
	c.recvSeq++

	if c.recvSeq == testMsgCnt {
		WaitGrp.Done()
	}
	return l, nil
}

func (c *PingClientApp) OnClose(conn types.NetConn) {
	//log.Printf("Client [%s] is closed\n", conn.Key())
	WaitGrp.Done()
}
