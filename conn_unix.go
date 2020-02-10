// +build linux

package goserver

/*
工作协程：
	1.数据接收协程
持久channel：
	1.recvChan 接收信道
短期协程创建时机:
	1.每次调用readPacket方法读取数据帧时
	2.
框架现存的问题:
	1.消息前后的关联处理不方便 20190914
	2.代码比较粗糙 20190914
*/
import (
	"context"
	"net"
	"time"
)

//Conn net.Conn proxy object
type Conn struct {
	rwc     net.Conn        //row connection
	option  ConnOption      //connection option object
	state   *ConnState      //connection state
	context context.Context //global context
	cancel  func()          //global context cancel function
	isDebug bool            //is open inner debug message flag
	pipe    Pipe            //connection handle pipeline
}

//NewConn return a wrap of raw conn
func NewConn(ctx context.Context, rwc net.Conn, option ConnOption, hs []Handle) Conn {
	result := Conn{
		rwc:    rwc,
		option: option,
		state: &ConnState{
			ActiveTime: time.Now(),
			RemoteAddr: rwc.RemoteAddr().String(),
		},
	}
	result.valid()
	result.context, result.cancel = context.WithCancel(ctx)
	result.pipe = NewPipe(result.context, hs)
	return result
}

//OnWriteable .
func (c Conn) OnWriteable() {
	c.option.Logger.Info("conn_unix.go: conn OnWriteable")
}

//OnReadable 服务端建立的连接处理方法
func (c Conn) OnReadable() {
	pch := <-c.readPacketOne()
	if pch != nil {
		c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnMessage(ctx, c, pch, next) })
	}
}

//readPacketOne .
func (c Conn) readPacketOne() <-chan Packet {
	result := make(chan Packet, 1)
	c.safeFn(func() {
		defer close(result)
		var p Packet
		c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) {
			temp := h.ReadPacket(ctx, c, next)
			//防止内部调用next()方法重复覆盖p的值
			//当前机制保证在管道处理流程中,只要有一个handle的ReadPacket方法返回值不为nil时才有效,之后无效
			if temp != nil && p != nil {
				panic("goserver.Conn.readPacket: 禁止在管道链路中重复读取生成Packet,在管道中读取数据帧,只能有一个管道返回Packet,其余只能返回nil")
			}
			if temp != nil && p == nil {
				p = temp
			}
		})
		select {
		case <-c.context.Done():
			return
		case result <- p:
			c.state.RecvPacketCount++
		}
	})
	return result
}
