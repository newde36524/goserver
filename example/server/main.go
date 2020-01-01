package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/newde36524/goserver"
	customer "github.com/newde36524/goserver/example/Server/customer"

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
	server, err := goserver.TCPServer(goserver.ModOption(func(opt *goserver.ConnOption) {
		logger, err := goserver.NewDefaultLogger()
		if err != nil {
			fmt.Println(err)
		}
		opt.SendTimeOut = time.Minute //发送消息包超时时间
		opt.RecvTimeOut = time.Minute //接收消息包超时时间
		opt.HandTimeOut = time.Minute //处理消息包超时时间
		opt.Logger = logger           //日志打印对象
	}))
	if err != nil {
		logs.Error(err)
	}
	server.Use(new(customer.LogHandle))
	server.Use(new(customer.RootHandle))
	server.UseDebug()
	server.Binding(address)
	logs.Infof("服务器开始监听...  监听地址:%s", address)
	fmt.Scanln()

	// signalCh := make(chan os.Signal)
	// signal.Notify(signalCh, os.Interrupt)
	// go func() {
	// 	for {
	// 		select {
	// 		case sign := <-signalCh:
	// 			fmt.Println("接收到消息:", sign)
	// 			sign.Signal()
	// 		}
	// 	}
	// }()

	<-context.Background().Done()
}
