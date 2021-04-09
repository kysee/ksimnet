package p2p

import (
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"net"
	"testing"
)

func TestSimMsgCodec(t *testing.T) {
	m1 := NewReqPeers(1)
	bz, err := m1.Encode()
	require.NoError(t, err)

	m2 := NewReqPeers(2)
	m2.Decode(bz)

	require.Equal(t, m1, m2)

	log.Println(m1)
	log.Println(m2)

	m2.MsgType = 0
	require.NotEqual(t, m1, m2)

	m3 := NewAckPeers(3)

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

	bz, err = m3.Encode()
	require.NoError(t, err)

	m4 := NewAckPeers(0)
	m4.Decode(bz)

	require.Equal(t, m3, m4)

	log.Println(m3)
	log.Println(m4)

	amsg := []byte("Test Message")
	m5 := NewBytesSimMsg(0, 10, amsg)
	bz, err = m5.Encode()
	require.NoError(t, err)

	m6 := &BytesSimMsg{}
	m6.Decode(bz)

	require.Equal(t, m5, m6)
	require.Equal(t, amsg, m6.Body)

	log.Println(m5)
	log.Println(m6)
}
