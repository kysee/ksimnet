package p2p

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
)

type SimMsg interface {
	ID() uint64
	Type() uint16

	Encode() ([]byte, error)
	Decode([]byte) error
	String() string
}

const (
	QUERY uint16 = 1 + iota
	REPLY
	REQ_PEERS
	ACK_PEERS
	USER_MSG_TYPE
)

const HeaderSize = 8 + 2

type Header struct {
	Id      uint64 `json:"id"`
	MsgType uint16 `json:"type"`
}

func (h *Header) ID() uint64 {
	return h.Id
}

func (h *Header) Type() uint16 {
	return h.MsgType
}

func (h *Header) Encode(buf []byte) error {
	if len(buf) < HeaderSize {
		return errors.New("too small buffer")
	}

	binary.BigEndian.PutUint64(buf, h.Id)
	binary.BigEndian.PutUint16(buf[binary.Size(h.Id):], h.MsgType)
	return nil
}

func (h *Header) Decode(buf []byte) error {
	if len(buf) < HeaderSize {
		return errors.New("too small buffer")
	}
	h.Id = binary.BigEndian.Uint64(buf)
	h.MsgType = binary.BigEndian.Uint16(buf[binary.Size(h.Id):])
	return nil
}

func (h *Header) String() string {
	bz, err := json.Marshal(h)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}

type ReqPeers struct {
	Header
}

func NewReqPeers(id uint64) *ReqPeers {
	return &ReqPeers{
		Header: Header{
			Id:      id,
			MsgType: REQ_PEERS,
		},
	}
}

var _ SimMsg = (*ReqPeers)(nil)

func (m *ReqPeers) Encode() ([]byte, error) {
	buf := make([]byte, HeaderSize)
	if err := m.Header.Encode(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

type AckPeers struct {
	Header
	Addrs []net.TCPAddr `json:"addrs"`
}

func NewAckPeers(id uint64) *AckPeers {
	return &AckPeers{
		Header: Header{
			Id:      id,
			MsgType: ACK_PEERS,
		},
	}
}

var _ SimMsg = (*AckPeers)(nil)

func (m *AckPeers) AddPeer(addr *net.TCPAddr) {
	m.Addrs = append(m.Addrs, *addr)
}

func (m *AckPeers) Encode() ([]byte, error) {
	// one addr will be encoded 6byte = ip(4byte) + port(2byte)
	buf := make([]byte, HeaderSize+4+len(m.Addrs)*6)

	if err := m.Header.Encode(buf); err != nil {
		return nil, err
	}

	body := buf[HeaderSize:]
	binary.BigEndian.PutUint32(body[:], uint32(len(m.Addrs)))
	pbuf := body[4:]
	for i, a := range m.Addrs {
		ipv4 := a.IP.To4()
		copy(pbuf[i*6:i*6+4], ipv4[:])
		binary.BigEndian.PutUint16(pbuf[i*6+4:i*6+6], uint16(a.Port))
	}

	return buf, nil
}

func (m *AckPeers) Decode(buf []byte) error {
	if err := m.Header.Decode(buf); err != nil {
		return err
	}

	body := buf[HeaderSize:]
	addrCnt := binary.BigEndian.Uint32(body[:4])
	m.Addrs = make([]net.TCPAddr, addrCnt)

	pbuf := body[4:]
	for i := uint32(0); i < addrCnt; i++ {
		ip := pbuf[i*6 : i*6+4]
		m.Addrs[i] = net.TCPAddr{
			IP:   net.IPv4(ip[0], ip[1], ip[2], ip[3]),
			Port: int(binary.BigEndian.Uint16(pbuf[i*6+4 : i*6+6])),
		}
	}
	return nil
}

func (m *AckPeers) String() string {
	bz, err := json.Marshal(m)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}

type BytesSimMsg struct {
	Header
	Body []byte `json:"body"`
}

var _ SimMsg = (*BytesSimMsg)(nil)

func NewBytesSimMsg(id uint64, mtype uint16, data []byte) *BytesSimMsg {
	return &BytesSimMsg{
		Header: Header{
			Id:      id,
			MsgType: mtype,
		},
		Body: data,
	}
}

func (m *BytesSimMsg) SetType(ty uint16) {
	m.Header.MsgType = ty
}

func (m *BytesSimMsg) SetBody(body []byte) {
	m.Body = body
}

func (m *BytesSimMsg) Encode() ([]byte, error) {
	buf := make([]byte, HeaderSize+len(m.Body))

	if err := m.Header.Encode(buf); err != nil {
		return nil, err
	}

	body := buf[HeaderSize:]
	copy(body, m.Body)

	return buf, nil
}

func (m *BytesSimMsg) Decode(buf []byte) error {
	if err := m.Header.Decode(buf); err != nil {
		return err
	}

	m.Body = make([]byte, len(buf)-HeaderSize)
	copy(m.Body, buf[HeaderSize:])

	return nil
}

func (m *BytesSimMsg) MarshalJSON() ([]byte, error) {
	type Alias BytesSimMsg
	return json.Marshal(&struct {
		Body string `json:"body"`
		*Alias
	}{
		Body:  hex.EncodeToString(m.Body),
		Alias: (*Alias)(m),
	})
}

func (m *BytesSimMsg) UnmarshalJSON(bz []byte) error {
	type Alias BytesSimMsg
	alias := &struct {
		Body string `json:"body"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	err := json.Unmarshal(bz, alias)
	if err != nil {
		return err
	}

	m.Body, err = hex.DecodeString(alias.Body)
	if err != nil {
		return err
	}
	return nil
}

func (m *BytesSimMsg) String() string {
	bz, err := json.Marshal(m)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}

type StrSimMsg struct {
	Header
	Body string `json:"body"`
}

func NewStrSimMsg(id uint64, mtype uint16, data string) *StrSimMsg {
	return &StrSimMsg{
		Header: Header{
			Id:      id,
			MsgType: mtype,
		},
		Body: data,
	}
}

func (m *StrSimMsg) String() string {
	bz, err := json.Marshal(m)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}
