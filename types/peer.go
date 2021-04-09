package types

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
)

type Peer interface {
	ID() PeerID

	Start(listenPort int) error
	Send(d []byte) (int, error)
	Stop()

	PeerCnt() int
	HasPeer(id PeerID) bool

	ClientWorker
	ServerWorker
}

const PeerIDSize = 32 // 32 bytes == 256 bits
type PeerID [PeerIDSize]byte

func NewPeerID() PeerID {
	var r PeerID
	_, err := rand.Read(r[:])
	if err != nil {
		panic(err)
	}
	return r
}

func (pid *PeerID) Equal(o PeerID) bool {
	return pid.Equal(o)
}

func (pid *PeerID) String() string {
	return hex.EncodeToString(pid[:])
}

func (pid *PeerID) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(pid[:]))
}

func (pid *PeerID) UnmarshalJSON(s []byte) error {
	hexStr := ""
	if err := json.Unmarshal(s, &hexStr); err != nil {
		return err
	}

	bz, err := hex.DecodeString(hexStr)
	if err != nil {
		return err
	}
	copy(pid[:], bz)
	return nil
}
