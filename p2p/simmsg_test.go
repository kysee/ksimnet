package p2p

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"net"
	"testing"
)

func TestSimMsgCodec(t *testing.T) {
	m1 := NewReqPeers(10)
	bz := m1.Encode()

	m2 := NewReqPeers(0)
	m2.Decode(bz)

	require.Equal(t, m1, m2)

	log.Println(m1)
	log.Println(m2)

	m3 := NewAckPeers(20)

	cnt := rand.Intn(5) + 10
	for i := 0; i < cnt; i++ {
		a := 1  //rand.Intn(255) + 1
		b := 1  //rand.Intn(255) + 1
		c := 1  //rand.Intn(255) + 1
		d := i  //rand.Intn(255) + 1
		p := 22 //rand.Intn(65567) + 1

		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%d.%d.%d.%d:%d", a, b, c, d, p))
		require.NoError(t, err)
		m3.AddPeer(addr)
	}

	bz = m3.Encode()
	m4 := NewAckPeers(0)
	m4.Decode(bz)

	require.Equal(t, m3, m4)

	log.Println(m3)
	log.Println(m4)

}
