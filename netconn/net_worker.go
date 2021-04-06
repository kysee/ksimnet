package netconn

type NetWorker interface {
	OnRecv(*NetPoint, []byte, int) (int, error)
	OnClose(*NetPoint) error
}

type ClientWorker interface {
	OnConnect(*NetPoint) error
	NetWorker
}

type ServerWorker interface {
	OnAccept(*NetPoint) error
	NetWorker
}
