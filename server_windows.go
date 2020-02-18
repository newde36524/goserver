// +build windows

package goserver

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"time"
)

//Server tcp服务器
type Server struct {
	isDebug   bool      //是否开始debug日志
	pipe      Pipe      //连接处理程序管道
	network   string    //网络
	modOption ModOption //连接配置项
	opt       *ConnOption
	ctx       context.Context
	listener  net.Listener
	cancle    func()
}

//Binding start server
func (s *Server) Binding(address string) {
	listener, err := net.Listen(s.network, address)
	if err != nil {
		return
	}
	opt := initOptions(s.modOption)
	s.opt = opt
	s.listener = listener
	go s.run()
}

func (s *Server) run() {
	defer func() {
		defer recover()
		if err := recover(); err != nil {
			if s.opt.Logger != nil {
				s.opt.Logger.Error(err)
				s.opt.Logger.Error(string(debug.Stack()))
			} else {
				fmt.Println(err)
				fmt.Println(string(debug.Stack()))
			}
		}
		s.cancle()
		s.listener.Close()
	}()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.opt.Logger.Error(err)
			<-time.After(time.Second)
			continue
		}
		s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(rwc net.Conn) {
	conn := NewConn(s.ctx, rwc, *s.opt)
	conn.UsePipe(s.pipe)
	if s.isDebug {
		conn.UseDebug()
	}
	conn.Run()
}
