package goserver

import "context"

//Handle 处理类
type Handle interface {
	ReadPacket(ctx context.Context, conn Conn, next func(context.Context)) Packet      //读取包
	OnConnection(ctx context.Context, conn Conn, next func(context.Context))           //连接建立时处理
	OnMessage(ctx context.Context, conn Conn, p Packet, next func(context.Context))    //每次获取到消息时处理
	OnRecvError(ctx context.Context, conn Conn, err error, next func(context.Context)) //连接数据接收异常
	OnRecvTimeOut(ctx context.Context, conn Conn, next func(context.Context))          //接收数据超时处理
	OnHandTimeOut(ctx context.Context, conn Conn, next func(context.Context))          //处理数据超时处理
	OnClose(ctx context.Context, state *ConnState, next func(context.Context))         //连接关闭时处理
	OnPanic(ctx context.Context, conn Conn, err error, next func(context.Context))     //Panic时处理
}

//CoreHandle 包装接口实现类
type CoreHandle struct {
	handle Handle
	prev   *CoreHandle
	next   *CoreHandle
}

//NewCoreHandle 实例化
//@h 连接处理程序接口
//@return 返回实例
func NewCoreHandle(h Handle) *CoreHandle {
	return &CoreHandle{handle: h}
}

//NextHandle 链式调用
//方法解读:框架希望实现类似管道处理，类似AOP的执行效果，
//虽然TCPHandle接口的实现类可以在应用层通过装饰器等其他手段实现AOP效果，但如果封装进框架中会有难度。
//这里采用链表的形式包装每个传进来的TCPHandle接口，并实现链式查找。
//框架遇到的问题是，在接口实现的方法内部，并不知道当前是哪个方法，所以希望这个接口的当前方法调用下一个接口的当前方法，有困难。
//这里采用闭包的形式，在框架的每一处调用接口方法的地方都创建一个闭包，并传入回调，把返回的方法再次传递给接口方法，
//那么每个接口方法的实现通过调用传递进去的方法，能各自访问各自创建的闭包，从而实现管道调用之间不会互相影响，
//至此完成管道的处理流程，关键是闭包的应用
func (h *CoreHandle) NextHandle(callback func(*CoreHandle, func())) {
	var next func()
	next = func() {
		if h.next != nil {
			h = h.next
			callback(h.prev, next)
		} else {
			callback(h, next)
		}
	}
	next()
}

//Link 为当前节点连接并返回下一个节点
func (h *CoreHandle) Link(next *CoreHandle) *CoreHandle {
	h.next = next
	next.prev = h
	return next
}

//First 获取传入节点链路中第一个节点
func First(curr *CoreHandle) *CoreHandle {
	for {
		if curr.prev != nil {
			curr = curr.prev
		} else {
			break
		}
	}
	return curr
}

//Last 获取传入节点链路中最后一个节点
func Last(curr *CoreHandle) *CoreHandle {
	for {
		if curr.next != nil {
			curr = curr.next
		} else {
			break
		}
	}
	return curr
}

//Next 获取当前节点的下一个节点
func (h *CoreHandle) Next() *CoreHandle { return h.next }

//Prev 获取当前节点的上一个节点
func (h *CoreHandle) Prev() *CoreHandle { return h.prev }

//ReadPacket .
func (h *CoreHandle) ReadPacket(ctx context.Context, conn Conn, next func(context.Context)) Packet {
	p := h.handle.ReadPacket(ctx, conn, next)
	return p
}

//OnConnection .
func (h *CoreHandle) OnConnection(ctx context.Context, conn Conn, next func(context.Context)) {
	h.handle.OnConnection(ctx, conn, next)
}

//OnMessage .
func (h *CoreHandle) OnMessage(ctx context.Context, conn Conn, p Packet, next func(context.Context)) {
	h.handle.OnMessage(ctx, conn, p, next)
}

//OnClose .
func (h *CoreHandle) OnClose(ctx context.Context, state *ConnState, next func(context.Context)) {
	h.handle.OnClose(ctx, state, next)
}

//OnRecvTimeOut .
func (h *CoreHandle) OnRecvTimeOut(ctx context.Context, conn Conn, next func(context.Context)) {
	h.handle.OnRecvTimeOut(ctx, conn, next)
}

//OnPanic .
func (h *CoreHandle) OnPanic(ctx context.Context, conn Conn, err error, next func(context.Context)) {
	h.handle.OnPanic(ctx, conn, err, next)
}

//OnRecvError .
func (h *CoreHandle) OnRecvError(ctx context.Context, conn Conn, err error, next func(context.Context)) {
	h.handle.OnRecvError(ctx, conn, err, next)
}

//OnHandTimeOut .
func (h *CoreHandle) OnHandTimeOut(ctx context.Context, conn Conn, next func(context.Context)) {
	h.handle.OnHandTimeOut(ctx, conn, next)
}

//EmptyHandle .
type EmptyHandle struct {
	handle Handle
}

//ReadPacket .
func (h *EmptyHandle) ReadPacket(ctx context.Context, conn Conn, next func(context.Context)) Packet {
	return nil
}

//OnConnection .
func (h *EmptyHandle) OnConnection(ctx context.Context, conn Conn, next func(context.Context)) {
}

//OnMessage .
func (h *EmptyHandle) OnMessage(ctx context.Context, conn Conn, p Packet, next func(context.Context)) {
}

//OnClose .
func (h *EmptyHandle) OnClose(ctx context.Context, state *ConnState, next func(context.Context)) {
}

//OnRecvTimeOut .
func (h *EmptyHandle) OnRecvTimeOut(ctx context.Context, conn Conn, next func(context.Context)) {
}

//OnPanic .
func (h *EmptyHandle) OnPanic(ctx context.Context, conn Conn, err error, next func(context.Context)) {
}

//OnRecvError .
func (h *EmptyHandle) OnRecvError(ctx context.Context, conn Conn, err error, next func(context.Context)) {
}

//OnHandTimeOut .
func (h *EmptyHandle) OnHandTimeOut(ctx context.Context, conn Conn, next func(context.Context)) {
}

//BaseHandle .
type BaseHandle struct {
	Handle
}

//ReadPacket .
func (h *BaseHandle) ReadPacket(ctx context.Context, conn Conn, next func(context.Context)) Packet {
	next(ctx)
	return nil
}

//OnConnection .
func (h *BaseHandle) OnConnection(ctx context.Context, conn Conn, next func(context.Context)) {
	next(ctx)
}

//OnMessage .
func (h *BaseHandle) OnMessage(ctx context.Context, conn Conn, p Packet, next func(context.Context)) {
	next(ctx)
}

//OnRecvError .
func (h *BaseHandle) OnRecvError(ctx context.Context, conn Conn, err error, next func(context.Context)) {
	next(ctx)
}

//OnRecvTimeOut .
func (h *BaseHandle) OnRecvTimeOut(ctx context.Context, conn Conn, next func(context.Context)) {
	next(ctx)
}

//OnHandTimeOut .
func (h *BaseHandle) OnHandTimeOut(ctx context.Context, conn Conn, next func(context.Context)) {
	next(ctx)
}

//OnClose .
func (h *BaseHandle) OnClose(ctx context.Context, state *ConnState, next func(context.Context)) {
	next(ctx)
}

//OnPanic .
func (h *BaseHandle) OnPanic(ctx context.Context, conn Conn, err error, next func(context.Context)) {
	next(ctx)
}
