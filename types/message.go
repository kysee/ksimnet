package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const MsgIDSize = 16 //sha256.Size
const MsgTypeSize = 2

type Message interface {
	ID() MsgID
	Type() uint16
	Src() PeerID
	Dst() PeerID

	Encode() ([]byte, error)
	Decode([]byte) error
	String() string
}

func NewMsgID(d []byte) MsgID {
	h := sha256.Sum256(d)

	var mid MsgID
	copy(mid[:], h[:MsgIDSize])
	return mid
}

type MsgID [MsgIDSize]byte

func (mid *MsgID) String() string {
	return hex.EncodeToString(mid[:])
}

func (mid *MsgID) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(mid[:]))
}

func (mid *MsgID) UnmarshalJSON(s []byte) error {
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
