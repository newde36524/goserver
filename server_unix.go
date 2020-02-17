// +build linux darwin netbsd freebsd openbsd dragonfly

package goserver

import (
	"context"
	"errors"
	"fmt"
	"net"
)

const (
	EPOLLET   = 1 << 31
	maxEvents = 1000000
)

//Server tcp服务器
type Server struct {
	isDebug   bool      //是否开始debug日志
	pipe      Pipe      //连接处理程序管道
	network   string    //网络
	modOption ModOption //连接配置项
	listener  net.Listener
	option    *ConnOption
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
	option := initOptions(s.modOption)
	s.listener = listener
	s.option = option
	s.np = newNetpoll(maxEvents, newGoPool(option.MaxGopollTasks, option.MaxGopollExpire))
	go s.np.Polling()
	s.run()
}

func (s *Server) run() {
	for {
		rwc, err := s.listener.Accept()
		if err != nil {
			s.option.Logger.Error(err)
			return
		}
		connFd, err := netConnToConnFD(rwc)
		if err != nil {
			s.option.Logger.Error(err)
			return
		}
		// if err := syscall.SetNonblock(int(connFd), true); err != nil { //设置非阻塞模式
		// 	os.Exit(1)
		// }
		conn := NewConn(s.ctx, rwc, *s.option)
		conn.UsePipe(s.pipe)
		conn.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnConnection(ctx, conn, next) })
		if err := s.np.Register(connFd, conn); err != nil {
			fmt.Println(err)
		}
	}
}

func netConnToConnFD(conn net.Conn) (connFD int32, err error) {
	switch v := interface{}(conn).(type) {
	case *net.TCPConn:
		if raw, err := v.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				connFD = int32(fd)
			})
			return connFD, nil
		}
	case *net.UDPConn:
		if raw, err := v.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				connFD = int32(fd)
			})
			return connFD, nil
		}
	default:
		return 0, errors.New("type can not get fd")
	}
	return
}
