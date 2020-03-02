// +build linux darwin netbsd freebsd openbsd dragonfly

package goserver

import (
	"context"
	"fmt"
	"net"
)

const (
	// EPOLLET   = 1 << 31
	maxEvents = 1000000
)

//Server tcp服务器
type Server struct {
	isDebug   bool      //是否开始debug日志
	pipe      Pipe      //连接处理程序管道
	network   string    //网络
	modOption ModOption //连接配置项
	listener  net.Listener
	opt       *ConnOption
	np        *netPoll
	ctx       context.Context
	cancle    func()
}

//Binding start server
func (s *Server) Binding(address string) {
	listener, err := net.Listen(s.network, address)
	if err != nil {
		panic(err)
	}
	opt := initOptions(s.modOption)
	s.listener = listener
	s.opt = opt
	//协程池的perItemTaskNum设置为0防止netPoll重复生成任务,为0时并不会阻塞协程池任务调度
	//由于netPoll的特性,产生的任务允许丢弃
	s.np = newNetpoll(maxEvents, newgPoll(s.ctx, 0, opt.MaxGopollExpire, opt.ParallelSize))
	go s.np.Polling()
	// s.run()

	listenFd, err := netListenerToListenFD(listener)
	if err != nil {
		panic(err)
	}
	if err := s.np.Regist(listenFd, s); err != nil {
		fmt.Println(err)
	}
}

//OnReadable .
func (s *Server) OnReadable() {
	rwc, err := s.listener.Accept()
	if err != nil {
		s.opt.Logger.Error(err)
		return
	}
	connFd, err := netConnToConnFD(rwc)
	if err != nil {
		s.opt.Logger.Error(err)
		return
	}
	conn := NewConn(s.ctx, rwc, *s.opt)
	conn.UsePipe(s.pipe)
	conn.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnConnection(ctx, conn, next) })
	if err := s.np.Regist(connFd, conn); err != nil {
		fmt.Println(err)
	}
}

//OnWriteable .
func (s *Server) OnWriteable() {
	s.opt.Logger.Info("goserver.server_unix.go: Server OnWriteable")
}
