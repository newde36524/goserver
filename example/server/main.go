package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/newde36524/goserver"
	customer "github.com/newde36524/goserver/example/Server/customer"

	"github.com/issue9/logs"
)

var port = flag.Int("p", 12336, "The port to listen on for tcp requests.")

func init() {
	flag.Parse()
	err := logs.InitFromXMLFile("./logs.xml")
	if err != nil {
		fmt.Println(err)
		<-time.After(10 * time.Second)
		return
	}
	go func() {
		buf := bufio.NewReader(os.Stdin)
		for {
			buf.ReadLine()
			fmt.Printf("当前协程数:%d\n", runtime.NumGoroutine())
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			fmt.Printf("当前内存资源:%d KB\n", m.Alloc/1024)
		}
	}()
	fmt.Println(runtime.GOOS)
	// var rLimit syscall.Rlimit
	// rLimit.Cur = 200000
	// rLimit.Max = 200000
	// if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
	// 	logs.Error(err)
	// }
}

func main() {
	address := fmt.Sprintf("0.0.0.0:%d", *port)
	server, err := goserver.TCPServer(goserver.ModOption(func(opt *goserver.ConnOption) {
		opt.SendTimeOut = time.Minute     //发送消息包超时时间
		opt.RecvTimeOut = 5 * time.Second //接收消息包超时时间
		opt.HandTimeOut = time.Minute     //处理消息包超时时间
		// opt.ParallelSize = 10
		// opt.MaxGopollExpire = 3 * time.Second
	}))
	if err != nil {
		logs.Error(err)
	}
	server.UsePipe().
		// Regist(new(customer.LogHandle)).
		Regist(new(customer.RootHandle))
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
