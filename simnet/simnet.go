package simnet

import (
	"errors"
	"fmt"
	"github.com/ksimnet/netconn"
	"net"
	"strconv"
	"sync"
)

var gmtx sync.Mutex

var servers map[string]*Server
var sessions map[string]*netconn.Session

var (
	a byte = 1
	b byte = 0
	c byte = 0
	d byte = 0
)
var ports map[string]int = make(map[string]int)

func init() {
	servers = make(map[string]*Server)
	sessions = make(map[string]*netconn.Session)
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

func bindPort(host string) (*net.TCPAddr, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(NewPort(host)))
	if err != nil {
		return nil, err
	}
	return tcpAddr, nil
}

func Connect(worker netconn.ClientWorker, hostIP, toAddr string) (netconn.NetConn, error) {
	bindAddr, err := bindPort(hostIP)
	if err != nil {
		return nil, err
	}

	c := netconn.NewNetPoint(worker, bindAddr)

	_, err = buildSession(c, toAddr)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func buildSession(client *netconn.NetPoint, to string) (*netconn.Session, error) {
	s := findServer(to)
	if s == nil {
		return nil, errors.New("not reachanble address: " + to)
	}

	sess := netconn.NewSession(client, nil)
	client.SetSession(sess)

	s.listenCh <- sess
	<-sess.EvtCh

	client.SetRemotePoint(sess.GetNetPoint(1))

	// add a session
	sessions[sess.Key()] = sess

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
