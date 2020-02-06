package handle

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/newde36524/goserver"
)

type traceHandle struct {
	goserver.Handle
}

//NewTraceHandle .
func NewTraceHandle() goserver.Handle {
	return &traceHandle{}
}

//OnMessage .
func (h *traceHandle) OnMessage(ctx context.Context, conn goserver.Conn, p goserver.Packet, next func(context.Context)) {
	next(ctx)
	trace()
}

//trace 跟踪调用链
func trace() {
	sb := strings.Builder{}
	sb.WriteString("========================= start =========================\n")
	index := 1
	for {
		if funcName, file, line, ok := runtime.Caller(index); ok {
			index++
			sb.WriteString(fmt.Sprintf(`
	%s  		file: %s, line: %d
					^
					|
					|
					|
	`, runtime.FuncForPC(funcName).Name(), file, line))
		} else {
			break
		}
	}
	sb.WriteString("========================== end ==========================\n")
	fmt.Println(sb.String())
	// logs.Trace(sb.String())
}
