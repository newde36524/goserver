// +build linux

package goserver

import (
	"fmt"
	"sync"
	"syscall"
)

type (
	//eventHandle .
	eventHandle interface {
		OnReadable()
		OnWriteable()
	}

	eventHandleDec struct {
		eventHandle
		gopool *goPool
	}

	netPoll struct {
		epfd   int
		events []syscall.EpollEvent
		fdMap  sync.Map
		gopool *goPool
	}
)

//newNetpoll .
func newNetpoll(maxEvents int, gopool *goPool) *netPoll {
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	return &netPoll{
		epfd:   epfd,
		events: make([]syscall.EpollEvent, maxEvents),
		gopool: gopool,
	}
}

//Register .
func (e *netPoll) Register(fd int32, evh eventHandle) error {
	if err := syscall.EpollCtl(e.epfd, syscall.EPOLL_CTL_ADD, int(fd), &syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     fd,
	}); err != nil {
		return err
	}
	e.fdMap.Store(fd, evh)
	// e.fdMap.Store(fd, eventHandleDec{evh, e.gopool})
	return nil
}

//Remove .
func (e *netPoll) Remove(fd int32) {
	e.fdMap.Delete(fd)
}

//Polling .
func (e *netPoll) Polling() {
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
		isWriteEvent = func(event syscall.EpollEvent) bool {
			return event.Events&OutEvents == 1
		}
		isReadEvent = func(event syscall.EpollEvent) bool {
			return event.Events&InEvents == 1
		}
	)
	for {
		// 注意: 客户端断开时,直到服务端调用Close断开连接之前的时间内,EpollWait不会阻塞
		// 只要文件描述符存在，且处于io中断状态，EpollWait便不会等待
		// epoll原理就是通过中断来通知内核的，而客户端断开连接就使得文件描述符处于中断状态
		eventCount, err := syscall.EpollWait(e.epfd, e.events, -1)
		if err != nil {
			fmt.Println(err)
			continue
		}
		wg := sync.WaitGroup{}
		wg.Add(eventCount)
		for i := 0; i < eventCount; i++ {
			event := e.events[i]
			v, ok := e.fdMap.Load(event.Fd)
			if !ok || v == nil {
				fmt.Println("netpoll.Polling: no fd ", event.Fd)
				continue
			}
			evh, ok := v.(eventHandle)
			if !ok {
				continue
			}
			if isWriteEvent(event) {
				e.gopool.Schedule(func() {
					evh.OnWriteable()
					wg.Done()
				})
				// evh.OnWriteable()
			} else if isReadEvent(event) {
				e.gopool.Schedule(func() {
					evh.OnReadable()
					wg.Done()
				})
				// evh.OnReadable()
			}
		}
		wg.Wait()
	}
}
