package goserver

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
	ctx     context.Context //global context
	cancel  func()          //global context cancel function
	isDebug bool            //is open inner debug message flag
	pipe    Pipe            //connection handle pipeline
}

//NewConn return a wrap of raw conn
func NewConn(ctx context.Context, rwc net.Conn, option ConnOption) Conn {
	c := Conn{
		rwc:    rwc,
		option: option,
		state: &ConnState{
			ActiveTime: time.Now(),
			RemoteAddr: rwc.RemoteAddr().String(),
		},
	}
	c.valid()
	c.ctx, c.cancel = context.WithCancel(ctx)
	return c
}

func (c Conn) valid() {
	if c.option.MaxWaitCountByHandTimeOut <= 0 {
		panic("goserver.Conn.valid: option.MaxWaitCountByHandTimeOut不允许设置为0,这会导致无法处理数据包")
	}
}

//UseDebug open inner debug log
func (c *Conn) UseDebug() {
	if c.isDebug = c.option.Logger != nil; !c.isDebug {
		fmt.Println("goserver.Conn.UseDebug: c.option.Logger is nil")
	}
}

//UsePipe create registrable pipeline and return
//注意这里的类型是指针，如果取值类型，由于值拷贝，导致外部的连接对象无法读取存进来的pipe，因为他被赋值给拷贝出来的新Conn实例了
//这里引起的思考:如果结构体方法不需要结构体指针就取 值类型方法，否则就选指针方法
//当创建出来的实例是值类型时，如果实例的字段是设计为只读的，那么最好定义值类型方法，如果需要修改字段值，那么就选择定义为指针方法
//值类型由于存在于内存栈中不会被GC扫描，可以提高程序性能
func (c *Conn) UsePipe(pipe ...Pipe) Pipe {
	if len(pipe) != 0 {
		c.pipe = pipe[0]
	}
	if c.pipe == nil {
		c.pipe = NewPipe(c.ctx)
	}
	return c.pipe
}

//RemoteAddr get remote client's ip address
func (c Conn) RemoteAddr() string {
	return c.rwc.RemoteAddr().String()
}

//LocalAddr get host ip address
func (c Conn) LocalAddr() string {
	return c.rwc.LocalAddr().String()
}

//Raw get row connection
func (c Conn) Raw() net.Conn {
	return c.rwc
}

//Run start run server and receive and handle and send packet
func (c Conn) Run() {
	go c.safeFn(func() {
		c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnConnection(ctx, c, next) })
		c.recv(1)
	})
}

//Read read a data frame from connection
func (c Conn) Read(b []byte) (n int, err error) {
	c.rwc.SetReadDeadline(time.Now().Add(c.option.RecvTimeOut))
	n, err = c.rwc.Read(b)
	if err != nil {
		c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnRecvError(ctx, c, err, next) })
	}
	return
}

//Write send a packet to remote connection
func (c Conn) Write(packet Packet) (err error) {
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
func (c Conn) Close(msg ...string) {
	defer func() {
		select {
		case <-c.ctx.Done():
			c.rwc.SetDeadline(time.Time{}) //set deadline timeout 设置客户端链接超时，是至关重要的。否则，一个超慢或已消失的客户端，可能会泄漏文件描述符，并最终导致异常
			c.rwc.Close()
			c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnClose(ctx, c.state, next) })
		}
	}()
	c.cancel()
	c.state.Message = strings.Join(msg, ",")
	c.state.ComplateTime = time.Now()

	// runtime.GC()         //强制GC      待定可能有问题
	// debug.FreeOSMemory() //强制释放内存 待定可能有问题
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

//safeFn proxy agent,used to safe invoke and recover panic
func (c Conn) safeFn(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			defer recover()
			c.pipe.schedule(func(h Handle, ctx context.Context, next func(context.Context)) { h.OnPanic(ctx, c, err.(error), next) })
			if c.option.Logger != nil {
				c.option.Logger.Errorf("goserver.Conn.safeFn: %s", err)
				c.option.Logger.Errorf("goserver.Conn.safeFn: %s", string(debug.Stack()))
			}
		}
	}()
	fn()
}
