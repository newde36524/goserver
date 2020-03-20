package goserver

import (
	"sync"
	"time"
)

//entity .
type entity struct {
	start time.Time
	delay time.Duration
	task  func(remove func())
}

type loopTask struct {
	tasks []entity
	delay time.Duration
	mu    sync.Mutex
}

func (l *loopTask) Add(delay time.Duration, task func(remove func())) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tasks = append(l.tasks, entity{
		start: time.Now(),
		delay: delay,
		task:  task,
	})
}

func (l *loopTask) Len() int {
	return len(l.tasks)
}

func (l *loopTask) Start() {
	var t *time.Timer
	t = time.AfterFunc(l.delay, func() {
		for i := 0; i < len(l.tasks); i++ {
			var (
				once     sync.Once
				entity   = l.tasks[i]
				isRemove = false
				pop      = func(i int) {
					front := l.tasks[:i]
					back := l.tasks[i+1:]
					l.tasks = append(front, back...)
				}
				remove = func() {
					once.Do(func() {
						isRemove = true
						pop(i)
						i--
					})
				}
			)
			if time.Now().Sub(entity.start) >= entity.delay {
				entity.start = time.Now().Add(entity.delay)
				entity.task(remove)
				if !isRemove {
					pop(i)
					l.tasks = append(l.tasks, entity)
				}
			} else {
				/*
					1. start为内部指定当前时间,一定是递增的,这里退出避免无效遍历
					2. 假设当前时间减去第一个任务时间为6分钟剩余,而delay为10分钟,那么只要再等待4分钟就足够了
					3. 只针对于固定的时间间隔
				*/
				t.Reset(time.Now().Sub(entity.start))
				break
			}
		}
		t.Reset(l.delay)
	})
}

//loopTaskPool 循环任务池
type loopTaskPool struct {
	pool  sync.Pool
	once  sync.Once
	mu    sync.Mutex
	loops []*loopTask
	idx   int
}

//Schdule .
func (l *loopTaskPool) Schdule(delay time.Duration, task func(remove func())) {
	l.once.Do(func() {
		l.pool.New = func() interface{} {
			v := &loopTask{
				delay: delay,
			}
			v.Start()
			l.loops = append(l.loops, v)
			return v
		}
	})
	v := l.pool.Get()
	l.pool.Put(&v)
	v = l.loops[l.idx%len(l.loops)] //012012012012012
	loopTask := v.(*loopTask)
	loopTask.Add(delay, task)
	l.mu.Lock()
	l.idx++
	l.mu.Unlock()
}
