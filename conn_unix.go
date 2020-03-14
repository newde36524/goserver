// +build linux darwin netbsd freebsd openbsd dragonfly

package goserver

import (
	"context"
	"errors"
	"time"
)

var (
	errReadPacket = errors.New("禁止在管道链路中重复读取生成Packet,在管道中读取数据帧,只能有一个管道返回Packet,其余只能返回nil")
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
			if p == nil { // read the first of packet from handle pipe only
				p = h.ReadPacket(ctx, c, next)
			} else {
				panicError(errReadPacket.Error())
			}
		})
	})
	return p
}
