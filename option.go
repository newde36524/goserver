package goserver

import (
	"time"
)

//ConnOption 连接配置项
type ConnOption struct {
	RecvTimeOut     time.Duration //接收包消息处理超时时间
	SendTimeOut     time.Duration //发送数据超时时间
	ParallelSize    int           //多连接触发事件并行处理数量
	MaxGopollExpire time.Duration //协程池协程生命周期
}

//ModOption .
type ModOption func(*ConnOption)

func initOptions(opts ...ModOption) *ConnOption {
	var (
		recvTimeOut     time.Duration
		sendTimeOut     time.Duration
		parallelSize    = 10000
		maxGopollExpire = time.Second
	)
	opt := &ConnOption{
		RecvTimeOut:     recvTimeOut,
		SendTimeOut:     sendTimeOut,
		ParallelSize:    parallelSize,
		MaxGopollExpire: maxGopollExpire,
	}
	for _, o := range opts {
		o(opt)
	}
	if opt.RecvTimeOut == recvTimeOut {
		panicError(ErrRecvTimeOutNotSet.Error())
	}
	if opt.SendTimeOut == sendTimeOut {
		panicError(ErrSendTimeOutNotSet.Error())
	}
	if opt.ParallelSize < 1 {
		panicError(ErrParallelSize.Error())
	}
	return opt
}
