package goserver

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"time"
)

//TCPServer create tcp server
func TCPServer(modOption ModOption) (*Server, error) {
	return New("tcp", modOption)
}

//Server tcp服务器
type Server struct {
	isDebug   bool      //是否开始debug日志
	handles   []Handle  //连接处理程序管道
	network   string    //网络
	modOption ModOption //连接配置项
}

//New new server
//@network network 类型，具体参照ListenUDP ListenTCP等
//@addr local address
//@opt connection options
func New(network string, modOption ModOption) (srv *Server, err error) {
	// 根据服务器开启多CPU功能
	// runtime.GOMAXPROCS(runtime.NumCPU())
	srv = &Server{
		network:   network,
		modOption: modOption,
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
	option := initOptions(s.modOption)
	go func() {
		ctx, cancle := context.WithCancel(context.Background())
		defer cancle()
		defer listener.Close()
		defer func() {
			defer recover()
			if err := recover(); err != nil {
				if option.Logger != nil {
					option.Logger.Error(err)
					option.Logger.Error(debug.Stack())
				} else {
					fmt.Println(err)
					fmt.Println(debug.Stack())
				}
			}
		}()
		for {
			conn, err := listener.Accept()
			if err != nil {
				option.Logger.Error(err)
				<-time.After(time.Second)
				continue
			}
			c := NewConn(ctx, conn, *option, s.handles)
			if s.isDebug {
				c.UseDebug()
			}
			c.Run()
		}
	}()
}
