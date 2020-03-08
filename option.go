package goserver

import "time"

//ConnOption 连接配置项
type ConnOption struct {
	RecvTimeOut               time.Duration //接收包消息处理超时时间
	SendTimeOut               time.Duration //发送数据超时时间
	HandTimeOut               time.Duration //处理消息超时时间
	MaxWaitCountByHandTimeOut int           //最大处理消息协程堆积数量 只在windows下有效
	ParallelSize              int           //多连接触发事件并行处理数量
	MaxGopollExpire           time.Duration //协程池声明周期
}

//ModOption .
type ModOption func(*ConnOption)

func initOptions(opts ...ModOption) *ConnOption {
	var (
		recvTimeOut               time.Duration
		sendTimeOut               time.Duration
		handTimeOut               time.Duration
		maxWaitCountByHandTimeOut int
		parallelSize              = 10000
		maxGopollExpire           = time.Second
	)
	opt := &ConnOption{
		RecvTimeOut:               recvTimeOut,
		SendTimeOut:               sendTimeOut,
		HandTimeOut:               handTimeOut,
		MaxWaitCountByHandTimeOut: 1,
		ParallelSize:              parallelSize,
		MaxGopollExpire:           maxGopollExpire,
	}
	for _, o := range opts {
		o(opt)
	}
	if opt.RecvTimeOut == recvTimeOut {
		panicError("option.go: recvTimeOut option not set")
	}
	if opt.SendTimeOut == sendTimeOut {
		panicError("option.go: sendTimeOut option not set")
	}
	if opt.HandTimeOut == handTimeOut {
		panicError("option.go: handTimeOut option not set")
	}
	if opt.MaxWaitCountByHandTimeOut < maxWaitCountByHandTimeOut {
		panicError("option.go: The maxWaitCountByHandTimeOut value must be greater than or equal to 1")
	}
	if opt.ParallelSize < 1 {
		panicError("option.go: The parallelSize value must be greater than or equal to 1")
	}
	return opt
}
