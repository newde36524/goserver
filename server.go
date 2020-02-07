package goserver

//TCPServer create tcp server
func TCPServer(modOption ModOption) (*Server, error) {
	return New("tcp", modOption)
}

//Use middleware
func (s *Server) Use(h Handle) {
	s.handles = append(s.handles, h)
}

//UseDebug 开启debug日志
func (s *Server) UseDebug() {
	s.isDebug = true
}
