package goserver

type (
	//Handle 处理类
	Handle interface {
		ReadPacket(ctx ReadContext) Packet    //read a packet
		OnConnection(ctx ConnectionContext)   //handle on client is connect
		OnMessage(ctx Context)                //handle on read a packet complate
		OnRecvTimeOut(ctx RecvTimeOutContext) //handle on recv data timeout
		OnClose(ctx CloseContext)             //handle on close connect
		OnPanic(ctx PanicContext)             //handle on panic
	}

	//BaseHandle .
	BaseHandle struct{}
)

var _ Handle = (*BaseHandle)(nil)

//ReadPacket .
func (h *BaseHandle) ReadPacket(ctx ReadContext) Packet {
	ctx.Next()
	return nil
}

//OnConnection .
func (h *BaseHandle) OnConnection(ctx ConnectionContext) {
	ctx.Next()
}

//OnMessage .
func (h *BaseHandle) OnMessage(ctx Context) {
	ctx.Next()
}

//OnRecvTimeOut .
func (h *BaseHandle) OnRecvTimeOut(ctx RecvTimeOutContext) {
	ctx.Next()
}

//OnClose .
func (h *BaseHandle) OnClose(ctx CloseContext) {
	ctx.Next()
}

//OnPanic .
func (h *BaseHandle) OnPanic(ctx PanicContext) {
	ctx.Next()
}
