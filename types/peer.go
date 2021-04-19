package types

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type Peer interface {
	ID() PeerID

	Start(int) error
	Send(MsgBody) (int, error)
	SendTo(PeerID, MsgBody) (int, error)
	Stop()

	PeerCnt() int
	HasPeer(PeerID) bool

	ClientWorker
	ServerWorker
}

const PeerIDSize = 32 // 32 bytes == 256 bits
type PeerID [PeerIDSize]byte

func NewPeerID(seed []byte) PeerID {
	if seed == nil {
		return newRandPeerID()
	}

	var r PeerID
	h := sha256.Sum256(seed)
	copy(r[:], h[:])
	return r
}

func newRandPeerID() PeerID {
	var r PeerID
	_, err := rand.Read(r[:])
	if err != nil {
		panic(err)
	}
	return r
}

func NewZeroPeerID() PeerID {
	var r PeerID
	return r
}

func (pid PeerID) Equal(o PeerID) bool {
	return bytes.Equal(pid[:], o[:])
}

func (pid PeerID) String() string {
	return hex.EncodeToString(pid[:])
}

func (pid PeerID) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(pid[:]))
}

func (pid PeerID) UnmarshalJSON(s []byte) error {
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
