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
		Get(fd uint64) eventHandle
		Link(fd uint64, evh eventHandle)
		UnLink(fd uint64)
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
func (e *defaultAdapter) Get(fd uint64) eventHandle {
	if o, ok := e.Load(fd); ok {
		return o.(eventHandle)
	}
	return nil
}

//Link .
func (e *defaultAdapter) Link(fd uint64, evh eventHandle) {
	e.Store(fd, evh)
}

//UnLink .
func (e *defaultAdapter) UnLink(fd uint64) {
	e.Delete(fd)
}
