package customer

import (
	"runtime"
	"time"

	"github.com/issue9/logs"

	"github.com/newde36524/goserver"
)

//LogHandle tcpserver使用示例,打印相关日志
type LogHandle struct {
	goserver.BaseHandle
}

//OnMessage .
func (*LogHandle) OnMessage(ctx goserver.MessageContext) {
	logs.Infof("日志模块输出:  开始计时")
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		sub := endTime.Sub(startTime).Seconds() * 1000
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		logs.Infof("日志模块输出:  开始时间:%s  结束时间:%s 总耗时: %f ms,当前内存资源:%d KB", startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"), sub, m.Alloc/1024)
	}()
	ctx.Next()
}
