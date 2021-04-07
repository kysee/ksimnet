package simnet

import (
	"errors"
	"github.com/ksimnet/types"
	"log"
	"net"
	"sync"
)

type Listener struct {
	mtx sync.RWMutex

	worker       types.ServerWorker
	clients      map[string]types.NetConn
	listenAddr   *net.TCPAddr
	listenCh     chan *Session
	stopListenCh chan struct{}
}

func NewListener(app types.ServerWorker, laddr string) (*Listener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", laddr)
	if err != nil {
		return nil, err
	}
	return &Listener{
		worker:       app,
		clients:      make(map[string]types.NetConn),
		listenAddr:   tcpAddr,
		listenCh:     make(chan *Session, 256), // backlog = 256
		stopListenCh: make(chan struct{}),
	}, nil
}

func (lsn *Listener) Key() string {
	return lsn.listenAddr.String()
}

func (lsn *Listener) Listen() error {

	go func() {
	Loop:
		for {
			select {
			case sess, ok := <-lsn.listenCh:
				if ok {

					sess.RemoteRetCh <- lsn.accept(sess)

					//log.Printf("Constructed session: %v\n", sess)
				} else {
					panic("error in receiving from listenCh")
					break Loop
				}
			case <-lsn.stopListenCh:
				break Loop
			}
		}

		RemoveListener(lsn)

		log.Printf("Listener(%s) is shutdowned", lsn.Key())
	}()

	AddListener(lsn)
	log.Printf("Listener(%s) is started", lsn.Key())
	return nil
}

func (lsn *Listener) Shutdown() {
	for _, c := range lsn.clients {
		c.Close()
	}

	lsn.stopListenCh <- struct{}{}
}

func (lsn *Listener) accept(sess *Session) error {
	c1 := sess.GetNetConn(CLIENT)
	if c1 == nil {
		return errors.New("a session has no NetConn")
	}

	c2 := NewNetPoint(lsn.worker, lsn.listenAddr)
	sess.SetNetConn(SERVER, c2)
	c2.SetSession(sess)
	c2.SetRemotePoint(c1.(*NetPoint))

	err := lsn.worker.OnAccept(c2)
	if err == nil {
		lsn.AddNetConn(c2)
	}
	sess.listener = lsn
	return err
}

func (lsn *Listener) AddNetConn(c types.NetConn) {
	lsn.mtx.Lock()
	defer lsn.mtx.Unlock()

	lsn.clients[c.Key()] = c
}

func (lsn *Listener) RemoveNetConn(c types.NetConn) {
	lsn.mtx.Lock()
	defer lsn.mtx.Unlock()

	delete(lsn.clients, c.Key())
}
