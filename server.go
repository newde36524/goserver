package goserver

import (
	"net"
	"runtime/debug"
	"time"
)

//Server tcp服务器
type Server struct {
	listener   net.Listener //TCP监听对象
	connOption ConnOption   //连接配置项
	isDebug    bool         //是否开始debug日志
	handles    []Handle     //连接处理程序管道
}

//New new server
//@network network 类型，具体参照ListenUDP ListenTCP等
//@addr local address
//@connOption connection options
func New(network, addr string, connOption ConnOption) (srv *Server, err error) {
	// 根据服务器开启多CPU功能
	// runtime.GOMAXPROCS(runtime.NumCPU())
	listener, err := net.Listen(network, addr)
	if err != nil {
		return
	}
	srv = &Server{
		listener:   listener,
		connOption: connOption,
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
func (s *Server) Binding() {
	go func() {
		defer s.listener.Close()
		defer func() {
			defer recover()
			if err := recover(); err != nil {
				s.connOption.Logger.Error(err)
				s.connOption.Logger.Error(debug.Stack())
			}
		}()
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				s.connOption.Logger.Error(err)
				<-time.After(time.Second)
				continue
			}
			c := NewConn(conn, s.connOption, s.handles)
			if s.isDebug {
				c.UseDebug()
			}
			c.Run()
		}
	}()
}
