// +build linux darwin netbsd freebsd openbsd dragonfly

package goserver

import (
	"context"
	"errors"
	"time"
)

var (
	errReadPacket = errors.New("read the first of packet from handle pipe only")
)

//OnWriteable .
func (c Conn) OnWriteable() {
	logInfo("do nothing")
}

//OnReadable 服务端建立的连接处理方法
func (c Conn) OnReadable() {
	if p := c.readPacketOne(); p != nil {
		c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnMessage(ctx, c, p, next) })
	}
}

//OnRecvTimeHandler .
func (c Conn) OnRecvTimeHandler() {
	if time.Now().Sub(c.readTime) > c.option.RecvTimeOut {
		c.Close("conn recv data timeout")
	}
}

//readPacketOne .
func (c Conn) readPacketOne() Packet {
	var p Packet
	c.readTime = time.Now()
	c.safeFn(func() {
		c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) {
			if p == nil {
				p = h.ReadPacket(ctx, c, next)
			} else {
				panicError(errReadPacket.Error())
			}
		})
	})
	return p
}
