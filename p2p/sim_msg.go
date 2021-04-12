package p2p

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/kysee/ksimnet/types"
	"net"
)

const (
	QUERY uint16 = 1 + iota
	REPLY
	REQ_PEERS
	ACK_PEERS
	USER_MSG_TYPE
)

type SimMsg struct {
	Header `json:"header"`
	Body   interface{} `json:"body"`
}

func NewAnonySimMsg(body interface{}) *SimMsg {
	return NewSimMsg(types.NewZeroPeerID(), types.NewZeroPeerID(), body)
}
func NewSimMsg(src, dst types.PeerID, body interface{}) *SimMsg {
	var mtype uint16
	var bzBody []byte
	var err error

	switch body.(type) {
	case *ReqPeers:
		mtype = REQ_PEERS
		bzBody, err = body.(*ReqPeers).Encode()
	case *AckPeers:
		mtype = ACK_PEERS
		bzBody, err = body.(*AckPeers).Encode()
	case *BytesMsg:
		mtype = USER_MSG_TYPE
		bzBody, err = body.(*BytesMsg).Encode()
	default:
		panic("unknown message type")
	}

	if err != nil {
		return nil
	}

	return &SimMsg{
		Header: Header{
			MsgID:     types.NewMsgID(bzBody),
			MsgType:   mtype,
			SrcPeerID: src,
			DstPeerID: dst,
		},
		Body: body,
	}
}

var _ types.Message = (*SimMsg)(nil)

func (sm *SimMsg) Encode() ([]byte, error) {

	header, err := sm.Header.Encode()
	if err != nil {
		return nil, err
	}

	var body []byte

	switch sm.Body.(type) {
	case *ReqPeers:
		body, err = sm.Body.(*ReqPeers).Encode()
	case *AckPeers:
		body, err = sm.Body.(*AckPeers).Encode()
	case *BytesMsg:
		body, err = sm.Body.(*BytesMsg).Encode()
	default:
		panic("unknown message type")
	}

	return append(header, body...), nil
}

func (sm *SimMsg) Decode(buf []byte) error {
	if err := sm.Header.Decode(buf); err != nil {
		return err
	}

	switch sm.Header.MsgType {
	case REQ_PEERS:
		m := &ReqPeers{}
		if err := m.Decode(buf[HeaderSize:]); err != nil {
			return err
		}
		sm.Body = m
	case ACK_PEERS:
		m := &AckPeers{}
		if err := m.Decode(buf[HeaderSize:]); err != nil {
			return err
		}
		sm.Body = m
	case USER_MSG_TYPE:
		m := &BytesMsg{}
		if err := m.Decode(buf[HeaderSize:]); err != nil {
			return err
		}
		sm.Body = m
	default:
		panic("unknown message type")
	}
	return nil
}

func (sm *SimMsg) String() string {
	bz, err := json.Marshal(sm)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}

const HeaderSize = types.MsgIDSize + types.MsgTypeSize + types.PeerIDSize*2

type Header struct {
	MsgID     types.MsgID  `json:"msg_id"`
	MsgType   uint16       `json:"msg_type"`
	SrcPeerID types.PeerID `json:"src"`
	DstPeerID types.PeerID `json:"dst,omitempty"`
}

func (h *Header) ID() types.MsgID {
	return h.MsgID
}

func (h *Header) SetType(ty uint16) {
	h.MsgType = ty
}

func (h *Header) Type() uint16 {
	return h.MsgType
}

func (h *Header) SetSrc(peerId types.PeerID) {
	h.SrcPeerID = peerId
}

func (h *Header) Src() types.PeerID {
	return h.SrcPeerID
}

func (h *Header) SetDst(peerId types.PeerID) {
	h.DstPeerID = peerId
}

func (h *Header) Dst() types.PeerID {
	return h.DstPeerID
}

func (h *Header) Encode() ([]byte, error) {

	buf := make([]byte, HeaderSize)

	i := 0
	copy(buf[i:], h.MsgID[:])
	i += types.MsgIDSize
	binary.BigEndian.PutUint16(buf[i:], h.MsgType)
	i += types.MsgTypeSize
	copy(buf[i:], h.SrcPeerID[:])
	i += types.PeerIDSize
	copy(buf[i:], h.DstPeerID[:])
	return buf, nil
}

func (h *Header) Decode(buf []byte) error {
	if len(buf) < HeaderSize {
		return errors.New("too small buffer")
	}
	s, e := 0, types.MsgIDSize
	copy(h.MsgID[:], buf[s:e])

	s, e = e, e+types.MsgTypeSize
	h.MsgType = binary.BigEndian.Uint16(buf[s:e])

	s, e = e, e+types.PeerIDSize
	copy(h.SrcPeerID[:], buf[s:e])

	s, e = e, e+types.PeerIDSize
	copy(h.DstPeerID[:], buf[s:e])

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
	//Header
	ExportAddr *net.TCPAddr `json:"export_addr"`
}

func NewReqPeers(exAddr *net.TCPAddr) *ReqPeers {
	return &ReqPeers{
		ExportAddr: exAddr,
	}
}

const EncodedTCPAddrSize = 6

