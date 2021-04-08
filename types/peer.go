package types

type Peer interface {
	ID() int64

	Start(listenPort int) error
	Connect()
	Send()
	Stop()

	PeerCnt() int
	FindPeer(id int64) Peer

	ClientWorker
	ServerWorker
}
