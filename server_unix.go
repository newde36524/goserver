// +build linux

package goserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"syscall"
	"time"
)

const (
	EPOLLET        = 1 << 31
	MaxEpollEvents = 32
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
	epfd      int
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

//Use middleware
func (s *Server) Use(h Handle) {
	s.handles = append(s.handles, h)
}

//UseDebug 开启debug日志
func (s *Server) UseDebug() {
	s.isDebug = true
}

var clientMap map[int]*Conn = make(map[int]*Conn)



//Binding start server
func (s *Server) Binding(address string) {
	listener, err := net.Listen(s.network, address)
	if err != nil {
		return
	}
	option := initOptions(s.modOption)
	go s.eventLoop()
	go s.listen(listener, option)
}

func(s *Server) listen(listener net.Listener,option *ConnOption) {
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

		connFd := netConnToConnFD(conn)
		if err := syscall.EpollCtl(s.epfd, syscall.EPOLL_CTL_ADD, connFd, &syscall.EpollEvent{
			Events: syscall.EPOLLIN | EPOLLET,
			Fd:     int32(connFd),
		}); err != nil { //给epollFD 增加一个连接FD
			fmt.Println("epoll_ctl error: ", connFd, err)
			os.Exit(1)
		}

		c := NewConn(ctx, conn, connFd, *option, s.handles)
		clientMap[int(connFd)] = c
		if s.isDebug {
			c.UseDebug()
		}
		// c.Run()
	}
}


func (s *Server) eventLoop() {
	const (
		// ErrEvents represents exceptional events that are not read/write, like socket being closed,
		// reading/writing from/to a closed socket, etc.
		ErrEvents = syscall.EPOLLERR | syscall.EPOLLHUP | syscall.EPOLLRDHUP
		// OutEvents combines EPOLLOUT event and some exceptional events.
		OutEvents = ErrEvents | syscall.EPOLLOUT
		// InEvents combines EPOLLIN/EPOLLPRI events and some exceptional events.
		InEvents = ErrEvents | syscall.EPOLLIN | syscall.EPOLLPRI
	)

	var (
		isReadEvent = func(events uint32) bool {
			return events&InEvents == 1
		}
		isWriteEvent = func(events uint32) bool {
			return events&OutEvents == 1
		}
	)


	var events [MaxEpollEvents]syscall.EpollEvent //指定一次获取多少个就绪事件
	for {
		eventCount, err := syscall.EpollWait(s.epfd, events[:], -1) //获取就绪事件
		if err != nil {
			fmt.Println("epoll_wait: ", err)
			break
		}
		for i := 0; i < eventCount; i++ { //遍历每个事件
			event := events[i]
			if isWriteEvent(event.Events) {
				fmt.Println("write event",event.Events&OutEvents)
			}
			if isReadEvent(event.Events) {
				fmt.Println("read event",event.Events&InEvents)
				if conn, ok := clientMap[int(event.Fd)]; ok {
					conn.react()
					buf := make([]byte, 1024)
					n,_:= conn.Read(buf)
					conn.Write(&P{
						Data: buf[:n],
					})
				}
			}
		}
	}
}

func netListenerToListenFD(listener net.Listener) (listenFD int, err error) {
	switch v := interface{}(listener).(type) {
	case *net.TCPListener:
		if raw, err := v.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				listenFD = int(fd)
			})
		} else {
			return 0, err
		}
	default:
		err = errors.New("type can not get fd")
	}
	return
}

func netConnToConnFD(conn net.Conn) (connFD int) {
	switch v := interface{}(conn).(type) {
	case *net.TCPConn:
		if raw, err := v.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				connFD = int(fd)
			})
		}
	case *net.UDPConn:
		if raw, err := v.SyscallConn(); err == nil {
			raw.Control(func(fd uintptr) {
				connFD = int(fd)
			})
		}
	default:
	}
	return
}
