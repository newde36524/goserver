package goserver

import "context"

type (
	//Handle 处理类
	Handle interface {
		ReadPacket(ctx context.Context, c Conn, next func(context.Context)) Packet   //读取包
		OnConnection(ctx context.Context, c Conn, next func(context.Context))        //连接建立时处理
		OnMessage(ctx context.Context, c Conn, p Packet, next func(context.Context)) //每次获取到消息时处理
		OnRecvTimeOut(ctx context.Context, c Conn, next func(context.Context))       //接收数据超时处理
		OnHandTimeOut(ctx context.Context, c Conn, next func(context.Context))       //处理数据超时处理
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
