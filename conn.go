package goserver

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"time"
)

//Conn net.Conn proxy object
type Conn struct {
	rwc      net.Conn        //row connection
	option   ConnOption      //connection option object
	state    *ConnState      //connection state
	context  context.Context //global context
	recvChan <-chan Packet   //receive packet chan
	sendChan chan<- Packet   //send packet chan
	handChan chan<- Packet   //hand packet chan
	cancel   func()          //global context cancel function
	isDebug  bool            //is open inner debug message flag
	handles  []Handle        //connection handle pipeline
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

//pipe pipeline provider
func (c *Conn) pipe(fn func(Handle, func())) {
	index := 0
	var next func()
	next = func() {
		defer func() {
			if err := recover(); err != nil {
				if c.option.Logger != nil {
					c.option.Logger.Errorf("%s: goserver.Conn.Next: pipeline excute error: %s", c.RemoteAddr(), err)
					c.option.Logger.Error(string(debug.Stack()))
				}
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

//fnProxy proxy agent,used to checking invoke timeout and recover panic
func (c *Conn) fnProxy(fn func()) <-chan struct{} {
	result := make(chan struct{}, 1)
	go func() {
		defer func() {
			close(result)
			if err := recover(); err != nil {
				defer recover()
				c.pipe(func(h Handle, next func()) { h.OnPanic(c, err.(error), next) })
				if c.option.Logger != nil {
					c.option.Logger.Errorf("goserver.Conn.fnProxy: %s", err)
					c.option.Logger.Errorf("goserver.Conn.fnProxy: %s", string(debug.Stack()))
				}
			}
		}()
		fn()
		result <- struct{}{}
	}()
	return result
}

//safeFn proxy agent,used to safe invoke and recover panic
func (c *Conn) safeFn(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			defer recover()
			c.pipe(func(h Handle, next func()) { h.OnPanic(c, err.(error), next) })
			if c.option.Logger != nil {
				c.option.Logger.Errorf("goserver.Conn.safeFn: %s", err)
				c.option.Logger.Errorf("goserver.Conn.safeFn: %s", string(debug.Stack()))
			}
		}
	}()
	fn()
}

//UseDebug open inner debug message
func (c *Conn) UseDebug() {
	c.isDebug = true
	if c.option.Logger == nil {
		fmt.Println("goserver.Conn.UseDebug: c.option.Logger is nil")
	}
}

//Read read a data frame from connection
func (c *Conn) Read(b []byte) (n int, err error) {
	c.rwc.SetReadDeadline(time.Now().Add(c.option.ReadDataTimeOut))
	n, err = c.rwc.Read(b)
	if err != nil {
		c.pipe(func(h Handle, next func()) { h.OnRecvError(c, err, next) })
	}
	return
}

//RemoteAddr get client ip address
func (c *Conn) RemoteAddr() string {
	return c.rwc.RemoteAddr().String()
}

//LocalAddr get host ip address
func (c *Conn) LocalAddr() string {
	return c.rwc.LocalAddr().String()
}

//Raw get row connection
func (c *Conn) Raw() net.Conn {
	return c.rwc
}

//Write send a packet to remote connection
func (c *Conn) Write(packet Packet) {
	c.safeFn(func() {
		if packet == nil {
			if c.isDebug && c.option.Logger != nil {
				c.option.Logger.Debugf("%s: goserver.Conn.Write: packet is nil,do nothing", c.RemoteAddr())
			}
			return
		}
		select {
		case <-c.context.Done():
			return
		default:
		}
		select {
		case <-c.context.Done():
			return
		case c.sendChan <- packet:
		}
	})
}

//Close close connection
func (c *Conn) Close() {
	defer func() {
		select {
		case <-c.context.Done():
			c.rwc.SetDeadline(time.Now().Add(1 * time.Second)) //set deadline timeout 设置客户端链接超时，是至关重要的。否则，一个超慢或已消失的客户端，可能会泄漏文件描述符，并最终导致异常
			c.rwc.Close()
			c.pipe(func(h Handle, next func()) { h.OnClose(c.state, next) })
			// switch v := c.rwc.(type) {
			// case *net.TCPConn:
			// 	v.SetKeepAlive(false)
			// 	f, err := v.File()
			// 	if err != nil {
			// 		c.option.Logger.Errorf("goserver.Conn.Close: %s", err)
			// 	}
			// 	syscall.Shutdown(syscall.Handle(f.Fd()), 0)
			// case *net.UDPConn:
			// 	f, err := v.File()
			// 	if err != nil {
			// 		c.option.Logger.Errorf("goserver.Conn.Close: %s", err)
			// 	}
			// 	syscall.Shutdown(syscall.Handle(f.Fd()), 0)
			// case *net.UnixConn:
			// 	f, err := v.File()
			// 	if err != nil {
			// 		c.option.Logger.Errorf("goserver.Conn.Close: %s", err)
			// 	}
			// 	syscall.Shutdown(syscall.Handle(f.Fd()), 0)
			// case *net.IPConn:
			// 	f, err := v.File()
			// 	if err != nil {
			// 		c.option.Logger.Errorf("goserver.Conn.Close: %s", err)
			// 	}
			// 	syscall.Shutdown(syscall.Handle(f.Fd()), 0)
			// }
		}
	}()
	c.cancel()
	c.state.Message = "conn is closed"
	c.state.ComplateTime = time.Now()

	// runtime.GC()         //强制GC      待定可能有问题
	// debug.FreeOSMemory() //强制释放内存 待定可能有问题
}

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
框架现存的问题:
	1.单链接基础协程数太多,单机最大负载量受影响 20190914
	2.消息前后的关联处理不方便 20190914
	3.代码比较粗糙 20190914
*/
//Run start run server and receive and handle and send packet
func (c *Conn) Run() {
	sendHeartBeat := c.heartBeat(c.option.SendTimeOut, func() {
		c.pipe(func(h Handle, next func()) { h.OnTimeOut(c, SendTimeOutCode, next) })
	})
	c.sendChan = c.send(c.option.MaxSendChanCount, sendHeartBeat)
	go c.safeFn(func() {
		select {
		case <-c.fnProxy(func() {
			c.pipe(func(h Handle, next func()) { h.OnConnection(c, next) })
		}):
		case <-time.After(c.option.HandTimeOut):
			if c.isDebug && c.option.Logger != nil {
				c.option.Logger.Debugf("%s: goserver.Conn.run: the goserver.Handle.OnConnection function invoke time was too long", c.RemoteAddr())
			}
		}
		handHeartBeat := c.heartBeat(c.option.HandTimeOut, func() {
			c.pipe(func(h Handle, next func()) { h.OnTimeOut(c, HandTimeOutCode, next) })
		})
		c.handChan = c.message(1, handHeartBeat)
		recvHeartBeat := c.heartBeat(c.option.RecvTimeOut, func() {
			c.pipe(func(h Handle, next func()) { h.OnTimeOut(c, RecvTimeOutCode, next) })
		})
		c.recvChan = c.recv(c.option.MaxRecvChanCount, recvHeartBeat)
		defer func() {
			close(c.handChan)
			if c.isDebug && c.option.Logger != nil {
				c.option.Logger.Debugf("%s: goserver.Conn.run: handChan is closed", c.RemoteAddr())
			}
			close(c.sendChan)
			if c.isDebug && c.option.Logger != nil {
				c.option.Logger.Debugf("%s: goserver.Conn.run: sendChan is closed", c.RemoteAddr())
				c.option.Logger.Debugf("%s: goserver.Conn.run: proxy goruntinue exit", c.RemoteAddr())
			}
		}()
		for {
			select {
			case <-c.context.Done():
				return
			case p, ok := <-c.recvChan:
				if !ok {
					if c.isDebug && c.option.Logger != nil {
						c.option.Logger.Debugf("%s: goserver.Conn.run: recvChan is closed", c.RemoteAddr())
					}
				}
				if p != nil {
					select {
					case <-c.context.Done():
						return
					case c.handChan <- p:
					}
				}
			}
		}
	})
}

//readPacket read a packet
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
		c.pipe(func(h Handle, next func()) {
			temp := h.ReadPacket(c, next)
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
		case result <- p:
		case <-c.context.Done():
			return
		}
	})
	return result
}

//recv create a receive only packet channel
func (c *Conn) recv(maxRecvChanCount int, heartBeat <-chan struct{}) <-chan Packet {
	result := make(chan Packet, maxRecvChanCount)
	go c.safeFn(func() {
		defer func() {
			close(result)
			if c.isDebug && c.option.Logger != nil {
				c.option.Logger.Debugf("%s: goserver.Conn.recv: recvChan is closed", c.RemoteAddr())
				c.option.Logger.Debugf("%s: goserver.Conn.recv: recv goruntinue exit", c.RemoteAddr())
			}
		}()
		for c.rwc != nil {
			pchan := c.readPacket()
			select {
			case <-c.context.Done():
				return
			case p, ok := <-pchan:
				if !ok {
					c.option.Logger.Debugf("%s: goserver.Conn.recv: packet channel prematurely closed", c.RemoteAddr())
				}
				select {
				case <-c.context.Done():
					return
				case result <- p:
					c.state.RecvPacketCount++
					if c.isDebug && c.option.Logger != nil {
						c.option.Logger.Debugf("%s: goserver.Conn.recv: read a packet", c.RemoteAddr())
					}
					select {
					case <-heartBeat:
					default:
					}
				}
			}
		}
	})
	return result
}

//send create a send only packet channel
func (c *Conn) send(maxSendChanCount int, heartBeat <-chan struct{}) chan<- Packet {
	result := make(chan Packet, maxSendChanCount)
	go c.safeFn(func() {
		defer func() {
			if c.isDebug && c.option.Logger != nil {
				c.option.Logger.Debugf("%s: goserver.Conn.send: send goruntinue exit", c.RemoteAddr())
			}
		}()
		for c.rwc != nil {
			select {
			case <-c.context.Done():
				return
			case packet, ok := <-result:
				c.state.SendPacketCount++
				if !ok {
					if c.isDebug && c.option.Logger != nil {
						c.option.Logger.Debugf("%s: goserver.Conn.send: send packet chan was closed", c.RemoteAddr())
					}
					return
				}
				if packet == nil {
					if c.isDebug && c.option.Logger != nil {
						c.option.Logger.Debugf("%s: goserver.Conn.send: the send packet is nil", c.RemoteAddr())
					}
					break
				}
				sendData, err := packet.Serialize()
				if err != nil {
					if c.option.Logger != nil {
						c.option.Logger.Error(err)
					}
				}
				c.rwc.SetWriteDeadline(time.Now().Add(c.option.WriteDataTimeOut))
				_, err = c.rwc.Write(sendData)
				if err != nil {
					c.pipe(func(h Handle, next func()) { h.OnSendError(c, packet, err, next) })
				} else {
					if c.isDebug && c.option.Logger != nil {
						c.option.Logger.Debugf("%s: goserver.Conn.send: send a packet", c.RemoteAddr())
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

//message create a send only hand packet channel
func (c *Conn) message(maxHandNum int, heartBeat <-chan struct{}) chan<- Packet {
	result := make(chan Packet, maxHandNum)
	go c.safeFn(func() {
		defer func() {
			if c.isDebug && c.option.Logger != nil {
				c.option.Logger.Debugf("%s: goserver.Conn.message: hand goruntinue exit", c.RemoteAddr())
			}
		}()
		for {
			select {
			case <-c.context.Done():
				return
			case p, ok := <-result:
				if !ok {
					if c.isDebug && c.option.Logger != nil {
						c.option.Logger.Debugf("%s: goserver.Conn.message: hand packet chan was closed", c.RemoteAddr())
					}
					return
				}
				c.pipe(func(h Handle, next func()) { h.OnMessage(c, p, next) })
				if c.isDebug && c.option.Logger != nil {
					c.option.Logger.Debugf("%s: goserver.Conn.message: hand a packet", c.RemoteAddr())
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

//heartBeat create a receive only heartBeat checking chan
func (c *Conn) heartBeat(timeOut time.Duration, callback func()) <-chan struct{} {
	result := make(chan struct{})
	go func() {
		defer func() {
			close(result)
			if c.isDebug && c.option.Logger != nil {
				c.option.Logger.Debugf("%s: goserver.Conn.heartBeat: heartBeat goruntinue exit", c.RemoteAddr())
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
