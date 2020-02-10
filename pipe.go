package goserver

import (
	"context"
)

//Pipe .
type Pipe interface {
	schedule(fn func(Handle, context.Context, func(context.Context)))
}

type pipeLine struct {
	index   int
	handles []Handle
	context context.Context
}

//NewPipe .
func NewPipe(ctx context.Context, hs []Handle) Pipe {
	return &pipeLine{
		index:   0,
		context: ctx,
		handles: hs,
	}
}

//schedule pipeline provider
func (p *pipeLine) schedule(fn func(Handle, context.Context, func(context.Context))) {
	p.index = 0
	var next func(context.Context)
	next = func(ctx context.Context) {
		if p.index < len(p.handles) {
			p.index++
			fn(p.handles[p.index-1], ctx, next)
		}
	}
	next(p.context)
	return
}
