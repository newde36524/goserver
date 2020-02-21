// +build windows

package goserver

import (
	"context"
	"time"
)

//Run start run server and receive and handle and send packet
func (c Conn) Run() {
	go c.safeFn(func() {
		c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnConnection(ctx, c, next) })
		c.recv(1)
	})
}

//readPacket read a packet
func (c Conn) readPacket(size int) <-chan Packet {
	result := make(chan Packet, size)
	go c.safeFn(func() {
		defer close(result)
		for {
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
		}
	})
	return result
}

//recv create a receive only packet channel
func (c Conn) recv(size int) {
	go c.safeFn(func() {
		defer func() {
			// fmt.Println("===========================")
			if c.isDebug {
				c.option.Logger.Debugf("%s: goserver.Conn.recv: recv goruntinue exit", c.RemoteAddr())
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
				c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnRecvTimeOut(ctx, c, next) })
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
						c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnMessage(ctx, c, p, next) })
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
						c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnHandTimeOut(ctx, c, next) })
					}
				}
			}
		}
	})
}
