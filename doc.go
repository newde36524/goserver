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

	//ReadPacket .
	func (RootHandle) ReadPacket(ctx context.Context, conn goserver.Conn, next func(context.Context)) goserver.Packet {
		defer next(ctx)
		b := make([]byte, 1024)
		n, _ := conn.Read(b)
		p := goserver.P(b[:n])
		return &p
	}

	func (RootHandle) OnConnection(ctx context.Context, conn goserver.Conn, next func(context.Context)) {
		defer next(ctx)
	}

	func (RootHandle) OnMessage(ctx context.Context, conn goserver.Conn, p goserver.Packet, next func(context.Context)) {
		defer next(ctx)
		conn.Write(p)
	}

	func (RootHandle) OnClose(ctx context.Context, state *goserver.ConnState, next func(context.Context)) {
		defer next(ctx)
	}
*/

package goserver
