package goserver

import (
	"context"
	"time"
)

//goPool 普通协程池
type goPool struct {
	work    chan func()
	sem     chan struct{}
	timeout time.Duration
	taskMap map[interface{}]chan func()
}

//newGoPool .
func newGoPool(size int, toExit time.Duration) *goPool {
	return &goPool{
		work:    make(chan func()),
		sem:     make(chan struct{}, size),
		timeout: toExit,
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

//gPool .
//针对key值进行并行调用的协程池,同一个key下的任务串行,不同key下的任务并行
type gPool struct {
	ctx     context.Context
	taskNum int
	exp     time.Duration
	m       map[interface{}]*gItem
	sign    chan struct{}
}

func newgPoll(ctx context.Context, taskNum int, exp time.Duration, size int) gPool {
	g := gPool{
		ctx:     ctx,
		taskNum: taskNum,
		exp:     exp,
		m:       make(map[interface{}]*gItem),
		sign:    make(chan struct{}, size),
	}
	return g
}

func (g gPool) SchduleByKey(key interface{}, task func()) {
	if v, ok := g.m[key]; ok {
		v.DoOrInChan(task)
	} else {
		select {
		case g.sign <- struct{}{}:
		}
		g.m[key] = newgItem(g.ctx, g.taskNum, g.exp, func() {
			delete(g.m, key)
			select {
			case <-g.sign:
			default:
			}
		})
		g.m[key].DoOrInChan(task)
	}
}

type gItem struct {
	tasks  chan func()     //任务通道
	sign   chan struct{}   //是否加入任务通道信号
	ctx    context.Context //退出协程信号
	exp    time.Duration
	onExit func()
}

func newgItem(ctx context.Context, taskNum int, exp time.Duration, onExit func()) *gItem {
	return &gItem{
		tasks:  make(chan func(), taskNum),
		sign:   make(chan struct{}, 1),
		ctx:    ctx,
		exp:    exp,
		onExit: onExit,
	}
}

func (g *gItem) DoOrInChan(task func()) {
	select {
	case g.sign <- struct{}{}: //保证只会开启一个协程
		go g.worker()
	default:
	}
	select {
	case <-g.ctx.Done():
	case g.tasks <- task: //
	case g.sign <- struct{}{}:
		go g.worker()
		select {
		case <-g.ctx.Done():
		case g.tasks <- task:
		}
	}
}

func (g *gItem) worker() {
	timer := time.NewTimer(g.exp)
	defer timer.Stop()
	defer func() {
		select {
		case <-g.sign:
		default:
		}
		if g.onExit != nil {
			g.onExit()
		}
	}()
	for {
		select {
		case <-g.ctx.Done():
			return
		case task := <-g.tasks: //执行任务优先
			timer.Reset(g.exp)
			task()
		case <-timer.C:
			return
		}
	}
}
