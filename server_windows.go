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
	ctx       context.Context
	cancle    func()
}

//Binding start server
func (s *Server) Binding(address string) {
	listener, err := net.Listen(s.network, address)
	if err != nil {
		return
	}
	option := initOptions(s.modOption)
	defer s.cancle()
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
		rwc, err := listener.Accept()
		if err != nil {
			option.Logger.Error(err)
			<-time.After(time.Second)
			continue
		}
		conn := NewConn(s.ctx, rwc, *option)
		conn.UsePipe(s.pipe)
		if s.isDebug {
			conn.UseDebug()
		}
		conn.Run()
	}
}
