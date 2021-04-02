package knetsim

var gmtx sync.Mutex
var servers map[string]*Host

func init() {
	servers = make(map[string]*Host)
}

func hkey(addr string, port int) string {
	return addr + ":" + strconv.Itoa(port)
}

func findHost(k string) *Host {
	if h, ok := servers[k]; ok {
		return h
	}
	return nil
}

func AddServer(h *Host) error {
	hk := h.Key()

	gmtx.Lock()
	defer gmtx.Unlock()

	if h := findHost(hk); h != nil {
		return fmt.Errorf("Already exist")
	}

	servers[hk] = h

	return nil
}

func RemoveServer(h *Host) error {
	hk := h.Key()

	gmtx.Lock()
	defer gmtx.Unlock()

	if h := findHost(hk); h != nil {
		return fmt.Errorf("Already exist")
	}

	servers[hk] = nil

	return nil
}

func RemoveHost(h *Host) error {

}

func FindHost(addr string, port int) *Host {
	gmtx.Lock()
	defer gmtx.Unlock()

	return findHost(hkey(addr, port))

}

func PrintServers() {
	gmtx.Lock()
	defer gmtx.Unlock()

	for hk, h := range servers {
		fmt.Println(hk, h.Ip.String(), h.Port)
	}
}

