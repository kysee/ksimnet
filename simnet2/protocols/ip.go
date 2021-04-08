package protocols

import "net"

type IPPack struct {
	src     net.IP
	dst     net.IP
	payload []byte
}

func DecodeIPPack(d []byte) (*IPPack, error) {
	return nil, nil
}

func (ippack *IPPack) Encode() ([]byte, error) {
	return nil, nil
}
