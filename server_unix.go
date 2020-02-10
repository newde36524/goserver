// +build linux

package goserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
)

const (
	EPOLLET        = 1 << 31
	MaxEpollEvents = 1000000
)

//Server tcp服务器
type Server struct {
	// EventHandle
	isDebug   bool      //是否开始debug日志
	handles   []Handle  //连接处理程序管道
	network   string    //网络
	modOption ModOption //连接配置项
	listener  net.Listener
	option    *ConnOption
	ep        *epoll
}

//Binding start server
func (s *Server) Binding(address string) {
	listener, err := net.Listen(s.network, address)
	if err != nil {
		return
	}
	option := initOptions(s.modOption)
	s.listener = listener
	s.option = option
	s.ep = NewEpoll(MaxEpollEvents, option.MaxGopollTasks, option.MaxGopollExpire)
	s.run()
}

func (s *Server) run() {
	listenfd, err := netListenerToListenFD(s.listener)
	if err != nil {
		s.option.Logger.Errorf("server.epoll: %s\n", err)
	}
	if err = syscall.SetNonblock(int(listenfd), true); err != nil { //设置非阻塞模式
		fmt.Println("setnonblock1: ", err)
		os.Exit(1)
	}
	s.ep.Register(listenfd, s)
	go s.ep.Polling()
}

//OnReadable .
func (s *Server) OnReadable() {
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
	if err := s.ep.Register(connFd, NewConn(context.Background(), rwc, *s.option, s.handles)); err != nil {
		fmt.Println(err)
	}
}

//OnWriteable .
func (s *Server) OnWriteable() {
	s.option.Logger.Info("server_unix.go: Server OnWriteable")
}

func netListenerToListenFD(listener net.Listener) (listenFD int32, err error) {
	switch v := interface{}(listener).(type) {
	case *net.TCPListener:
		if raw, err := v.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				listenFD = int32(fd)
			})
		} else {
			return 0, err
		}
	default:
		return 0, errors.New("type can not get fd")
	}
	return
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
