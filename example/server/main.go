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
	// go func() {
	// 	// buf := bufio.NewReader(os.Stdin)
	// 	for {
	// 		fmt.Printf("当前协程数:%d\n", runtime.NumGoroutine())
	// 		time.Sleep(time.Second)
	// 		// buf.ReadLine()
	// 	}
	// }()
	fmt.Println(runtime.GOOS)
	// var rLimit syscall.Rlimit
	// rLimit.Cur = 200000
	// rLimit.Max = 200000
	// if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
	// 	logs.Error(err)
	// }
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
		// opt.ParallelSize = 10
		// opt.MaxGopollExpire = 3 * time.Second
	}))
	if err != nil {
		logs.Error(err)
	}
	server.UsePipe().
		// Regist(new(customer.LogHandle)).
		Regist(new(customer.RootHandle))
	// Regist(handle.NewTraceHandle())
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
