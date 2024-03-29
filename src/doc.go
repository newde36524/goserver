/*
	package main

	import (
		"flag"
		"fmt"
		"time"

		"github.com/newde36524/goserver"
	)

	var port = flag.Int("port", 12336, "The port to listen on for tcp requests.")

	func init() {
		flag.Parse()
	}

	func main() {
		address := fmt.Sprintf("0.0.0.0:%d", *port)
		server, err := goserver.TCPServer(goserver.ModOption(func(opt *goserver.ConnOption) {
			opt.SendTimeOut = time.Minute //发送消息包超时时间
			opt.RecvTimeOut = time.Minute //接收消息包超时时间
		}))
		if err != nil {
			fmt.Println(err)
		}
		server.UsePipe().
			Regist(new(RootHandle))
		server.Binding(address)
		select{}
	}

	type RootHandle struct {
		goserver.BaseHandle
	}

	func (RootHandle) ReadPacket(ctx goserver.ReadContext) goserver.Packet {
		defer ctx.Next()
		b := make([]byte, 1024)
		n, _ := ctx.Conn().Read(b)
		p := goserver.P(b[:n])
		return &p
	}

	func (RootHandle) OnConnection(ctx goserver.ConnectionContext) {
		defer ctx.Next()
	}

	func (RootHandle) OnMessage(ctx goserver.MessageContext) {
		defer ctx.Next()
		ctx.Conn().Write(ctx.Packet())
	}

	func (RootHandle) OnClose(ctx goserver.CloseContext) {
		defer ctx.Next()
	}
*/

package goserver
