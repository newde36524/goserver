package goserver

import (
	"context"
	"net"
	"runtime/debug"
	"time"
)

//Conn 连接代理对象
type Conn struct {
	rwc      net.Conn        //tcp原始连接对象
	option   ConnOption      //连接配置项
	state    *ConnState      //连接状态
	context  context.Context //全局上下文
	recvChan <-chan Packet   //接收消息队列
	sendChan chan<- Packet   //发送消息队列
	handChan chan<- Packet   //处理消息队列
	cancel   func()          //全局上下文取消函数
	isDebug  bool            //是否打印框架内部debug信息
	handles  []Handle        //连接处理程序管道
}

//NewConn returns a wrapper of raw conn
func NewConn(rwc net.Conn, option ConnOption, hs []Handle) (result *Conn) {
	result = &Conn{
		rwc:     rwc,
		option:  option,
		handles: hs,
		state: &ConnState{
			ActiveTime: time.Now(),
			RemoteAddr: rwc.RemoteAddr().String(),
		},
		isDebug: false,
	}
	result.context, result.cancel = context.WithCancel(context.Background())
	return
}

//Next 使用下一个处理程序
func (c *Conn) Next(fn func(Handle, func())) {
	index := 0
	var next func()
	next = func() {
		defer func() {
			if err := recover(); err != nil {
				c.option.Logger.Errorf("%s: Conn.Next: pipeline excute error: %s", c.RemoteAddr(), err)
				c.option.Logger.Error(string(debug.Stack()))
			}
		}()
		if index < len(c.handles) {
			index++
			fn(c.handles[index-1], next)
		}
	}
	next()
	return
}

//fnProxy 代理执行方法,用于检测执行超时
func (c *Conn) fnProxy(fn func()) <-chan struct{} {
	result := make(chan struct{}, 1)
	go func() {
		defer func() {
			close(result)
			if err := recover(); err != nil {
				defer recover()
				c.Next(func(h Handle, next func()) { h.OnPanic(c, err.(error), next) })
				c.option.Logger.Error(string(debug.Stack()))
			}
		}()
		fn()
		result <- struct{}{}
	}()
	return result
}

//safeFn 代理方法，用于安全调用方法，恢复panic
func (c *Conn) safeFn(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			defer recover()
			c.Next(func(h Handle, next func()) { h.OnPanic(c, err.(error), next) })
			c.option.Logger.Error(string(debug.Stack()))
		}
	}()
	fn()
}

//UseDebug 打开框架内部Debug信息
func (c *Conn) UseDebug() {
	c.isDebug = true
}

//Read 从tcp连接中读取数据帧
func (c *Conn) Read(b []byte) (n int, err error) {
	c.rwc.SetReadDeadline(time.Now().Add(c.option.ReadDataTimeOut))
	n, err = c.rwc.Read(b)
	if err != nil {
		c.Next(func(h Handle, next func()) { h.OnRecvError(c, err, next) })
	}
	return
}

//RemoteAddr 客户端IP地址
func (c *Conn) RemoteAddr() string {
	return c.rwc.RemoteAddr().String()
}

//LocalAddr 服务器IP地址
func (c *Conn) LocalAddr() string {
	return c.rwc.LocalAddr().String()
}

//Raw 获取原始连接
func (c *Conn) Raw() net.Conn {
	return c.rwc
}

