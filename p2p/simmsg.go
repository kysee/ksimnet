package p2p

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
)

type SimMsg interface {
	Src() int64
	Type() byte
	Encode() []byte
	Decode(bz []byte) error
	String() string
}

const (
	QUERY byte = 1 + iota
	REPLY
	REQ_PEERS
	ACK_PEERS
)

const ENCODED_HEADER_SIZE = 9

type Header struct {
	SrcId   int64 `json:"src_id"`
	MsgType byte  `json:"type"`
}

func (h *Header) Src() int64 {
	return h.SrcId
}

func (h *Header) Type() byte {
	return h.MsgType
}

func (h *Header) Encode() []byte {
	bz := make([]byte, ENCODED_HEADER_SIZE)
	binary.BigEndian.PutUint64(bz[:8], uint64(h.SrcId))
	bz[8] = h.MsgType
	return bz
}

func (h *Header) Decode(bz []byte) error {
	if len(bz) < ENCODED_HEADER_SIZE {
		return errors.New("wrong message header")
	}
	h.SrcId = (int64)(binary.BigEndian.Uint64(bz[:8]))
	h.MsgType = bz[8]
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

func NewReqPeers(srcId int64) *ReqPeers {
	return &ReqPeers{
		Header: Header{
			SrcId:   srcId,
			MsgType: REQ_PEERS,
		},
	}
}

var _ SimMsg = (*ReqPeers)(nil)

func (m *ReqPeers) Encode() []byte {
	return m.Header.Encode()
}

func (m *ReqPeers) Decode(bz []byte) error {
	return m.Header.Decode(bz)
}

type AckPeers struct {
	Header
	Addrs []*net.TCPAddr `json:"addrs"`
}

func NewAckPeers(srcId int64) *AckPeers {
	return &AckPeers{
		Header: Header{
			SrcId:   srcId,
			MsgType: ACK_PEERS,
		},
	}
}

var _ SimMsg = (*AckPeers)(nil)

func (m *AckPeers) AddPeer(addr *net.TCPAddr) {
	m.Addrs = append(m.Addrs, addr)
}

func (m *AckPeers) Encode() []byte {
	header := m.Header.Encode()

	body := make([]byte, 4+len(m.Addrs)*6) // one addr will be encoded 6byte = ip(4byte) + port(2byte)
	binary.BigEndian.PutUint32(body[:], uint32(len(m.Addrs)))
	pbuf := body[4:]
	for i, a := range m.Addrs {
		ipv4 := a.IP.To4()
		copy(pbuf[i*6:i*6+4], ipv4[:])
		binary.BigEndian.PutUint16(pbuf[i*6+4:i*6+6], uint16(a.Port))
	}

	return append(header, body...)
}

func (m *AckPeers) Decode(bz []byte) error {
	if err := m.Header.Decode(bz); err != nil {
		return err
	}

	body := bz[ENCODED_HEADER_SIZE:]
	addrCnt := binary.BigEndian.Uint32(body[:4])
	m.Addrs = make([]*net.TCPAddr, addrCnt)

	pbuf := body[4:]
	for i := uint32(0); i < addrCnt; i++ {
		ip := pbuf[i*6 : i*6+4]
		m.Addrs[i] = &net.TCPAddr{
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

type ReqQuery struct{}

type AckQuery struct{}
