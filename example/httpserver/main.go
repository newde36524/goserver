package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	srv "github.com/newde36524/goserver"
	"github.com/newde36524/goserver/example/httpserver/handles"

	"github.com/issue9/logs"
)

func init() {
	err := logs.InitFromXMLFile("./logs.xml")
	if err != nil {
		fmt.Println(err)
		<-time.After(10 * time.Second)
		return
	}
	go func() {
		for {
			fmt.Printf("当前协程数:%d\n", runtime.NumGoroutine())
			time.Sleep(time.Second)
		}
	}()
}

func main() {
	address := "0.0.0.0:12336"
	logger, err := srv.NewDefaultLogger()
	opt := srv.ConnOption{
		SendTimeOut: 1 * time.Minute, //发送消息包超时时间
		RecvTimeOut: 1 * time.Minute, //接收消息包超时时间
		Logger:      logger,          //日志打印对象
	}
	server, err := srv.TCPServer(opt)
	if err != nil {
		logs.Error(err)
	}
	server.Use(handles.RootHandle{})
	server.UseDebug()
	server.Binding(address)
	logs.Infof("服务器开始监听...  监听地址:%s", address)
	fmt.Scanln()
	<-context.Background().Done()
}
