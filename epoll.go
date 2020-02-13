// +build linux

package goserver

import (
	"fmt"
	"sync"
	"syscall"
)

type (
	//EventHandle .
	eventHandle interface {
		OnReadable()
		OnWriteable()
	}

	netpoll struct {
		epfd   int
		events []syscall.EpollEvent
		fdMap  sync.Map
		gopool *GoPool
	}
)

//NewNetpoll .
func NewNetpoll(maxEvents int, gopool *GoPool) *netpoll {
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	return &netpoll{
		epfd:   epfd,
		events: make([]syscall.EpollEvent, maxEvents), //指定一次获取多少个就绪事件
		gopool: gopool,                                //指定协程池容量
	}
}

//Register .
func (e *netpoll) Register(fd int32, evh eventHandle) error {
	if err := syscall.EpollCtl(e.epfd, syscall.EPOLL_CTL_ADD, int(fd), &syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     fd, //设置监听描述符
	}); err != nil {
		return err
	}
	e.fdMap.Store(fd, evh)
	return nil
}

//Remove .
func (e *netpoll) Remove(fd int32) {
	e.fdMap.Delete(fd)
}

//Polling .
func (e *netpoll) Polling() {
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
	for {
		eventCount, err := syscall.EpollWait(e.epfd, e.events, -1) //获取就绪事件 单位:毫秒 1000毫秒=1秒，-1时无限等待
		if err != nil {
			fmt.Println(err)
			continue
		}
		for i := 0; i < eventCount; i++ { //遍历每个事件
			event := e.events[i]
			v, ok := e.fdMap.Load(event.Fd)
			if !ok || v == nil {
				fmt.Println("epoll.Polling: no ", event.Fd)
				continue
			}
			evh := v.(eventHandle)
			if isWriteEvent(event.Events) {
				e.gopool.Schedule(evh.OnWriteable)
			} else if isReadEvent(event.Events) {
				e.gopool.Schedule(evh.OnReadable)
			}
		}
	}
}
