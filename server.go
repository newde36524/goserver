package goserver

import (
	"net"
	"runtime/debug"
	"time"
)

//TCPServer create tcp server
func TCPServer(opt ConnOption) (*Server, error) {
	return New("tcp", opt)
}

//Server tcp服务器
type Server struct {
	isDebug bool       //是否开始debug日志
	handles []Handle   //连接处理程序管道
	network string     //网络
	opt     ConnOption //连接配置项
}

//New new server
//@network network 类型，具体参照ListenUDP ListenTCP等
//@addr local address
//@connOption connection options
func New(network string, opt ConnOption) (srv *Server, err error) {
	// 根据服务器开启多CPU功能
	// runtime.GOMAXPROCS(runtime.NumCPU())
	srv = &Server{
		network: network,
		opt:     opt,
	}
	return
}

//Use middleware
func (s *Server) Use(h Handle) {
	s.handles = append(s.handles, h)
}

//UseDebug 开启debug日志
func (s *Server) UseDebug() {
	s.isDebug = true
}

//Binding start server
func (s *Server) Binding(address string) {
	listener, err := net.Listen(s.network, address)
	if err != nil {
		return
	}
	go func() {
		defer listener.Close()
		defer func() {
			defer recover()
			if err := recover(); err != nil {
				s.opt.Logger.Error(err)
				s.opt.Logger.Error(debug.Stack())
			}
		}()
		for {
			conn, err := listener.Accept()
			if err != nil {
				s.opt.Logger.Error(err)
				<-time.After(time.Second)
				continue
			}
			c := NewConn(conn, s.opt, s.handles)
			if s.isDebug {
				c.UseDebug()
			}
			c.Run()
		}
	}()
}
