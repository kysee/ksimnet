package protocols

type TCPPack struct {
	seq     int
	src     int
	dst     int
	payload []byte
}

func DecodeTCPPack(d []byte) (*TCPPack, error) {
	return nil, nil
}

func (tcppack *TCPPack) Encode() ([]byte, error) {
	return nil, nil
}
