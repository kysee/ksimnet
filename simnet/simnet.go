package simnet

import (
	"errors"
	"fmt"
	"github.com/ksimnet/types"
	"net"
	"strconv"
	"sync"
)

var addrMtx sync.Mutex
var listenerMtx sync.Mutex
var sessionMtx sync.Mutex

var listeners map[string]*Listener
var sessions map[string]*Session

var (
	a byte = 1
	b byte = 0
	c byte = 0
	d byte = 0
)
var ports map[string]int = make(map[string]int)

func init() {
	listeners = make(map[string]*Listener)
	sessions = make(map[string]*Session)
}

func NewIP() string {
	d++
	return net.IPv4(a, b, c, d).String()
}

func NewPort(host string) int {
	addrMtx.Lock()
	defer addrMtx.Unlock()

	if p, ok := ports[host]; ok {
		ports[host] = p + 1
		return p + 1
	}
	ports[host] = 1
	return 1
}

func bindPort(host string) (*net.TCPAddr, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(NewPort(host)))
	if err != nil {
		return nil, err
	}
	return tcpAddr, nil
}

func Connect(worker types.ClientWorker, hostIP, toAddr string) (types.NetConn, error) {
	bindAddr, err := bindPort(hostIP)
	if err != nil {
		return nil, err
	}

	c := NewNetPoint(worker, bindAddr)

	_, err = BuildSession(c, toAddr)
	if err != nil {
		return nil, err
	}

	worker.OnConnect(c)
	return c, nil
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
	err := <-sess.RemoteRetCh

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
