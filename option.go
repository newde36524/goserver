package goserver

import (
	"errors"
	"time"
)

var (
	//ErrRecvTimeOutNotSet .
	ErrRecvTimeOutNotSet = errors.New("recvTimeOut option not set")
	//ErrSendTimeOutNotSet .
	ErrSendTimeOutNotSet = errors.New("sendTimeOut option not set")
	//ErrHandTimeOutOutNotSet .
	ErrHandTimeOutOutNotSet = errors.New("handTimeOut option not set")
	//ErrMaxWaitCountByHandTimeOut .
	ErrMaxWaitCountByHandTimeOut = errors.New("the maxWaitCountByHandTimeOut value must be greater than or equal to 1")
	//ErrParallelSize .
	ErrParallelSize = errors.New("the parallelSize value must be greater than or equal to 1")
)

//ConnOption 连接配置项
type ConnOption struct {
	RecvTimeOut               time.Duration //接收包消息处理超时时间
	SendTimeOut               time.Duration //发送数据超时时间
	HandTimeOut               time.Duration //处理消息超时时间
	MaxWaitCountByHandTimeOut int           //最大处理消息协程堆积数量 只在windows下有效
	ParallelSize              int           //多连接触发事件并行处理数量
	MaxGopollExpire           time.Duration //协程池协程生命周期
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
		panicError(ErrRecvTimeOutNotSet.Error())
	}
	if opt.SendTimeOut == sendTimeOut {
		panicError(ErrSendTimeOutNotSet.Error())
	}
	if opt.HandTimeOut == handTimeOut {
		panicError(ErrHandTimeOutOutNotSet.Error())
	}
	if opt.MaxWaitCountByHandTimeOut < maxWaitCountByHandTimeOut {
		panicError(ErrMaxWaitCountByHandTimeOut.Error())
	}
	if opt.ParallelSize < 1 {
		panicError(ErrParallelSize.Error())
	}
	return opt
}
