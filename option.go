package goserver

import "time"

//ConnOption 连接配置项
type ConnOption struct {
	RecvTimeOut time.Duration //接收包消息处理超时时间
	SendTimeOut time.Duration //发送数据超时时间
	Logger      Logger        //日志打印对象
}
