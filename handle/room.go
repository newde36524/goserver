package handle

import (
	"context"

	"github.com/newde36524/goserver"
)

//roomHandle .
type roomHandle struct {
	goserver.BaseHandle
	clientMap map[string]goserver.Conn
	ctxKey    string
}

//NewRoomHandle .
func NewRoomHandle(key string, initcap int64) goserver.Handle {
	return &roomHandle{
		clientMap: make(map[string]goserver.Conn, initcap),
		ctxKey:    key,
	}
}

//OnConnection .
func (h *roomHandle) OnConnection(ctx context.Context, conn goserver.Conn, next func(context.Context)) {
	if _, ok := h.clientMap[conn.RemoteAddr()]; !ok {
		h.clientMap[conn.RemoteAddr()] = conn
	}
	next(ctx)
}

//OnMessage .
func (h *roomHandle) OnMessage(ctx context.Context, conn goserver.Conn, p goserver.Packet, next func(context.Context)) {
	ctx = context.WithValue(ctx, h.ctxKey, h.clientMap)
	next(ctx)
}

//OnClose .
func (h *roomHandle) OnClose(ctx context.Context, state *goserver.ConnState, next func(context.Context)) {
	delete(h.clientMap, state.RemoteAddr)
	next(ctx)
}
