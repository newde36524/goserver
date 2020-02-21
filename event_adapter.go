package goserver

import "sync"

type (
	//eventHandle .
	eventHandle interface {
		OnReadable()
		OnWriteable()
	}

	//eventAdapter .
	eventAdapter interface {
		Get(fd int32) eventHandle
		Link(fd int32, evh eventHandle)
		UnLink(fd int32)
	}

	defaultAdapter struct {
		eventAdapter
		sync.Map
	}
)

func newdefaultAdapter() *defaultAdapter {
	return &defaultAdapter{}
}

//Get .
func (e *defaultAdapter) Get(fd int32) eventHandle {
	if o, ok := e.Load(fd); ok {
		return o.(eventHandle)
	}
	return nil
}

//Link .
func (e *defaultAdapter) Link(fd int32, evh eventHandle) {
	e.Store(fd, evh)
}

//UnLink .
func (e *defaultAdapter) UnLink(fd int32) {
	e.Delete(fd)
}
