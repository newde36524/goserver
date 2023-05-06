package goserver

import (
	"context"
	"net"
)

var (
	isDebug bool //is open inner debug message flag
)

type (
	//eventHandle .
	eventHandle interface {
		OnReadable()
		OnWriteable()
	}
)

//TCPServer create tcp server
func TCPServer(modOption ModOption) (*Server, error) {
	return New("tcp", modOption)
}

//New new server
//@network network 类型，具体参照ListenUDP ListenTCP等
//@modOption to set option
func New(network string, modOption ModOption) (srv *Server, err error) {
	// runtime.GOMAXPROCS(runtime.NumCPU())
	srv = &Server{
		network:   network,
		modOption: modOption,
		pipe:      &pipeLine{},
		isNoDelay: true,
	}
	srv.ctx, srv.cancle = context.WithCancel(context.Background())
	return
}

//UsePipe .
func (s *Server) UsePipe(p ...Pipe) Pipe {
	if len(p) != 0 {
		s.pipe = p[0]
	}
	return s.pipe
}

//UseDebug 开启debug日志
func (s *Server) UseDebug() {
	s.isDebug = true
}

//UseDebug 开启debug日志
func (s *Server) SetNoDelay(b bool) {
	s.isNoDelay = b
}

func netConnToConnFD(conn net.Conn) (connFD uint64, err error) {
	switch v := interface{}(conn).(type) {
	case *net.TCPConn:
		if raw, err := v.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				connFD = uint64(fd)
			})
			return connFD, nil
		}
	case *net.UDPConn:
		if raw, err := v.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				connFD = uint64(fd)
			})
			return connFD, nil
		}
	default:
		return 0, errorFdNotfound
	}
	return
}

func netListenerToListenFD(listener net.Listener) (listenFD uint64, err error) {
	switch v := interface{}(listener).(type) {
	case *net.TCPListener:
		if raw, err := v.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				listenFD = uint64(fd)
			})
		} else {
			return 0, err
		}
	default:
		return 0, errorFdNotfound
	}
	return
}
