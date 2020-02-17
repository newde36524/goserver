package goserver

import "time"

//用于检测连接时间，避免连接长时间占用文件描述符造成文件描述符泄露

//通过一个定时器注册组件来完成所有连接的超时工作

/*
issure:
	1. 需要能增删定时任务节点
	2. 任务节点不能阻塞
	3.



*/

type TimeWheel struct {
}

func (*TimeWheel) Regist(fd int, d time.Duration, handle interface{}) {

}

func (*TimeWheel) Remove(fd int) {

}

func (*TimeWheel) Brush(fd int, d time.Duration) {

}

func (*TimeWheel) cycleCheck() {

}
