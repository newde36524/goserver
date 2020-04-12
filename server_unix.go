// +build linux darwin netbsd freebsd openbsd dragonfly

package goserver

import (
	"context"
	"net"
	"runtime/debug"
)

const (
	// EPOLLET   = 1 << 31
	maxEvents = 1000000
)

//Server tcp服务器
type Server struct {
	isDebug      bool            //是否开始debug日志
	pipe         Pipe            //连接处理程序管道
	network      string          //网络
	modOption    ModOption       //连接配置项
	loopTaskPool loopTaskPool    //定时任务池
	listener     net.Listener    //.
	opt          *ConnOption     //.
	np           *netPoll        //.
	ctx          context.Context //.
	cancle       func()          //.
}

//Binding start server
func (s *Server) Binding(address string) {
	listener, err := net.Listen(s.network, address)
	if err != nil {
		panicError(err.Error())
	}
	opt := initOptions(s.modOption)
	s.listener = listener
	s.opt = opt
	//协程池的perItemTaskNum设置为0防止netPoll重复生成任务,为0时并不会阻塞协程池任务调度
	//由于netPoll的特性,产生的任务允许丢弃
	s.np = newNetpoll(maxEvents, newgPoll(s.ctx, 0, opt.MaxGopollExpire, opt.ParallelSize))
	listenFd, err := netListenerToListenFD(listener)
	if err != nil {
		panicError(err.Error())
	}
	if err := s.np.Regist(listenFd, s); err != nil {
		logError(err.Error())
	}
	go s.np.Polling()
}

//OnReadable .
func (s *Server) OnReadable() {
	defer func() {
		if err := recover(); err != nil {
			logError(err.(error).Error())
			logError(string(debug.Stack()))
		}
	}()
	rwc, err := s.listener.Accept()
	if err != nil {
		logError(err.Error())
		return
	}
	connFd, err := netConnToConnFD(rwc)
	if err != nil {
		logError(err.Error())
		return
	}
	conn := NewConn(s.ctx, rwc, *s.opt)
	conn.UseDebug(s.isDebug)
	conn.UsePipe(s.pipe)
	conn.pipe.schedule(func(h Handle, ctx interface{}) { h.OnConnection(ctx.(ConnectionContext)) }, newConnectionContext(conn))
	if err := s.np.Regist(connFd, conn); err != nil {
		logError(err.Error())
	}
	s.loopTaskPool.Schdule(s.opt.RecvTimeOut, func(remove func()) {
		defer func() {
			if err := recover(); err != nil {
				logError(err.(error).Error())
				logError(string(debug.Stack()))
			}
		}()
		if b := conn.IsRecvTimeOut(); b {
			select {
			case <-conn.ctx.Done():
				remove()
			default:
				conn.pipe.schedule(func(h Handle, ctx interface{}) { h.OnRecvTimeOut(ctx.(RecvTimeOutContext)) }, newRecvTimeOutContext(conn))
			}
		}
	})
}

//OnWriteable .
func (s *Server) OnWriteable() {
	logInfo("Server OnWriteable")
}
