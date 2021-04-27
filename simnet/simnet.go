package simnet

import (
	"errors"
	"fmt"
	"sync"
)

var listenerMtx sync.Mutex
var sessionMtx sync.Mutex

var listeners map[string]*Listener
var sessions map[string]*Session

func init() {
	Reset()
}

func Reset() {
	listenerMtx.Lock()
	listeners = make(map[string]*Listener)
	listenerMtx.Unlock()

	sessionMtx.Lock()
	sessions = make(map[string]*Session)
	sessionMtx.Unlock()

	ResetIPPorts()
}

func AddListener(s *Listener) error {
	k := s.Key()

	listenerMtx.Lock()
	defer listenerMtx.Unlock()

	if h := findListener(k); h != nil {
		return fmt.Errorf("Already exist")
	}

	listeners[k] = s

	return nil
}

func RemoveListener(lsn *Listener) {
	k := lsn.Key()

	listenerMtx.Lock()
	defer listenerMtx.Unlock()

	delete(listeners, k)
}

func findListener(k string) *Listener {
	if h, ok := listeners[k]; ok {
		return h
	}
	return nil
}

func FindListener(addr string) *Listener {
	listenerMtx.Lock()
	defer listenerMtx.Unlock()

	return findListener(addr)
}

func PrintListener() {
	listenerMtx.Lock()
	defer listenerMtx.Unlock()

	for k, s := range listeners {
		fmt.Println(k, s.listenAddr.String())
	}
}

func BuildSession(client *NetPoint, to string) (*Session, error) {
	s := FindListener(to)
	if s == nil {
		return nil, errors.New("not reachanble address: " + to)
	}

	sess := NewSession(client, nil)
	client.SetSession(sess)

	s.listenCh <- sess
	err := <-sess.ReqCh

	if err != nil {
		return nil, err
	}

	// NOTE: The following code is moved to Listener::accept() at listener.go
	// Because OnAccept() is already called,
	//  the server can send something before client.SetRemotePoint() is called in client side.
	// In the case, the client's NetPoint object has no NetConn object for the server side.
	// To resolve this problem,
	//  the 'client.SetRemotePoint()' SHOULD BE CALLED before ServerWorker.OnAccept() is called in server side.
	//
	// client.SetRemotePoint(sess.GetNetConn(SERVER).(*NetPoint))

	// add a session
	sessionMtx.Lock()
	defer sessionMtx.Unlock()

	sessions[sess.Key()] = sess

	return sess, nil
}

func RemoveSession(sess *Session) {
	sessionMtx.Lock()
	defer sessionMtx.Unlock()

	delete(sessions, sess.Key())
	sess.listener.RemoveNetConn(sess.GetNetConn(SERVER))
}
