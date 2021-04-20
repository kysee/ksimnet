package pingpong

import (
	"github.com/kysee/ksimnet/simnet"
	"github.com/kysee/ksimnet/types"
	"net"
	"strconv"
	"sync"
)

var WaitGrp *sync.WaitGroup = &sync.WaitGroup{}
var serverIp = "10.0.0.1"
var serverPort = 8888
var serverTcpAddr string = "10.0.0.1:" + strconv.Itoa(serverPort)
var sapp *PongServerApp

func init() {

	sapp = &PongServerApp{
		hostIp:  net.ParseIP(serverIp),
		clients: make([]types.NetConn, 0),
		recvBuf: make(map[string][]string),
		sendBuf: make(map[string][]string),
	}

	sapp.listener = simnet.NewListener(sapp)
	if err := sapp.listener.Listen(serverPort); err != nil {
		panic(err)
	}
}

type PongServerApp struct {
	recvBufMtx sync.Mutex
	sendBufMtx sync.Mutex
	clients    []types.NetConn
	recvBuf    map[string][]string
	sendBuf    map[string][]string

	listener *simnet.Listener
	hostIp   net.IP
}

func (s *PongServerApp) Shutdown() {
	s.listener.Shutdown()
}

var _ types.ServerWorker = (*PongServerApp)(nil)

func (s *PongServerApp) HostIP() net.IP {
	return s.hostIp
}

func (s *PongServerApp) OnAccept(conn types.NetConn) error {
	s.clients = append(s.clients, conn)
	//fmt.Println("accept", conn.RemoteAddr())

	return nil
}

func (s *PongServerApp) OnRecv(conn types.NetConn, d []byte, l int) error {
	//log.Printf("Listener %s received '%s' from %s\n", conn.LocalAddr(), string(d), conn.RemoteAddr())

	s.recvBufMtx.Lock()
	serverRecv, ok := s.recvBuf[conn.RemoteAddr().String()]
	if !ok {
		serverRecv = make([]string, 0, testMsgCnt*clientCnt)
	}
	serverRecv = serverRecv[:len(serverRecv)+1]
	serverRecv[len(serverRecv)-1] = string(d)
	s.recvBuf[conn.RemoteAddr().String()] = serverRecv
	s.recvBufMtx.Unlock()

	resp := "response for " + string(d)
	conn.Write([]byte(resp))

	s.sendBufMtx.Lock()
	serverSend, ok := s.sendBuf[conn.RemoteAddr().String()]
	if !ok {
		serverSend = make([]string, 0, testMsgCnt*clientCnt)
	}
	serverSend = serverSend[:len(serverSend)+1]
	serverSend[len(serverSend)-1] = resp
	s.sendBuf[conn.RemoteAddr().String()] = serverSend
	s.sendBufMtx.Unlock()

	return nil
}

func (s *PongServerApp) OnClose(conn types.NetConn) error {
	//log.Printf("Server [%s] is closed\n", conn.Key())
	k := conn.Key()
	for i, c := range s.clients {
		if c.Key() == k {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
	WaitGrp.Done()
	return nil
}
