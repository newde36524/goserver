// +build windows

package goserver

import (
	"fmt"
	"net"
	"time"
)

//Run start run server and receive and handle and send packet
func (c *Conn) Run() {
	c.safeFn(func() {
		c.pipe.schedule(func(h Handle, ctx interface{}) { h.OnConnection(ctx.(ConnectionContext)) }, newConnectionContext(c))
		c.recv(1)
	})
}

//readPacket read a packet
func (c *Conn) readPacket(size int) <-chan Packet {
	result := make(chan Packet, size)
	go c.safeFn(func() {
		defer close(result)
		for {
			var p Packet
			c.pipe.schedule(func(h Handle, ctx interface{}) {
				temp := h.ReadPacket(ctx.(ReadContext))
				//防止内部调用next()方法重复覆盖p的值
				//当前机制保证在管道处理流程中,只要有一个handle的ReadPacket方法返回值不为nil时才有效,之后无效
				if p == nil {
					p = temp
				} else if temp != nil {
					panicError(errReadPacket.Error())
				}

			}, newReadContext(c))
			select {
			case <-c.ctx.Done():
				return
			case result <- p:
				c.state.RecvPacketCount++
			}
		}
	})
	return result
}

//recv create a receive only packet channel
func (c *Conn) recv(size int) {
	go c.safeFn(func() {
		defer func() {
			if isDebug {
				logError(fmt.Sprintf("%s: recv goruntinue exit", c.RemoteAddr()))
			}
		}()
		recvTimer := time.NewTimer(c.option.RecvTimeOut)
		defer recvTimer.Stop()
		pch := c.readPacket(size)
		for c.rwc != nil {
			recvTimer.Reset(c.option.RecvTimeOut)
			select {
			case <-c.ctx.Done():
				return
			case <-recvTimer.C:
				c.pipe.schedule(func(h Handle, ctx interface{}) { h.OnRecvTimeOut(ctx.(RecvTimeOutContext)) }, newRecvTimeOutContext(c))
			case p := <-pch: //通道关闭时p 为 nil,或者ReadPacket()读取异常返回nil
				if p == nil {
					break
				}
				c.pipe.schedule(func(h Handle, ctx interface{}) { h.OnMessage(ctx.(MessageContext)) }, newMessageContext(c, p))
			}
		}
	})
}

//SetNoDelay .
func (c *Conn) SetNoDelay(b bool) {
	switch v := c.rwc.(type) {
	case *net.TCPConn:
		v.SetNoDelay(b)
	default:
	}
}
