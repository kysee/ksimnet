package types

type NetInf interface {
	Key() string
	OnRx(NetInf, []byte) error
	DoTx(NetInf, []byte) error
}
