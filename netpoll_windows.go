// +build windows

package goserver

import (
	"fmt"
	"syscall"
)

type netPoll struct {
	handle       syscall.Handle
	eventAdapter eventAdapter
	gPool        *gPool
	complateKey  uint32
}

//newNetpoll .
func newNetpoll(maxEvents int, gPool *gPool) *netPoll {
	handle, err := syscall.CreateIoCompletionPort(syscall.InvalidHandle, 0, 0, 0)
	if err != nil {
		panic(err)
	}
	return &netPoll{
		handle:       handle,
		gPool:        gPool,
		eventAdapter: newdefaultAdapter(),
	}
}

//Regist .
func (e *netPoll) Regist(fd uint64, evh eventHandle) error {
	handle, err := syscall.CreateIoCompletionPort(syscall.Handle(fd), e.handle, e.complateKey, 0)
	if err != nil {
		return err
	}
	e.eventAdapter.Link(uint64(handle), evh)
	return nil
}

//Remove .
func (e *netPoll) Remove(fd uint64) {
	e.eventAdapter.UnLink(uint64(fd))
}

//Polling .
func (e *netPoll) Polling() {
	var (
		isWriteEvent = func(events uint32) bool {
			return false
		}
		isReadEvent = func(events uint32) bool {
			return true
		}
	)
	e.polling(func(fd uint64, event uint32) error {
		evh := e.eventAdapter.Get(fd)
		if evh == nil {
			fmt.Printf("goserver.netpoll_linux.go: no fd %d \n", fd)
			return nil
		}
		//在协程池中运行要保证同一个通道下的通信是串行的
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
	fmt.Println("=========================")
	bufLen := uint32(0)
	overlapped := new(syscall.Overlapped)
	for {
		fmt.Println("1")
		err := syscall.GetQueuedCompletionStatus(e.handle, &bufLen, &e.complateKey, &overlapped, 1000) // syscall.INFINITE
		fmt.Println("2")
		if err != nil {
			fmt.Printf("syscall.GetQueuedCompletionStatus: %v\n", err)
		}
		if overlapped == nil {
			continue
		}
		fmt.Println(overlapped)
	}
}
