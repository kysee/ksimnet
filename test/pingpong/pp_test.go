package pingpong

import (
	"github.com/ksimnet/netconn"
	"github.com/ksimnet/simnet"
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"testing"
)


var clientCnt = 200
var testMsgCnt = 10000

var done chan<- struct{}
var sapp *PongServerApp
var WaitGrp *sync.WaitGroup

func init() {

	srvAddr, err := net.ResolveTCPAddr("tcp", sAddr)
	if err != nil {
		panic(err)
	}

	sapp = &PongServerApp{
		clients: make([]*netconn.NetPoint, 0),
		recvBuf: make(map[string][]string),
		sendBuf: make(map[string][]string),
	}

	s, err := simnet.NewServer(sapp, srvAddr.String())
	if err != nil {
		panic(err)
	}

	done, err = s.Listen()

	if err != nil {
		panic(err)
	}

	WaitGrp = &sync.WaitGroup{}
}

func TestConnect(t *testing.T) {
	routine := func(capp *PingClientApp) {
		WaitGrp.Add(1)
		for i := 0; i< testMsgCnt; i++ {
			txt := "client [" + capp.conn.LocalAddr().String() + "] sends a msg:" +strconv.Itoa(i)
			_, err := capp.conn.Write([]byte(txt))
			require.NoError(t, err)

			capp.sendBuf[capp.sendSeq] = txt
			capp.sendSeq++

			//res := make([]byte, len(txt) + 5)
			//_, err = c.Read(res)
			//require.NoError(t, err)
		}
	}

	log.Printf("Create clients...")
	capps := make([]*PingClientApp, clientCnt)
	for i, _ := range capps {
		capps[i] = &PingClientApp{
			recvBuf: make([]string, testMsgCnt),
			sendBuf: make([]string, testMsgCnt),
		}

		hostIp := "192.0.0."+strconv.Itoa(rand.Intn(255)+1)
		conn, err := simnet.Connect(capps[i], hostIp, sAddr)
		require.NoError(t, err)

		capps[i].conn = conn
	}

	log.Printf("Run clients...")
	for _, capp := range capps {
		routine(capp)
	}

	log.Printf("Wait ...")
	WaitGrp.Wait()

	done <- struct{}{}

	log.Printf("Validate ...")

	for _, capp := range capps {
		//fmt.Println("Validating for", capp.conn.LocalAddr())
		// client send == server received
		clientSendBuf := capp.sendBuf
		serverRecvBuf, ok := sapp.recvBuf[capp.conn.LocalAddr().String()]
		require.True(t, ok)
		require.Equal(t, testMsgCnt, len(serverRecvBuf))
		require.Equal(t, len(clientSendBuf), len(serverRecvBuf))

		for i, m := range serverRecvBuf {
			require.Equal(t, m, clientSendBuf[i])
			//if i > 0 && i % 6000 == 0 {
			//	fmt.Println("validating...", m, "==", serverRecvBuf[i])
			//}
		}

		// clientRecv == serverSend
		clientRecvBuf := capp.recvBuf
		serverSendBuf, ok := sapp.sendBuf[capp.conn.LocalAddr().String()]
		require.True(t, ok)
		require.Equal(t, testMsgCnt, len(serverSendBuf))
		require.Equal(t, len(clientRecvBuf), len(serverSendBuf))

		for i, m := range serverSendBuf {
			require.Equal(t, m, clientRecvBuf[i])
			//if i > 0 && i % 6000 == 0 {
			//	fmt.Println("validating...", m, "==", serverSendBuf[i])
			//}
		}
	}
}