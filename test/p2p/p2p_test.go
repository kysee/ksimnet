package p2p_test

import (
	"fmt"
	"github.com/kysee/ksimnet/test/p2p"
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"
)

var PeerCnt = 100
var MsgCnt = 1000

func TestP2P(t *testing.T) {

	log.Println("Create Peers...")
	peers := make([]*p2p.Peer, PeerCnt)
	for i := 0; i < PeerCnt; i++ {
		peers[i] = p2p.NewPeer(net.IPv4(byte(1), byte(1), byte(1), byte(i+1)).String() + ":55555")
	}

	log.Println("Start Peers and Connecting ...")

	p2p.WG.Add(PeerCnt)

	for _, p := range peers {
		p.Start()
	}

	p2p.WG.Wait()

	log.Println("Broadcast messages...")

	p2p.WG.Add(PeerCnt * MsgCnt)

	for i := 0; i < MsgCnt; i++ {
		j := rand.Intn(PeerCnt)
		peers[j].Broadcast(uint64(i), []byte("Message Number is "+strconv.Itoa(i)), nil)

		rn := time.Duration(rand.Intn(100) + 1)
		time.Sleep(time.Millisecond * rn)
	}

	p2p.WG.Wait()

	for _, p := range peers {
		ip := p.LocalIP()
		c := p.RecvMsgCnt()
		log.Printf("Validating for %d messages of the peer(%s)\n", c, ip)

		require.Equal(t, MsgCnt, c, fmt.Sprintf("peer(%s) has %d messages.", ip, c))
		for i := 0; i < c; i++ {
			m := p.RecvMsg(uint64(i))
			require.Equal(t, []byte("Message Number is "+strconv.Itoa(i)), m)
		}
	}
}
