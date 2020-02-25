// +build linux darwin netbsd freebsd openbsd dragonfly

package goserver

import (
	"context"
	"time"
)

//OnWriteable .
func (c Conn) OnWriteable() {
	c.option.Logger.Info("goserver.conn_unix.go: do nothing")
}

//OnReadable 服务端建立的连接处理方法
func (c Conn) OnReadable() {
	c.readTime = time.Now()
	pch := <-c.readPacketOne()
	if pch != nil {
		c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnMessage(ctx, c, pch, next) })
	}
}

//OnRecvTimeHandler .
func (c Conn) OnRecvTimeHandler() {
	if time.Now().Sub(c.readTime) > c.option.RecvTimeOut {
		c.Close("goserver.conn_unix.go: conn recv data timeout")
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
		case <-c.ctx.Done():
			return
		case result <- p:
			c.state.RecvPacketCount++
		}
	})
	return result
}
