package goserver

import (
	"time"
)

//goPool .
type goPool struct {
	work    chan func()
	sem     chan struct{}
	timeout time.Duration
	taskMap map[interface{}]chan func()
}

//newGoPool .
func newGoPool(size int, forExit time.Duration) *goPool {
	return &goPool{
		work:    make(chan func()),
		sem:     make(chan struct{}, size),
		timeout: forExit,
		taskMap: make(map[interface{}]chan func(), 1024),
	}
}

//Grow .
func (p *goPool) Grow(num int) error {
	newSem := make(chan struct{}, num)
loop:
	for {
		select {
		case sign := <-p.sem:
			select {
			case newSem <- sign:
			default:
			}
		default:
			break loop
		}
	}
	p.sem = newSem
	return nil
}

//Schedule 把方法加入协程池并被执行
func (p *goPool) Schedule(task func()) error {
	select {
	case p.work <- task:
	case p.sem <- struct{}{}:
		go p.worker(p.timeout, task)
	}
	return nil
}

// //实现一个标识，相同标识下的任务串行，不同标识下的任务并行

// //ScheduleFlag 把方法加入协程池并被执行
// //实现一个功能，相同标识下的任务串行，不同标识下的任务并行
// //在并发中识别任务并串行
// func (p *goPool) ScheduleFlag(flag interface{}, task func()) func() {
// 	if _, ok := p.taskMap[flag]; !ok {
// 		p.taskMap[flag] = make(chan func(), 1)
// 	}
// 	p.taskMap[flag] <- task

// 	select {
// 	case p.work <- func() {
// 		t := <-p.taskMap[flag]
// 		t()
// 	}:
// 	case p.sem <- struct{}{}:
// 		go p.worker(p.timeout, func() {
// 			t := <-p.taskMap[flag]
// 			t()
// 		})
// 	}
// 	return nil
// }

func (p *goPool) worker(delay time.Duration, task func()) {
	defer func() { <-p.sem }()
	timer := time.NewTimer(delay)
	for {
		task()
		timer.Reset(delay)
		select {
		case task = <-p.work:
		case <-timer.C:
			return
		}
	}
}