func encodeTcpAddr(addr *net.TCPAddr) []byte {
	buf := make([]byte, 6)
	ipv4 := addr.IP.To4()
	copy(buf[:], ipv4[:])
	binary.BigEndian.PutUint16(buf[4:], uint16(addr.Port))
	return buf
}

func decodeTcpAddr(d []byte) *net.TCPAddr {
	return &net.TCPAddr{
		IP:   net.IPv4(d[0], d[1], d[2], d[3]),
		Port: int(binary.BigEndian.Uint16(d[4:EncodedTCPAddrSize])),
	}
}

func (m *ReqPeers) Encode() ([]byte, error) {

	buf := make([]byte, EncodedTCPAddrSize)
	copy(buf, encodeTcpAddr(m.ExportAddr))

	return buf, nil
}

func (m *ReqPeers) Decode(buf []byte) error {

	m.ExportAddr = decodeTcpAddr(buf)
	return nil
}

func (m *ReqPeers) String() string {
	bz, err := json.Marshal(m)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}

type AckPeers struct {
	//Header
	Addrs []*net.TCPAddr `json:"addrs"`
}

func NewAckPeers() *AckPeers {
	return &AckPeers{
		//Header: Header{
		//	MsgType: ACK_PEERS,
		//},
	}
}

//var _ types.Message = (*AckPeers)(nil)

func (m *AckPeers) AddPeerAddr(addr *net.TCPAddr) {
	m.Addrs = append(m.Addrs, addr)
}

func (m *AckPeers) Encode() ([]byte, error) {
	// buf = [Header + Addr Size + Addr1 + Addr2 + ... + Addr N]
	//buf := make([]byte, HeaderSize+4+len(m.Addrs)*EncodedTCPAddrSize)

	//if err := m.Header.Encode(buf); err != nil {
	//	return nil, err
	//}

	// [Addr Size + Addr1 + Addr2 + ... + Addr N]
	buf := make([]byte, 4+len(m.Addrs)*EncodedTCPAddrSize) //buf[HeaderSize:]
	binary.BigEndian.PutUint32(buf[:], uint32(len(m.Addrs)))
	tcp_addr_bufs := buf[4:]
	for i, a := range m.Addrs {
		copy(tcp_addr_bufs[i*EncodedTCPAddrSize:], encodeTcpAddr(a))
	}

	return buf, nil
}

func (m *AckPeers) Decode(buf []byte) error {
	//if err := m.Header.Decode(buf); err != nil {
	//	return err
	//}

	//body := buf[HeaderSize:]
	addrCnt := binary.BigEndian.Uint32(buf[:4]) // addr count
	m.Addrs = make([]*net.TCPAddr, addrCnt)

	tcp_addr_bufs := buf[4:]
	for i := uint32(0); i < addrCnt; i++ {
		m.Addrs[i] = decodeTcpAddr(tcp_addr_bufs[i*EncodedTCPAddrSize:])
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

type BytesMsg struct {
	//Header
	Data []byte `json:"data"`
}

//var _ types.Message = (*BytesMsg)(nil)

func NewBytesMsg(data []byte) *BytesMsg {
	return &BytesMsg{
		//Header: Header{
		//	MsgID:   id,
		//	MsgType: mtype,
		//	SrcPeerID: src,
		//	DstPeerID: dst,
		//},
		Data: data,
	}
}

func (m *BytesMsg) SetData(body []byte) {
	m.Data = body
}

func (m *BytesMsg) Encode() ([]byte, error) {
	//buf := make([]byte, HeaderSize+len(m.Data))
	//if err := m.Header.Encode(buf); err != nil {
	//	return nil, err
	//}

	buf := make([]byte, len(m.Data))
	//body := buf[HeaderSize:]
	copy(buf, m.Data)

	return buf, nil
}

func (m *BytesMsg) Decode(buf []byte) error {
	//if err := m.Header.Decode(buf); err != nil {
	//	return err
	//}

	m.Data = make([]byte, len(buf))
	copy(m.Data, buf)

	return nil
}

func (m *BytesMsg) MarshalJSON() ([]byte, error) {
	type Alias BytesMsg
	return json.Marshal(&struct {
		Body string `json:"body"`
		*Alias
	}{
		Body:  hex.EncodeToString(m.Data),
		Alias: (*Alias)(m),
	})
}

func (m *BytesMsg) UnmarshalJSON(bz []byte) error {
	type Alias BytesMsg
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

	m.Data, err = hex.DecodeString(alias.Body)
	if err != nil {
		return err
	}
	return nil
}

func (m *BytesMsg) String() string {
	bz, err := json.Marshal(m)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}

//type StrSimMsg struct {
//	//Header
//	Data string `json:"body"`
//}
//
//func NewStrSimMsg(mtype uint16, data string) *StrSimMsg {
//	return &StrSimMsg{
//		Header: Header{
//			MsgType: mtype,
//		},
//		Data: data,
//	}
//}
//
//func (m *StrSimMsg) String() string {
//	bz, err := json.Marshal(m)
//	if err != nil {
//		return err.Error()
//	}
//	return string(bz)
//}
