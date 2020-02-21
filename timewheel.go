package goserver

import (
	"fmt"
	"time"
)

//用于检测连接时间，避免连接长时间占用文件描述符造成文件描述符泄露

//通过一个定时器注册组件来完成所有连接的超时工作

/*
issure:
	1. 需要能增删定时任务节点
	2. 任务节点不能阻塞
	3.
*/

type TimeWheel struct {
	data []map[time.Time]interface{}
}

func (t *TimeWheel) Regist(fd int, d time.Time, handle interface{}) {
	t.data = append(t.data, map[time.Time]interface{}{
		d: handle,
	})
}

func (*TimeWheel) Remove(fd int) {

}

func (*TimeWheel) Brush(fd int, d time.Duration) {

}

func (t *TimeWheel) cycleCheck() {
	for {
		for _, v := range t.data {
			for k, h := range v {
				if time.Now().Sub(k) > 0 {
					//todo
					fmt.Println(h)
				}
			}
		}
	}
}
