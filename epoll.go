package goserver

import (
	"fmt"
	"sync"
	"syscall"
	"time"
)

const (
	// ErrEvents represents exceptional events that are not read/write, like socket being closed,
	// reading/writing from/to a closed socket, etc.
	ErrEvents = syscall.EPOLLERR | syscall.EPOLLHUP | syscall.EPOLLRDHUP
	// OutEvents combines EPOLLOUT event and some exceptional events.
	OutEvents = ErrEvents | syscall.EPOLLOUT
	// InEvents combines EPOLLIN/EPOLLPRI events and some exceptional events.
	InEvents = ErrEvents | syscall.EPOLLIN | syscall.EPOLLPRI
)

type (
	//EventHandle .
	eventHandle interface {
		OnReadable()
		OnWriteable()
	}

	epoll struct {
		epfd   int
		events []syscall.EpollEvent
		fdMap  sync.Map
		gopoll *GoPoll
	}
)

//NewEpoll .
func NewEpoll(maxEpollEvents, maxGopollTasks int, maxGopollExpire time.Duration) *epoll {
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	return &epoll{
		epfd:   epfd,
		events: make([]syscall.EpollEvent, maxEpollEvents), //指定一次获取多少个就绪事件
		gopoll: NewGoPoll(maxGopollTasks, maxGopollExpire), //指定协程池容量
	}
}

//Register .
func (e *epoll) Register(fd int32, eventHandle eventHandle) error {
	if err := syscall.EpollCtl(e.epfd, syscall.EPOLL_CTL_ADD, int(fd), &syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     fd, //设置监听描述符
	}); err != nil {
		return err
	}
	e.fdMap.Store(fd, eventHandle)
	return nil
}

//Polling .
func (e *epoll) Polling() {
	var (
		isReadEvent = func(events uint32) bool {
			return events&InEvents == 1
		}
		isWriteEvent = func(events uint32) bool {
			return events&OutEvents == 1
		}
	)
	for {
		eventCount, err := syscall.EpollWait(e.epfd, e.events, -1) //获取就绪事件
		if err != nil {
			fmt.Println(err)
			continue
		}
		for i := 0; i < eventCount; i++ { //遍历每个事件
			event := e.events[i]
			if v, ok := e.fdMap.Load(event.Fd); ok {
				evh := v.(eventHandle)
				if isWriteEvent(event.Events) {
					e.gopoll.Schedule(evh.OnWriteable)
				}
				if isReadEvent(event.Events) {
					e.gopoll.Schedule(evh.OnReadable)
				}
			} else {
				fmt.Println("no ", event.Fd)
			}
		}
	}
}
