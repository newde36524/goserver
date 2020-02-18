package goserver

import "context"

//TCPServer create tcp server
func TCPServer(modOption ModOption) (*Server, error) {
	return New("tcp", modOption)
}

//New new server
//@network network 类型，具体参照ListenUDP ListenTCP等
//@addr local address
//@opt connection options
func New(network string, modOption ModOption) (srv *Server, err error) {
	// runtime.GOMAXPROCS(runtime.NumCPU())
	srv = &Server{
		network:   network,
		modOption: modOption,
	}
	srv.ctx, srv.cancle = context.WithCancel(context.Background())
	return
}

//UsePipe .
func (s *Server) UsePipe(p ...Pipe) Pipe {
	if len(p) != 0 {
		s.pipe = p[0]
	}
	if s.pipe == nil {
		s.pipe = newPipe(s.ctx)
	}
	return s.pipe
}

//UseDebug 开启debug日志
func (s *Server) UseDebug() {
	s.isDebug = true
}
