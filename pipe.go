package goserver

import (
	"context"
)

type (
	//Pipe .
	Pipe interface {
		Regist(h Handle) Pipe
		schedule(fn func(h Handle, ctx interface{}), ctx interface{})
	}

	//pipeLine .
	pipeLine struct {
		ctx     context.Context
		handles []Handle
	}

	//canNext .
	canNext interface {
		setNext(next func())
	}
)

//Regist .
func (p *pipeLine) Regist(h Handle) Pipe {
	p.handles = append(p.handles, h)
	return p
}

//schedule pipeline provider
func (p *pipeLine) schedule(fn func(h Handle, ctx interface{}), ctx interface{}) {
	index := 0
	var next func()
	next = func() {
		if index < len(p.handles) {
			index++
			fn(p.handles[index-1], ctx)
		}
	}
	if v, ok := ctx.(canNext); ok {
		v.setNext(next)
	}
	next()
	return
}
