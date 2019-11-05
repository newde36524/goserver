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
package goserver

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
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
	handles []Handle        //connection handle pipeline
}

//NewConn return a wrap of raw conn
func NewConn(rwc net.Conn, option ConnOption, hs []Handle) (result *Conn) {
	result = &Conn{
		rwc:     rwc,
		option:  option,
		handles: hs,
		state: &ConnState{
			ActiveTime: time.Now(),
			RemoteAddr: rwc.RemoteAddr().String(),
		},
	}
	result.context, result.cancel = context.WithCancel(context.Background())
	return
}

//UseDebug open inner debug log
func (c *Conn) UseDebug() {
	if c.isDebug = c.option.Logger != nil; !c.isDebug {
		fmt.Println("goserver.Conn.UseDebug: c.option.Logger is nil")
	}
}

//RemoteAddr get remote client's ip address
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

//Run start run server and receive and handle and send packet
func (c *Conn) Run() {
	go c.safeFn(func() {
		c.pipe(func(h Handle, next func()) { h.OnConnection(c, next) })
		c.recv(1)
	})
}

//Read read a data frame from connection
func (c *Conn) Read(b []byte) (n int, err error) {
	c.rwc.SetReadDeadline(time.Now().Add(c.option.RecvTimeOut))
	n, err = c.rwc.Read(b)
	if err != nil {
		c.pipe(func(h Handle, next func()) { h.OnRecvError(c, err, next) })
	}
	return
}

//Write send a packet to remote connection
func (c *Conn) Write(packet Packet) (err error) {
	if packet == nil {
		if c.isDebug {
			c.option.Logger.Debugf("%s: goserver.Conn.Write: packet is nil,do nothing", c.RemoteAddr())
		}
		return
	}
	sendData, err := packet.Serialize()
	if err != nil {
		return
	}
	c.rwc.SetWriteDeadline(time.Now().Add(c.option.SendTimeOut))
	_, err = c.rwc.Write(sendData)
	c.state.SendPacketCount++
	return
}

//Close close connection
func (c *Conn) Close(msg ...string) {
	defer func() {
		select {
		case <-c.context.Done():
			c.rwc.SetDeadline(time.Now().Add(time.Second)) //set deadline timeout 设置客户端链接超时，是至关重要的。否则，一个超慢或已消失的客户端，可能会泄漏文件描述符，并最终导致异常
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
	c.state.Message = strings.Join(msg, ",")
	c.state.ComplateTime = time.Now()

	// runtime.GC()         //强制GC      待定可能有问题
	// debug.FreeOSMemory() //强制释放内存 待定可能有问题
}

//readPacket read a packet
func (c *Conn) readPacket(size int) <-chan Packet {
	result := make(chan Packet, size)
	go c.safeFn(func() {
		defer close(result)
		for {
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
			case <-c.context.Done():
				return
			case result <- p:
			}
		}
	})
	return result
}

//recv create a receive only packet channel
func (c *Conn) recv(size int) {
	go c.safeFn(func() {
		defer func() {
			if c.isDebug {
				c.option.Logger.Debugf("%s: goserver.Conn.recv: recv goruntinue exit", c.RemoteAddr())
			}
		}()
		pch := c.readPacket(size)
		for c.rwc != nil {
			select {
			case <-c.context.Done():
				return
			case <-time.After(c.option.RecvTimeOut):
				c.pipe(func(h Handle, next func()) { h.OnRecvTimeOut(c, next) })
			case p := <-pch:
				c.state.RecvPacketCount++
				c.pipe(func(h Handle, next func()) { h.OnMessage(c, p, next) })
			}
		}
	})
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
