package customer

import (
	"context"
	"time"

	"github.com/issue9/logs"

	"github.com/newde36524/goserver"
)

//LogHandle tcpserver使用示例,打印相关日志
type LogHandle struct {
	goserver.BaseHandle
}

//OnMessage .
func (LogHandle) OnMessage(ctx context.Context, conn *goserver.Conn, p goserver.Packet, next func(context.Context)) {
	logs.Infof("日志模块输出:  开始计时")
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		sub := endTime.Sub(startTime).Seconds() * 1000
		logs.Infof("日志模块输出:  开始时间:%s  结束时间:%s 总耗时: %f ms", startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"), sub)
	}()
	next(context.WithValue(ctx, "logger", "日志模块正在记录日志哟"))
}

//OnClose .
func (LogHandle) OnClose(ctx context.Context, state *goserver.ConnState, next func(context.Context)) {
	logs.Info(state)
	state.Message = "日志模块修改当前信息"
	next(ctx)
}
