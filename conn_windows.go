// +build windows

package goserver

import (
	"fmt"
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
				if temp != nil && p != nil {
					panicError(errReadPacket.Error())
				}
				if temp != nil && p == nil {
					p = temp
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
		handTimer := time.NewTimer(c.option.HandTimeOut)
		defer handTimer.Stop()
		pch := c.readPacket(size)
		hch := make(chan struct{}, c.option.MaxWaitCountByHandTimeOut) //防止OnMessage协程堆积
		defer close(hch)
		for c.rwc != nil {
			recvTimer.Reset(c.option.RecvTimeOut)
			select {
			case <-c.ctx.Done():
				return
			case <-recvTimer.C:
				c.pipe.schedule(func(h Handle, ctx interface{}) { h.OnRecvTimeOut(ctx.(RecvTimeOutContext)) }, newRecvTimeOutContext(c))
			case p := <-pch:
				select {
				case <-c.ctx.Done():
					return
				case hch <- struct{}{}:
					sign := make(chan struct{})
					go func() {
						defer func() {
							select {
							case <-sign:
							default:
								close(sign)
							}
						}()
						c.pipe.schedule(func(h Handle, ctx interface{}) { h.OnMessage(ctx.(Context)) }, newContext(c, p))
						select {
						case <-c.ctx.Done():
							return
						case <-hch:
						}
					}()
					handTimer.Reset(c.option.HandTimeOut)
					select {
					case <-sign:
					case <-handTimer.C:
						c.pipe.schedule(func(h Handle, ctx interface{}) { h.OnHandTimeOut(ctx.(Context)) }, newContext(c, p))
					}
				}
			}
		}
	})
}
