package handle

import (
	"context"

	"github.com/newde36524/goserver"
)

//roomHandle .
type roomHandle struct {
	goserver.BaseHandle
	clientMap map[string]goserver.Conn
}

func NewRoomHandle() goserver.Handle {
	return &roomHandle{
		clientMap: make(map[string]goserver.Conn, 1024),
	}
}

//OnConnection .
func (h *roomHandle) OnConnection(ctx context.Context, conn goserver.Conn, next func(context.Context)) {
	if _, ok := h.clientMap[conn.RemoteAddr()]; !ok {
		h.clientMap[conn.RemoteAddr()] = conn
	}
	key := "room"
	if ctx.Value(key) == nil {
		ctx = context.WithValue(ctx, key, h.clientMap)
	}
	next(ctx)
}

//OnClose .
func (h *roomHandle) OnClose(ctx context.Context, state *goserver.ConnState, next func(context.Context)) {
	delete(h.clientMap, state.RemoteAddr)
	next(ctx)
}
