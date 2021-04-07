package types

type NetWorker interface {
	OnRecv(NetConn, []byte, int) (int, error)
	OnClose(NetConn)
}

type ClientWorker interface {
	OnConnect(NetConn)
	NetWorker
}

type ServerWorker interface {
	OnAccept(NetConn) error
	NetWorker
}
