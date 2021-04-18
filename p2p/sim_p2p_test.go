package p2p_test

import (
	"github.com/kysee/ksimnet/p2p"
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"
)

var PeerCnt = 10
var MinPeerCnt = 6
var MaxPeerCnt = 9
var MsgCnt = 10

func TestSimP2P(t *testing.T) {

	log.Println("Start Seed ...")
	seedPeer := p2p.NewSimSeedPeer(net.IPv4(byte(200), byte(200), byte(200), byte(200)), MinPeerCnt, MaxPeerCnt)
	err := seedPeer.Start(55555)
	require.NoError(t, err)

	log.Println("Create SimPeers...")
	peers := make([]*p2p.SimPeer, PeerCnt)
	for i := 0; i < PeerCnt; i++ {
		peers[i] = p2p.NewSimPeer(net.IPv4(byte(1), byte(1), byte(1), byte(i+2)), MinPeerCnt, MaxPeerCnt, seedPeer.HostAddr())
	}

	log.Println("Start SimPeers and Connecting ...")

	for _, p := range peers {
		err := p.Start(55555)
		require.NoError(t, err)
	}

	for {
		time.Sleep(time.Second * 3)
		totalOthers := 0

		for _, p := range peers {
			n := p.PeerCnt()
			totalOthers += n
		}

		log.Printf("Total others number is %d\n", totalOthers)
		if totalOthers == PeerCnt*(PeerCnt-1) {
			break
		}
	}

	log.Println("Broadcast messages...")

	for i := 0; i < MsgCnt; i++ {
		j := rand.Intn(PeerCnt)
		_, err := peers[j].Send(p2p.NewBytesMsg([]byte("Message Number is " + strconv.Itoa(i))))
		require.NoError(t, err)

		rn := time.Duration(rand.Intn(100) + 1)
		time.Sleep(time.Millisecond * rn)
	}

	for {
		time.Sleep(time.Second * 3)
		totalMsgCnt := 0
		for _, p := range peers {
			n := p.HandledMsgCnt()
			totalMsgCnt += n
		}

		log.Printf("Total messages number is %d\n", totalMsgCnt)
		if totalMsgCnt == MsgCnt*PeerCnt {
			break
		}
	}

	for _, p := range peers {
		p.Stop()
	}

	log.Println("Validate messages...")

	msgIDs := peers[0].HandledMsgIDs()
	for _, p := range peers {
		_msgIDs := p.HandledMsgIDs()
		require.Equal(t, len(msgIDs), len(_msgIDs))
		for k := range msgIDs {
			_, ok := _msgIDs[k]
			require.True(t, ok)
			log.Printf("Peer(%s) had handled the message(%s)\n", p.HostIP(), &k)
		}
	}
}
