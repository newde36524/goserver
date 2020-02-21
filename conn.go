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
	rwc      net.Conn        //row connection
	option   ConnOption      //connection option object
	state    *ConnState      //connection state
	ctx      context.Context //global context
	cancel   func()          //global context cancel function
	isDebug  bool            //is open inner debug message flag
	pipe     Pipe            //connection handle pipeline
	readTime time.Time       //connection read event trigger time
}

//NewConn return a wrap of raw conn
func NewConn(ctx context.Context, rwc net.Conn, opt ConnOption) Conn {
	c := Conn{
		rwc:      rwc,
		option:   opt,
		readTime: time.Now(),
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
func (c *Conn) UsePipe(p ...Pipe) Pipe {
	if len(p) != 0 {
		c.pipe = p[0]
	}
	if c.pipe == nil {
		c.pipe = newPipe(c.ctx)
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

//Read read a data frame from connection
func (c Conn) Read(b []byte) (n int, err error) {
	c.rwc.SetReadDeadline(time.Now().Add(c.option.RecvTimeOut))
	return c.rwc.Read(b)
}

//Write send a packet to remote connection
func (c Conn) Write(p Packet) (err error) {
	if p == nil {
		if c.isDebug {
			c.option.Logger.Debugf("%s: goserver.Conn.Write: packet is nil,do nothing", c.RemoteAddr())
		}
		return
	}
	sendData, err := p.Serialize()
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
