package simnet

import (
	"errors"
	"fmt"
	"github.com/ksimnet/netapp"
	"net"
	"strconv"
	"sync"
)

var gmtx sync.Mutex

var servers map[string]*Server
var sessions map[string]*Session

var (
	a byte = 1
	b byte = 0
	c byte = 0
	d byte = 0
)
var ports map[string]int = make(map[string]int)

func init() {
	servers = make(map[string]*Server)
	sessions = make(map[string]*Session)
}

func hkey(addr string, port int) string {
	return addr + ":" + strconv.Itoa(port)
}

func Init(netAddr string) {
	gmtx.Lock()
	defer gmtx.Unlock()
}

func NewIP() string {
	d++
	return net.IPv4(a, b, c, d).String()
}

func NewPort(host string) int {
	if p, ok := ports[host]; ok {
		ports[host] = p + 1
		return p + 1
	}
	ports[host] = 1
	return 1
}

func Connect(app netapp.ClientApp, hostIP, toAddr string) (*ClientConn, error) {
	c, err := NewClientConn(app, hostIP)
	if err != nil {
		return nil, err
	}

	sess, err := buildSession(c, toAddr)
	if err != nil {
		return nil, err
	}
	c.SetRemoteConn(sess.GetNetConn(1))

	return c, nil
}

func buildSession(client *ClientConn, to string) (*Session, error) {
	s := findServer(to)
	if s == nil {
		return nil, errors.New("not reachanble address: " + to)
	}

	sess := NewSession(client.NetConn, nil)
	client.SetSession(sess)

	s.listenCh <- sess
	<-sess.evtCh

	return sess, nil
}

func AddServer(s *Server) error {
	k := s.Key()

	gmtx.Lock()
	defer gmtx.Unlock()

	if h := findServer(k); h != nil {
		return fmt.Errorf("Already exist")
	}

	servers[k] = s

	return nil
}

func RemoveServer(s *Server) error {
	k := s.Key()

	gmtx.Lock()
	defer gmtx.Unlock()

	if h := findServer(k); h != nil {
		return fmt.Errorf("Already exist")
	}

	servers[k] = nil

	return nil
}

func findServer(k string) *Server {
	if h, ok := servers[k]; ok {
		return h
	}
	return nil
}

func FindServer(addr string, port int) *Server {
	gmtx.Lock()
	defer gmtx.Unlock()

	return findServer(hkey(addr, port))

}

func PrintServers() {
	gmtx.Lock()
	defer gmtx.Unlock()

	for k, s := range servers {
		fmt.Println(k, s.listenAddr.String())
	}
}
