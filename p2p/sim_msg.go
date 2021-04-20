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
	Header SimMsgHeader  `json:"header"`
	Body   types.MsgBody `json:"body"`
}

func NewAnonySimMsg(body types.MsgBody) *SimMsg {
	return NewSimMsg(types.NewZeroPeerID(), types.NewZeroPeerID(), body)
}
func NewSimMsg(src, dst types.PeerID, body types.MsgBody) *SimMsg {
	return &SimMsg{
		Header: SimMsgHeader{
			MsgID:     types.CalMsgID(body),
			MsgType:   body.ConstType(),
			SrcPeerID: src,
			DstPeerID: dst,
		},
		Body: body,
	}
}

var _ types.Message = (*SimMsg)(nil)

func (sm *SimMsg) ID() types.MsgID {
	return sm.Header.ID()
}

func (sm *SimMsg) Type() uint16 {
	return sm.Header.Type()
}

func (sm *SimMsg) Src() types.PeerID {
	return sm.Header.Src()
}

func (sm *SimMsg) Dst() types.PeerID {
	return sm.Header.Dst()
}

func (sm *SimMsg) Encode() ([]byte, error) {

	header, err := sm.Header.Encode()
	if err != nil {
		return nil, err
	}

	body, err := sm.Body.Encode()
	if err != nil {
		return nil, err
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
	default:
		m := &BytesMsg{}
		if err := m.Decode(buf[HeaderSize:]); err != nil {
			return err
		}
		sm.Body = m
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

func (sm *SimMsg) ConstType() uint16 {
	return sm.Body.ConstType()
}

func (sm *SimMsg) Hash() []byte {
	return sm.Body.Hash()
}

const HeaderSize = types.MsgIDSize + types.MsgTypeSize + types.PeerIDSize*2

type SimMsgHeader struct {
	MsgID     types.MsgID  `json:"msg_id"`
	MsgType   uint16       `json:"msg_type"`
	SrcPeerID types.PeerID `json:"src"`
	DstPeerID types.PeerID `json:"dst,omitempty"`
}

func (h *SimMsgHeader) ID() types.MsgID {
	return h.MsgID
}

func (h *SimMsgHeader) Type() uint16 {
	return h.MsgType
}

func (h *SimMsgHeader) SetSrc(peerId types.PeerID) {
	h.SrcPeerID = peerId
}

func (h *SimMsgHeader) Src() types.PeerID {
	return h.SrcPeerID
}

func (h *SimMsgHeader) SetDst(peerId types.PeerID) {
	h.DstPeerID = peerId
}

func (h *SimMsgHeader) Dst() types.PeerID {
	return h.DstPeerID
}

func (h *SimMsgHeader) Encode() ([]byte, error) {

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

func (h *SimMsgHeader) Decode(buf []byte) error {
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

func (h *SimMsgHeader) String() string {
	bz, err := json.Marshal(h)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}

type SimMsgBodyFuncs struct{}

func (b *SimMsgBodyFuncs) String() string {
	bz, err := json.Marshal(b)
	if err != nil {
		return err.Error()
	}
	return string(bz)
}

type ReqPeers struct {
	SimMsgBodyFuncs
	ExportAddr *net.TCPAddr `json:"export_addr"`
}

var _ types.MsgBody = (*ReqPeers)(nil)

func NewReqPeers(exAddr *net.TCPAddr) *ReqPeers {
	return &ReqPeers{
		ExportAddr: exAddr,
	}
}

func (m *ReqPeers) ConstType() uint16 {
	return REQ_PEERS
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

func (m *ReqPeers) Hash() []byte {
	return types.CalMsgHash(m)
}

type AckPeers struct {
	SimMsgBodyFuncs
	Addrs []*net.TCPAddr `json:"addrs"`
}

var _ types.MsgBody = (*AckPeers)(nil)

func NewAckPeers() *AckPeers {
	return &AckPeers{}
}

func (m *AckPeers) AddPeerAddr(addr *net.TCPAddr) {
	m.Addrs = append(m.Addrs, addr)
}

func (m *AckPeers) ConstType() uint16 {
	return ACK_PEERS
}

func (m *AckPeers) Encode() ([]byte, error) {
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

func (m *AckPeers) Hash() []byte {
	return types.CalMsgHash(m)
}

type BytesMsg struct {
	SimMsgBodyFuncs
	Data []byte `json:"data"`
}

var _ types.MsgBody = (*BytesMsg)(nil)

func NewBytesMsg(data []byte) *BytesMsg {
	return &BytesMsg{
		Data: data,
	}
}

func (m *BytesMsg) SetData(body []byte) {
	m.Data = body
}

func (m *BytesMsg) ConstType() uint16 {
	return USER_MSG_TYPE
}

func (m *BytesMsg) Encode() ([]byte, error) {
	buf := make([]byte, len(m.Data))
	copy(buf, m.Data)

	return buf, nil
}

func (m *BytesMsg) Decode(buf []byte) error {
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

func (m *BytesMsg) Hash() []byte {
	return types.CalMsgHash(m)
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
