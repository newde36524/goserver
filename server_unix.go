// +build linux

package goserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
	// "math"
)

const (
	EPOLLET        = 1 << 31
	MaxEpollEvents = 1000000
)

//Server tcp服务器
type Server struct {
	isDebug   bool      //是否开始debug日志
	handles   []Handle  //连接处理程序管道
	network   string    //网络
	modOption ModOption //连接配置项
	epfd      int
	listener  net.Listener
	option    *ConnOption
}

//New new server
//@network network 类型，具体参照ListenUDP ListenTCP等
//@addr local address
//@opt connection options
func New(network string, modOption ModOption) (srv *Server, err error) {
	// 根据服务器开启多CPU功能
	// runtime.GOMAXPROCS(runtime.NumCPU())
	epfd, e := syscall.EpollCreate1(0)
	if e != nil {
		fmt.Println("epoll_create1: ", e)
		os.Exit(1)
	}
	srv = &Server{
		network:   network,
		modOption: modOption,
		epfd:      epfd,
	}
	return
}

// var clientMap map[int]Conn = make(map[int]Conn)
var clientMap = sync.Map{}

//Binding start server
func (s *Server) Binding(address string) {
	listener, err := net.Listen(s.network, address)
	if err != nil {
		return
	}
	option := initOptions(s.modOption)
	s.listener = listener
	s.option = option
	go s.epoll()
}

func (s *Server) epoll() {
	defer func() {
		defer recover()
		if err := recover(); err != nil {
			fmt.Println(err)
			fmt.Println(debug.Stack())
		}
	}()
	listenfd, err := netListenerToListenFD(s.listener)
	if err != nil {
		s.option.Logger.Errorf("server.epoll: %s\n", err)
	}
	if err := syscall.EpollCtl(s.epfd, syscall.EPOLL_CTL_ADD, int(listenfd), &syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(listenfd), //设置监听描述符
	}); err != nil {
		s.option.Logger.Error("epoll_ctl: ", err)
		os.Exit(1)
	}
	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()
	var events [MaxEpollEvents]syscall.EpollEvent //指定一次获取多少个就绪事件
	for {
		eventCount, err := syscall.EpollWait(s.epfd, events[:], -1) //获取就绪事件
		if err != nil {
			s.option.Logger.Error("epoll_wait: ", err)
			time.Sleep(time.Second)
			continue
		}
		for i := 0; i < eventCount; i++ { //遍历每个事件
			event := events[i]
			if event.Fd == listenfd {
				conn, err := s.listener.Accept()
				if err != nil {
					s.option.Logger.Error(err)
					<-time.After(time.Second)
					continue
				}
				connFd, err := netConnToConnFD(conn)
				if err != nil {
					s.option.Logger.Error(err)
					continue
				}
				if err := syscall.EpollCtl(s.epfd, syscall.EPOLL_CTL_ADD, connFd, &syscall.EpollEvent{
					Events: syscall.EPOLLIN | EPOLLET,
					Fd:     int32(connFd),
				}); err != nil { //给epollFD 增加一个连接FD
					s.option.Logger.Error("epoll_ctl error: ", connFd, err)
					continue
				}

				c := NewConn(ctx, conn, *s.option, s.handles)
				c.pipe(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnConnection(ctx, c, next) })
				clientMap.Store(int(connFd), c)

				if s.isDebug {
					c.UseDebug()
				}
			}
			if v, ok := clientMap.Load(int(event.Fd)); ok {
				conn := v.(Conn)
				conn.reactSrvEvent(event)
			}
		}
	}
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
		err = errors.New("type can not get fd")
	}
	return
}

func netConnToConnFD(conn net.Conn) (connFD int, err error) {
	switch v := interface{}(conn).(type) {
	case *net.TCPConn:
		if raw, err := v.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				connFD = int(fd)
			})
			return connFD, nil
		} else {
			return 0, err
		}
	case *net.UDPConn:
		if raw, err := v.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				connFD = int(fd)
			})
		} else {
			return 0, err
		}
	default:
	}
	return 0, errors.New("type can not get fd")
}
