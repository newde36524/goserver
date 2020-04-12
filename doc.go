/*
	package main

	import (
		"context"
		"fmt"
		"time"

		"github.com/newde36524/goserver"
	)

	func main() {
		address := "0.0.0.0:12336"
		server, err := goserver.TCPServer(goserver.ModOption(func(opt *goserver.ConnOption) {
			opt.SendTimeOut = time.Minute //发送消息包超时时间
			opt.RecvTimeOut = time.Minute //接收消息包超时时间
			opt.HandTimeOut = time.Minute //处理消息包超时时间
		}))
		if err != nil {
			fmt.Println(err)
		}
		server.UsePipe().
			Regist(new(RootHandle))
		server.Binding(address)
	}

	type RootHandle struct {
		goserver.BaseHandle
	}

	func (RootHandle) ReadPacket(ctx goserver.ReadContext) goserver.Packet {
		defer ctx.Next()
		b := make([]byte, 1024)
		n, _ := conn.Read(b)
		p := goserver.P(b[:n])
		return &p
	}

	func (RootHandle) OnConnection(ctx goserver.ConnectionContext) {
		defer ctx.Next()
	}

	func (RootHandle) OnMessage(ctx goserver.Context) {
		defer ctx.Next()
		conn.Write(p)
	}

	func (RootHandle) OnClose(ctx goserver.CloseContext) {
		defer ctx.Next()
	}
*/

package goserver
