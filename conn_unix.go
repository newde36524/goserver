// +build linux darwin netbsd freebsd openbsd dragonfly

package goserver

import (
	"time"
)

//OnWriteable .
func (c *Conn) OnWriteable() {
	logInfo("do nothing")
}

//OnReadable 服务端建立的连接处理方法
func (c *Conn) OnReadable() {
	c.safeFn(func() {
		if p := c.readPacketOne(); p != nil {
			c.state.RecvPacketCount++
			c.pipe.schedule(func(h Handle, ctx interface{}) { h.OnMessage(ctx.(MessageContext)) }, newMessageContext(c, p))
		}
	})
}

//IsRecvTimeOut 检查是否接受数据的时间是否超过设置的接受最大时间
func (c *Conn) IsRecvTimeOut() bool {
	return time.Now().Sub(c.readTime) > c.option.RecvTimeOut
}

//readPacketOne .
func (c *Conn) readPacketOne() Packet {
	var result Packet
	c.readTime = time.Now()
	c.pipe.schedule(func(h Handle, ctx interface{}) {
		pac := h.ReadPacket(ctx.(ReadContext))
		if result == nil {
			result = pac
		} else if pac != nil {
			panicError(errReadPacket.Error())
		}
	}, newReadContext(c))
	return result
}
