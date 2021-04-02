package knetsim

type Server struct {
	Host
	conns map[string]*Conn
}

func NewServer(laddr string, port int) (*Server, error) {
	h, err := NewHost(laddr, port)
	if err != nil {
		return nil, err
	}

	if err := AddServer(h); err != nil {
		return nil, err
	}
	h.conns = make(map[string]*Conn)
	
	return h, nil
}

func (s *Server) Listen() {
		
}

func (s *Server) Accept() *Conn {

}
	
