package goserver

import "context"

type (
	//Handle 处理类
	Handle interface {
		ReadPacket(ctx context.Context, c Conn, next func(context.Context)) Packet   //read a packet
		OnConnection(ctx context.Context, c Conn, next func(context.Context))        //handle on client is connect
		OnMessage(ctx context.Context, c Conn, p Packet, next func(context.Context)) //handle on read a packet complate
		OnRecvTimeOut(ctx context.Context, c Conn, next func(context.Context))       //handle on recv data timeout
		OnHandTimeOut(ctx context.Context, c Conn, next func(context.Context))       //handle on handle data timeout
		OnClose(ctx context.Context, state *ConnState, next func(context.Context))   //连接关闭时处理
		OnPanic(ctx context.Context, c Conn, err error, next func(context.Context))  //Panic时处理
	}

	//BaseHandle .
	BaseHandle struct{}
)

//ReadPacket .
func (h *BaseHandle) ReadPacket(ctx context.Context, c Conn, next func(context.Context)) Packet {
	next(ctx)
	return nil
}

//OnConnection .
func (h *BaseHandle) OnConnection(ctx context.Context, c Conn, next func(context.Context)) {
	next(ctx)
}

//OnMessage .
func (h *BaseHandle) OnMessage(ctx context.Context, c Conn, p Packet, next func(context.Context)) {
	next(ctx)
}

//OnRecvTimeOut .
func (h *BaseHandle) OnRecvTimeOut(ctx context.Context, c Conn, next func(context.Context)) {
	next(ctx)
}

//OnHandTimeOut .
func (h *BaseHandle) OnHandTimeOut(ctx context.Context, c Conn, next func(context.Context)) {
	next(ctx)
}

//OnClose .
func (h *BaseHandle) OnClose(ctx context.Context, state *ConnState, next func(context.Context)) {
	next(ctx)
}

//OnPanic .
func (h *BaseHandle) OnPanic(ctx context.Context, c Conn, err error, next func(context.Context)) {
	next(ctx)
}