//run 固定处理流程
/*
工作协程：
	1.数据发送协程
	2.数据接收协程
	3.数据处理协程
	4.3个心跳协助协程
	5.数据搬运代理协程(用于从接收信道接收数据并发送数据到处理信道)
持久channel：
	1.sendChan 发送信道
	2.recvChan 接收信道
	3.handChan 处理信道
短期协程创建时机:
	1.每次调用readPacket方法读取数据帧时
	2.



*/
func (c *Conn) run() {
	c.sendChan = c.send(c.option.MaxSendChanCount)(c.heartBeat(c.option.SendTimeOut, func() {
		c.Next(func(h Handle, next func()) { h.OnTimeOut(c, SendTimeOutCode, next) })
	}))
	c.handChan = c.message(1)(c.heartBeat(c.option.HandTimeOut, func() {
		c.Next(func(h Handle, next func()) { h.OnTimeOut(c, HandTimeOutCode, next) })
	}))
	go c.safeFn(func() {
		select {
		case <-c.fnProxy(func() {
			c.Next(func(h Handle, next func()) { h.OnConnection(c, next) })
		}):
		case <-time.After(c.option.SendTimeOut):
			c.option.Logger.Debugf("%s: Conn.run: OnConnection funtion invoke used time was too long", c.RemoteAddr())
		}
		c.recvChan = c.recv(c.option.MaxRecvChanCount)(c.heartBeat(c.option.RecvTimeOut, func() {
			c.Next(func(h Handle, next func()) { h.OnTimeOut(c, RecvTimeOutCode, next) })
		}))
		defer func() {
			close(c.handChan)
			if c.isDebug {
				c.option.Logger.Debugf("%s: Conn.run: handChan is closed", c.RemoteAddr())
			}
			close(c.sendChan)
			if c.isDebug {
				c.option.Logger.Debugf("%s: Conn.run: sendChan is closed", c.RemoteAddr())
				c.option.Logger.Debugf("%s: Conn.run: proxy goruntinue exit", c.RemoteAddr())
			}
		}()
		for {
			select {
			case <-c.context.Done():
				return
			case p, ok := <-c.recvChan:
				if !ok {
					c.option.Logger.Errorf("%s: Conn.run: recvChan is closed", c.RemoteAddr())
				}
				select {
				case <-c.context.Done():
					return
				case c.handChan <- p:
				}
			}
		}
	})
}

//Write 发送消息到客户端
func (c *Conn) Write(packet Packet) {
	if packet == nil {
		c.option.Logger.Errorf("%s: Conn.Write: packet is nil,do nothing", c.RemoteAddr())
		return
	}
	select {
	case <-c.context.Done():
		return
	case c.sendChan <- packet:
	}
}

//Close 关闭服务器和客户端的连接
func (c *Conn) Close() {
	defer c.rwc.Close()
	c.rwc.SetReadDeadline(time.Time{})  //set read timeout
	c.rwc.SetWriteDeadline(time.Time{}) //set write timeout
	c.state.Message = "conn is closed"
	c.state.ComplateTime = time.Now()
	c.Next(func(h Handle, next func()) { h.OnClose(c.state, next) })

	c.cancel()
	// runtime.GC()         //强制GC      待定可能有问题
	// debug.FreeOSMemory() //强制释放内存 待定可能有问题
}

//readPacket 读取一个包
func (c *Conn) readPacket() <-chan Packet {
	result := make(chan Packet)
	go c.safeFn(func() {
		defer func() {
			close(result)
		}()
		select {
		case <-c.context.Done():
			return
		default:
		}
		var p Packet
		c.Next(func(h Handle, next func()) {
			_p := h.ReadPacket(c, next)
			//防止内部调用next()方法重复覆盖p的值
			//当前机制保证在管道处理流程中,只要有一个handle的ReadPacket方法返回值不为nil时才有效,之后无效
			if _p != nil && p != nil {
				panic("禁止在管道链路中重复读取生成Packet,在管道中读取数据帧，只能有一个管道返回Packet，其余只能返回nil")
			}
			if _p != nil && p == nil {
				p = _p
			}
		})

		result <- p
	})
	return result
}

