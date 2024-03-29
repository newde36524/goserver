// +build linux

package goserver

import (
	"fmt"
	"sync"
	"syscall"
)

type netPoll struct {
	epfd   int
	events []syscall.EpollEvent
	evhs   map[uint64]eventHandle
	mu     sync.Mutex
	gPool  *gPool
}

//newNetpoll .
func newNetpoll(maxEvents int, gPool *gPool) *netPoll {
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	return &netPoll{
		epfd:   epfd,
		events: make([]syscall.EpollEvent, maxEvents),
		gPool:  gPool,
		evhs:   make(map[uint64]eventHandle, 1024),
	}
}

//Regist .
func (e *netPoll) Regist(fd uint64, evh eventHandle) error {
	if err := syscall.EpollCtl(e.epfd, syscall.EPOLL_CTL_ADD, int(fd), &syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(fd),
	}); err != nil {
		return err
	}
	e.mu.Lock()
	e.evhs[fd] = evh
	e.mu.Unlock()
	return nil
}

//Remove .
func (e *netPoll) Remove(fd uint64) (err error) {
	if err := syscall.EpollCtl(e.epfd, syscall.EPOLL_CTL_DEL, int(fd), nil); err != nil {
		return err
	}
	e.mu.Lock()
	delete(e.evhs, fd)
	e.mu.Unlock()
	return
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
		isWriteEvent = func(events uint32) bool {
			return events&OutEvents == 1
		}
		isReadEvent = func(events uint32) bool {
			return events&InEvents == 1
		}
	)
	e.polling(func(fd uint64, event uint32) error {
		e.mu.Lock()
		evh := e.evhs[fd]
		e.mu.Unlock()
		if evh == nil {
			logError(fmt.Sprintf("no fd %d \n", fd))
			return nil
		}
		//在协程池中运行要保证同一个FD下的通信是串行的
		if isWriteEvent(event) {
			e.gPool.SchduleByKey(fd, evh.OnWriteable)
			// evh.OnWriteable()
		} else if isReadEvent(event) {
			e.gPool.SchduleByKey(fd, evh.OnReadable)
			// evh.OnReadable()
		}
		return nil
	})
}

func (e *netPoll) polling(onEventTrigger func(fd uint64, events uint32) error) {
	for {
		// 注意: 客户端断开时,直到服务端调用Close断开连接之前的时间内,EpollWait不会阻塞
		// 只要文件描述符存在，且处于io中断状态，EpollWait便不会等待
		// epoll原理就是通过中断来通知内核的，而客户端断开连接就使得文件描述符处于中断状态
		eventCount, err := syscall.EpollWait(e.epfd, e.events, -1)
		if err != nil && err != syscall.Errno(0x4) {
			logError(err.Error())
			continue
		}
		for i := 0; i < eventCount; i++ {
			event := e.events[i]
			err := onEventTrigger(uint64(event.Fd), event.Events)
			if err != nil {
				logError(err.Error())
				return
			}
		}
	}
}
