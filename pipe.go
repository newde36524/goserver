package goserver

import (
	"context"
)

type (
	//Pipe .
	Pipe interface {
		Regist(h Handle) Pipe
		schedule(fn func(Handle, context.Context, func(context.Context)))
	}

	pipeLine struct {
		context context.Context
		handles []Handle
	}
)

//newPipe .
func newPipe(ctx context.Context) Pipe {
	return &pipeLine{
		context: ctx,
	}
}

func (p *pipeLine) Regist(h Handle) Pipe {
	p.handles = append(p.handles, h)
	return p
}

//schedule pipeline provider
func (p *pipeLine) schedule(fn func(Handle, context.Context, func(context.Context))) {
	index := 0
	var next func(context.Context)
	next = func(ctx context.Context) {
		if index < len(p.handles) {
			index++
			fn(p.handles[index-1], ctx, next)
		}
	}
	next(p.context)
	return
}