//recv 创建一个可接收 packet channel
func (c *Conn) recv(maxRecvChanCount int) func(<-chan struct{}) <-chan Packet {
	return func(heartBeat <-chan struct{}) <-chan Packet {
		result := make(chan Packet, maxRecvChanCount)
		go c.safeFn(func() {
			defer func() {
				close(result)
				if c.isDebug {
					c.option.Logger.Debugf("%s: Conn.recv: recvChan is closed", c.RemoteAddr())
					c.option.Logger.Debugf("%s: Conn.recv: recv goruntinue exit", c.RemoteAddr())
				}
			}()
			for c.rwc != nil {
				ch := c.readPacket()
				select {
				case <-c.context.Done():
					return
				case result <- <-ch:
					c.state.RecvPacketCount++
					if c.isDebug {
						c.option.Logger.Debugf("%s: Conn.recv: read a packet", c.RemoteAddr())
					}
					select {
					case <-heartBeat:
					default:
					}
				}
			}
		})
		return result
	}
}

//send 创建一个可发送 packet channel
func (c *Conn) send(maxSendChanCount int) func(<-chan struct{}) chan<- Packet {
	return func(heartBeat <-chan struct{}) chan<- Packet {
		result := make(chan Packet, maxSendChanCount)
		go c.safeFn(func() {
			defer func() {
				if c.isDebug {
					c.option.Logger.Debugf("%s: Conn.send: send goruntinue exit", c.RemoteAddr())
				}
			}()
			for c.rwc != nil {
				select {
				case <-c.context.Done():
					return
				case packet, ok := <-result:
					c.state.SendPacketCount++
					if !ok {
						if c.isDebug {
							c.option.Logger.Errorf("%s: Conn.send: send packet chan was closed", c.RemoteAddr())
						}
						return
					}
					if packet == nil {
						c.option.Logger.Errorf("%s: Conn.send: the send packet is nil", c.RemoteAddr())
						break
					}
					sendData, err := packet.Serialize(nil)
					if err != nil {
						c.option.Logger.Error(err)
					}
					c.rwc.SetWriteDeadline(time.Now().Add(c.option.WriteDataTimeOut))
					_, err = c.rwc.Write(sendData)
					if err != nil {
						c.Next(func(h Handle, next func()) { h.OnSendError(c, packet, err, next) })
					} else {
						if c.isDebug {
							c.option.Logger.Debugf("%s: Conn.send: send a packet", c.RemoteAddr())
						}
					}
					select {
					case <-heartBeat:
					default:
					}
				}
			}
		})
		return result
	}
}

//message 创建一个可发送 hand packet channel
func (c *Conn) message(maxHandNum int) func(<-chan struct{}) chan<- Packet {
	return func(heartBeat <-chan struct{}) chan<- Packet {
		result := make(chan Packet, maxHandNum)
		go c.safeFn(func() {
			defer func() {
				if c.isDebug {
					c.option.Logger.Debugf("%s: Conn.message: hand goruntinue exit", c.RemoteAddr())
				}
			}()
			for {
				select {
				case <-c.context.Done():
					return
				case p, ok := <-result:
					if !ok {
						c.option.Logger.Errorf("%s: Conn.message: hand packet chan was closed", c.RemoteAddr())
						return
					}
					c.Next(func(h Handle, next func()) { h.OnMessage(c, p, next) })
					if c.isDebug {
						c.option.Logger.Debugf("%s: Conn.message: hand a packet", c.RemoteAddr())
					}
					select {
					case <-heartBeat:
					default:
					}
				}
			}
		})
		return result
	}
}

//heartBeat 协程心跳检测
func (c *Conn) heartBeat(timeOut time.Duration, callback func()) <-chan struct{} {
	result := make(chan struct{})
	go func() {
		defer func() {
			close(result)
			if c.isDebug {
				c.option.Logger.Debugf("%s: Conn.heartBeat: heartBeat goruntinue exit", c.RemoteAddr())
			}
		}()
		for {
			select {
			case result <- struct{}{}:
			case <-c.context.Done():
				return
			case <-time.After(timeOut):
				callback()
			}
		}
	}()
	return result
}
