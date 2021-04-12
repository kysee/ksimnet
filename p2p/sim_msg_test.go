package p2p

import (
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"net"
	"testing"
)

func TestSimMsgCodec(t *testing.T) {

	m1 := NewAnonySimMsg(NewReqPeers(&net.TCPAddr{IP: net.ParseIP("1.1.1.1"), Port: 12345}))
	require.Equal(t, REQ_PEERS, m1.Type())

	bz, err := m1.Encode()
	require.NoError(t, err)

	m2 := &SimMsg{}
	m2.Decode(bz)

	require.Equal(t, m1, m2)

	log.Println(m1)
	log.Println(m2)

	m2.DstPeerID[0] = 7
	require.NotEqual(t, m1, m2)

	ackPeers := NewAckPeers()
	cnt := rand.Intn(500) + 10
	for i := 0; i < cnt; i++ {
		a := byte(rand.Intn(255) + 1)
		b := byte(rand.Intn(256))
		c := byte(rand.Intn(256))
		d := byte(rand.Intn(255) + 1)
		p := rand.Intn(0xFFFF) + 1

		ackPeers.AddPeerAddr(
			&net.TCPAddr{
				IP:   net.IPv4(a, b, c, d),
				Port: p,
			})
	}

	m3 := NewAnonySimMsg(ackPeers)
	require.Equal(t, ACK_PEERS, m3.Type())

	bz, err = m3.Encode()
	require.NoError(t, err)

	m4 := &SimMsg{}
	m4.Decode(bz)

	require.Equal(t, m3, m4)

	log.Println(m3)
	log.Println(m4)

	amsg := []byte("Test Message")
	m5 := NewAnonySimMsg(NewBytesMsg(amsg))
	require.Equal(t, USER_MSG_TYPE, m5.Type())
	bz, err = m5.Encode()
	require.NoError(t, err)

	m6 := &SimMsg{}
	m6.Decode(bz)

	require.Equal(t, m5, m6)
	require.Equal(t, amsg, m6.Body.(*BytesMsg).Data)

	log.Println(m5)
	log.Println(m6)
}
