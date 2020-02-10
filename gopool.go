package goserver

import "time"

//GoPoll .
type GoPoll struct {
	work    chan func()
	sem     chan struct{}
	timeout time.Duration
}

//NewGoPoll .
func NewGoPoll(size int, forExit time.Duration) *GoPoll {
	return &GoPoll{
		work:    make(chan func()),
		sem:     make(chan struct{}, size),
		timeout: forExit,
	}
}

//Grow .
func (p *GoPoll) Grow(num int) error {
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
func (p *GoPoll) Schedule(task func()) error {
	select {
	case p.work <- task:
	case p.sem <- struct{}{}:
		go p.worker(p.timeout, task)
	}
	return nil
}

func (p *GoPoll) worker(delay time.Duration, task func()) {
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
