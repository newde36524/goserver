package goserver

import (
	"fmt"
	"time"
)

//ConnState .
type ConnState struct {
	ActiveTime      time.Time   //开始活动时间
	ComplateTime    time.Time   //结束活动时间
	InnerErr        error       //异常信息
	Message         interface{} //通知信息
	RemoteAddr      string      //客户端地址
	SendPacketCount int64       //发送包的数量
	RecvPacketCount int64       //接收包的数量
}

//String 格式化输出结构体信息
func (state *ConnState) String() string {
	return fmt.Sprintf(`
*客户端IP:%s
*开始活动时间:%s
*结束活动时间:%s
*发送包的数量:%d
*接收包的数量:%d
*异常信息:%s
*通知信息:%s`,
		state.RemoteAddr,
		state.ActiveTime.Format("2006-01-02 15:04:05"),
		state.ComplateTime.Format("2006-01-02 15:04:05"),
		state.SendPacketCount,
		state.RecvPacketCount,
		state.InnerErr,
		state.Message,
	)
}
