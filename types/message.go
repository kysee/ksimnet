package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const MsgIDSize = 16 //sha256.Size
const MsgTypeSize = 2

type MsgHeader interface {
	ID() MsgID
	Type() uint16
	Src() PeerID
	Dst() PeerID
}

type MsgBody interface {
	ConstType() uint16
	Encode() ([]byte, error)
	Decode([]byte) error
	String() string
	Hash() []byte
}

type Message interface {
	MsgHeader
	MsgBody
}

type MsgID [MsgIDSize]byte

func CalMsgID(b MsgBody) MsgID {
	var mid MsgID
	copy(mid[:], CalMsgHash(b)[:MsgIDSize])
	return mid
}

func CalMsgHash(b MsgBody) []byte {
	d, err := b.Encode()
	if err != nil {
		panic(err)
	}
	h := sha256.Sum256(d)
	return h[:]
}

func (mid MsgID) String() string {
	return hex.EncodeToString(mid[:])
}

func (mid MsgID) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(mid[:]))
}

func (mid MsgID) UnmarshalJSON(s []byte) error {
	hexStr := ""
	if err := json.Unmarshal(s, &hexStr); err != nil {
		return err
	}

	bz, err := hex.DecodeString(hexStr)
	if err != nil {
		return err
	}
	copy(mid[:], bz)
	return nil
}
