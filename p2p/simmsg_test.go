package p2p

import (
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

	m2.MsgType = 0
	require.NotEqual(t, m1, m2)

	m3 := NewAckPeers(20)

	cnt := rand.Intn(500) + 10
	for i := 0; i < cnt; i++ {
		a := byte(rand.Intn(255) + 1)
		b := byte(rand.Intn(256))
		c := byte(rand.Intn(256))
		d := byte(rand.Intn(255) + 1)
		p := rand.Intn(0xFFFF) + 1

		m3.AddPeer(
			&net.TCPAddr{
				IP:   net.IPv4(a, b, c, d),
				Port: p,
			})
	}

	bz = m3.Encode()
	m4 := NewAckPeers(0)
	m4.Decode(bz)

	require.Equal(t, m3, m4)

	log.Println(m3)
	log.Println(m4)

}
