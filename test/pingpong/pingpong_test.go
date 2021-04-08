package pingpong

import (
	"github.com/kysee/ksimnet/simnet"
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"net"
	"strconv"
	"testing"
)

func TestPingPong(t *testing.T) {
	routine := func(capp *PingClientApp) {
		WaitGrp.Add(1)
		for i := 0; i < testMsgCnt; i++ {
			txt := "client [" + capp.conn.LocalAddr().String() + "] sends a msg:" + strconv.Itoa(i)
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
		hostIp := "192.0.0." + strconv.Itoa(rand.Intn(255)+1)

		capps[i] = &PingClientApp{
			hostIp:  net.ParseIP(hostIp),
			recvBuf: make([]string, testMsgCnt),
			sendBuf: make([]string, testMsgCnt),
		}

		np := simnet.NewNetPoint(capps[i], 0)
		err := np.Connect(serverTcpAddr)
		require.NoError(t, err)
	}

	log.Printf("Run clients...")
	for _, capp := range capps {
		routine(capp)
	}

	log.Printf("Wait ...")
	WaitGrp.Wait()

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

	WaitGrp.Add(clientCnt * 2)

	sapp.Shutdown()

	WaitGrp.Wait()

	require.Zero(t, len(sapp.clients))
}
