package simnet

import (
	"errors"
	"github.com/kysee/ksimnet/types"
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

func NewListener(app types.ServerWorker) *Listener {
	return &Listener{
		worker:       app,
		clients:      make(map[string]types.NetConn),
		listenAddr:   &net.TCPAddr{IP: app.HostIP()},
		listenCh:     make(chan *Session, 256), // backlog = 256
		stopListenCh: make(chan struct{}),
	}
}

func (lsn *Listener) Key() string {
	return lsn.listenAddr.String()
}

func (lsn *Listener) Listen(port int) error {
	lsn.mtx.Lock()
	defer lsn.mtx.Unlock()

	if lsn.listenAddr.Port != 0 {
		return errors.New("this listener is already started")
	}

	lsn.listenAddr.Port = port

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

		lsn.mtx.Lock()
		defer lsn.mtx.Unlock()
		lsn.listenAddr.Port = 0
		//log.Printf("Listener(%s) is shutdowned", lsn.Key())
	}()

	AddListener(lsn)
	//log.Printf("Listener(%s) is started", lsn.Key())
	return nil
}

func (lsn *Listener) Shutdown() {
	for _, c := range lsn.clients {
		c.Close()
	}

	lsn.stopListenCh <- struct{}{}
}

func (lsn *Listener) ListenAddr() *net.TCPAddr {
	return lsn.listenAddr
}

func (lsn *Listener) accept(sess *Session) error {
	c1 := sess.GetNetConn(CLIENT)
	if c1 == nil {
		return errors.New("a session has no NetConn")
	}

	c2 := NewNetPoint(lsn.worker, lsn.listenAddr.Port, true)
	sess.SetNetConn(SERVER, c2)
	c2.SetSession(sess)
	c2.SetRemotePoint(c1.(*NetPoint))

	// This code comes from 'BuildSession()' in simnet.go.
	// See the comments in simnet.go for the reason.
	c1.(*NetPoint).SetRemotePoint(c2)

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
