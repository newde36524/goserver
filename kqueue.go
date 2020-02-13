// +build darwin netbsd freebsd openbsd dragonfly

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
		kqueueFd int
		events   []syscall.Kevent_t
		changes  []syscall.Kevent_t
		fdMap    sync.Map
		gopool   *GoPool
	}
)

//NewEpoll .
func NewNetpoll(maxEvents int, gopool *GoPool) *netpoll {
	kqueueFd, err := syscall.Kqueue()
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	_, err = syscall.Kevent(kqueueFd, []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}, nil, nil)
	if err != nil {
		panic(err)
	}
	changes := append([]syscall.Kevent_t,
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ,
		},
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE,
		},
	)
	return &netpoll{
		kqueueFd: kqueueFd,
		changes:  changes,
		events:   make([]syscall.Kevent_t, maxEvents), //指定一次获取多少个就绪事件
		gopool:   gopool,                              //指定协程池容量
	}
}

//Register .
func (e *netpoll) Register(fd int32, evh eventHandle) error {
	changes := append([]syscall.Kevent_t,
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.NOTE_TRIGGER, Filter: syscall.EVFILT_USER,
		},
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ,
		},
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE,
		},
	)
	if _, err := syscall.Kevent(e.kqueueFd, changes, nil, nil); err != nil {
		return err
	}
	e.fdMap.Store(fd, evh)
	return nil
}

//Remove .
func (e *netpoll) Remove(fd int32) {
	e.fdMap.Delete(uint64(fd))
}

//Polling .
func (e *netpoll) Polling() {
	var (
		isReadEvent = func(events uint64) bool {
			return true
		}
		isWriteEvent = func(events uint64) bool { //暂不知如何判断kqueue的读写事件
			return false
		}
	)
	for {
		eventCount, err := syscall.Kevent(e.kqueue, e.changes, e.events, nil)
		if err != nil && err != syscall.EINTR {
			fmt.Println(err)
			continue
		}
		for i := 0; i < eventCount; i++ { //遍历每个事件
			event := e.events[i]
			v, ok := e.fdMap.Load(event.Ident)
			if !ok || v == nil {
				fmt.Println("kqueue.Polling: no ", event.Ident)
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
