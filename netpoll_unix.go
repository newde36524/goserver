// +build darwin netbsd freebsd openbsd dragonfly

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

	netPoll struct {
		kqueueFd int
		events   []syscall.Kevent_t
		changes  []syscall.Kevent_t
		fdMap    sync.Map
		gopool   *goPool
	}
)

//newNetEpoll .
func newNetpoll(maxEvents int, gopool *goPool) *netPoll {
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
	changes := append([]syscall.Kevent_t{},
		syscall.Kevent_t{
			Ident: uint64(kqueueFd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ,
		},
		syscall.Kevent_t{
			Ident: uint64(kqueueFd), Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE,
		},
	)
	return &netpoll{
		kqueueFd: kqueueFd,
		changes:  changes,
		events:   make([]syscall.Kevent_t, maxEvents),
		gopool:   gopool,
	}
}

//Register .
func (e *netPoll) Register(fd int32, evh eventHandle) error {
	changes := append([]syscall.Kevent_t{},
		// syscall.Kevent_t{
		// 	Ident: uint64(fd), Flags: syscall.NOTE_TRIGGER, Filter: syscall.EVFILT_USER,
		// },
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
func (e *netPoll) Remove(fd int32) {
	e.fdMap.Delete(uint64(fd))
}

//Polling .
func (e *netPoll) Polling() {
	var (
		isWriteEvent = func(event syscall.Kevent_t) bool {
			return event.Filter == syscall.EVFILT_WRITE
		}
		isReadEvent = func(event syscall.Kevent_t) bool {
			return event.Filter == syscall.EVFILT_READ
		}
	)
	wg := sync.WaitGroup{}
	wg.Add(eventCount)
	for {
		eventCount, err := syscall.Kevent(e.kqueueFd, nil, e.events, nil)
		if err != nil && err != syscall.EINTR {
			fmt.Println(err)
			continue
		}
		for i := 0; i < eventCount; i++ { //遍历每个事件
			event := e.events[i]
			v, ok := e.fdMap.Load(event.Ident)
			if !ok || v == nil {
				fmt.Println("netpoll.Polling: no fd ", event.Ident)
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
	}
}
