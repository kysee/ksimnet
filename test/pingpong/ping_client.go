package pingpong

import "github.com/ksimnet/netconn"

type PingClientApp struct{
	conn netconn.NetConn

	recvSeq int
	recvBuf []string
	sendSeq int
	sendBuf []string
}
var _ netconn.ClientWorker = (*PingClientApp)(nil)

func (c *PingClientApp) OnConnect(conn *netconn.NetPoint) error {
	panic("implement me")
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

