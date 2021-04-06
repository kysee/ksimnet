package pingpong

import (
	"github.com/ksimnet/netconn"
)


var sAddr = "10.0.0.1:8888"

type PongServerApp struct {
	clients []*netconn.NetPoint
	recvBuf map[string][]string
	sendBuf map[string][]string
}
var _ netconn.ServerWorker = (*PongServerApp)(nil)

func (s *PongServerApp) OnAccept(conn *netconn.NetPoint) error {
	s.clients = append(s.clients, conn)
	//fmt.Println("accept", conn.RemoteAddr())
	return nil
}

func (s *PongServerApp) OnRecv(conn *netconn.NetPoint, d []byte, l int)  (int, error) {
	serverRecv, ok := s.recvBuf[conn.RemoteAddr().String()]
	if !ok {
		serverRecv = make([]string, 0, testMsgCnt*clientCnt)
	}
	serverRecv = serverRecv[:len(serverRecv)+1]
	serverRecv[len(serverRecv)-1] = string(d)

	resp := "response for "+string(d)
	conn.Write([]byte(resp))

	serverSend, ok := s.sendBuf[conn.RemoteAddr().String()]
	if !ok {
		serverSend = make([]string, 0, testMsgCnt*clientCnt)
	}
	serverSend = serverSend[:len(serverSend)+1]
	serverSend[len(serverSend)-1] = resp

	s.recvBuf[conn.RemoteAddr().String()] = serverRecv
	s.sendBuf[conn.RemoteAddr().String()] = serverSend

	return l, nil
}

func (s *PongServerApp) OnClose(conn *netconn.NetPoint) error {
	panic("implement me")
}




