package pingpong

import (
	"github.com/kysee/ksimnet/simnet"
	"github.com/kysee/ksimnet/types"
	"net"
	"sync"
)

var WaitGrp *sync.WaitGroup = &sync.WaitGroup{}
var sAddr string = "10.0.0.1:8888"
var sapp *PongServerApp

func init() {

	srvAddr, err := net.ResolveTCPAddr("tcp", sAddr)
	if err != nil {
		panic(err)
	}

	sapp = &PongServerApp{
		clients: make([]types.NetConn, 0),
		recvBuf: make(map[string][]string),
		sendBuf: make(map[string][]string),
	}

	lsn, err := simnet.NewListener(sapp, srvAddr.String())
	if err != nil {
		panic(err)
	}

	if err = lsn.Listen(); err != nil {
		panic(err)
	}

	sapp.listener = lsn
}

type PongServerApp struct {
	recvBufMtx sync.Mutex
	sendBufMtx sync.Mutex
	clients    []types.NetConn
	recvBuf    map[string][]string
	sendBuf    map[string][]string

	listener *simnet.Listener
}

func (s *PongServerApp) Shutdown() {
	s.listener.Shutdown()
}

var _ types.ServerWorker = (*PongServerApp)(nil)

func (s *PongServerApp) OnAccept(conn types.NetConn) error {
	s.clients = append(s.clients, conn)
	//fmt.Println("accept", conn.RemoteAddr())

	return nil
}

func (s *PongServerApp) OnRecv(conn types.NetConn, d []byte, l int) (int, error) {
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

	return l, nil
}

func (s *PongServerApp) OnClose(conn types.NetConn) {
	//log.Printf("Server [%s] is closed\n", conn.Key())
	k := conn.Key()
	for i, c := range s.clients {
		if c.Key() == k {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
	WaitGrp.Done()
}
