// +build darwin netbsd freebsd openbsd dragonfly

package goserver

import (
	"fmt"
	"syscall"
)

type netPoll struct {
	kqueueFd     int
	events       []syscall.Kevent_t
	changes      []syscall.Kevent_t
	eventAdapter eventAdapter
	gPool        *gPool
}

//newNetEpoll .
func newNetpoll(maxEvents int, gPool *gPool) *netPoll {
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
	return &netPoll{
		kqueueFd:     kqueueFd,
		changes:      changes,
		events:       make([]syscall.Kevent_t, maxEvents),
		gPool:        gPool,
		eventAdapter: newdefaultAdapter(),
	}
}

//Regist .
func (e *netPoll) Regist(fd uint64, evh eventHandle) error {
	changes := append([]syscall.Kevent_t{},
		// syscall.Kevent_t{
		// 	Ident: uint64(fd), Flags: syscall.NOTE_TRIGGER, Filter: syscall.EVFILT_USER,
		// },
		syscall.Kevent_t{
			Ident: fd, Flags: syscall.EV_ADD, Filter: syscall.EVFILT_READ,
		},
		syscall.Kevent_t{
			Ident: fd, Flags: syscall.EV_ADD, Filter: syscall.EVFILT_WRITE,
		},
	)
	if _, err := syscall.Kevent(e.kqueueFd, changes, nil, nil); err != nil {
		return err
	}
	e.eventAdapter.Link(fd, evh)
	return nil
}

//Remove .
func (e *netPoll) Remove(fd uint64) {
	e.eventAdapter.UnLink(fd)
}

//Polling .
func (e *netPoll) Polling() {
	var (
		isWriteEvent = func(events int16) bool {
			return events == syscall.EVFILT_WRITE
		}
		isReadEvent = func(events int16) bool {
			return events == syscall.EVFILT_READ
		}
	)
	e.polling(func(fd uint64, events int16) error {
		evh := e.eventAdapter.Get(fd)
		if evh == nil {
			logError(fmt.Sprintf("no fd %d \n", fd))
			return nil
		}
		//在协程池中运行要保证同一个通道下的通信是串行的
		if isWriteEvent(events) {
			e.gPool.SchduleByKey(fd, evh.OnWriteable)
			// evh.OnWriteable()
		} else if isReadEvent(events) {
			e.gPool.SchduleByKey(fd, evh.OnReadable)
			// evh.OnReadable()
		}
		return nil
	})
}

func (e *netPoll) polling(onEventTrigger func(fd uint64, events int16) error) {
	for {
		eventCount, err := syscall.Kevent(e.kqueueFd, nil, e.events, nil)
		if err != nil && err != syscall.Errno(0x4) {
			logError(err.Error())
			continue
		}
		for i := 0; i < eventCount; i++ { //遍历每个事件
			event := e.events[i]
			err := onEventTrigger(event.Ident, event.Filter)
			if err != nil {
				logError(err.Error())
				return
			}
		}
	}
}
