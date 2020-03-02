// +build windows

package goserver

// type netPoll struct {
// 	handle       syscall.Handle
// 	events       []syscall.Overlapped
// 	eventAdapter eventAdapter
// 	gPool        *gPool
// }

// //newNetpoll .
// func newNetpoll(maxEvents int, gPool *gPool) *netPoll {
// 	handle, err := syscall.CreateIoCompletionPort(-1, 0, 0, 0)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return &netPoll{
// 		handle: handle,
// 		// events:       make([]syscall.EpollEvent, maxEvents),
// 		gPool:        gPool,
// 		eventAdapter: newdefaultAdapter(),
// 	}
// }

// //Register .
// func (e *netPoll) Regist(fd uint64, evh eventHandle) error {
// 	if handle, err := syscall.CreateIoCompletionPort(fd, e.handle, 0, 0); err != nil {
// 		return err
// 	} else {
// 		e.eventAdapter.Link(uint64(handle), evh)
// 		return nil
// 	}
// 	return nil
// }

// //Remove .
// func (e *netPoll) Remove(fd uint64) {
// 	e.eventAdapter.UnLink(uint64(fd))
// }

// //Polling .
// func (e *netPoll) Polling() {
// 	var (
// 		isWriteEvent = func(events uint32) bool {
// 			return false
// 		}
// 		isReadEvent = func(events uint32) bool {
// 			return false
// 		}
// 	)
// 	e.polling(func(fd uint64, event uint32) error {
// 		evh := e.eventAdapter.Get(fd)
// 		if evh == nil {
// 			fmt.Printf("goserver.netpoll_linux.go: no fd %d \n", fd)
// 			return nil
// 		}
// 		//在协程池中运行要保证同一个通道下的通信是串行的
// 		if isWriteEvent(event) {
// 			e.gPool.SchduleByKey(fd, evh.OnWriteable)
// 			// evh.OnWriteable()
// 		} else if isReadEvent(event) {
// 			e.gPool.SchduleByKey(fd, evh.OnReadable)
// 			// evh.OnReadable()
// 		}
// 		return nil
// 	})
// }

// func (e *netPoll) polling(onEventTrigger func(fd uint64, events uint32) error) {
// 	dwBytesTransfered := uint32(0)
// 	var ctxId uint32
// 	var overlapped **syscall.Overlapped
// 	for {
// 		err := syscall.GetQueuedCompletionStatus(e.handle, &dwBytesTransfered,
// 			&ctxId, overlapped, syscall.INFINITE)
// 		if err != nil {
// 			fmt.Printf("syscall.GetQueuedCompletionStatus: %v\n", err)
// 		}

// 		if overlapped == nil {
// 			continue
// 		}

// 	}
// }
