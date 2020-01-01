package customer

import (
	"context"
	"io"
	"net"

	"github.com/newde36524/goserver"

	"github.com/issue9/logs"
)

//RootHandle tcpserver使用示例,回复相同的内容
type RootHandle struct {
	goserver.BaseHandle
	//可增加新的属性
	//可增加全局属性，比如多个客户端连接可选择转发数据给其他连接，而增加一个全局map
}

//ReadPacket .
func (RootHandle) ReadPacket(ctx context.Context, conn goserver.Conn, next func(context.Context)) goserver.Packet {
	//todo 定义读取数据帧的规则
	b := make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		logs.Error(err)
		//当客户端连接强制中断时,在wsl中err被识别为io.EOF  而在linux和windows中识别为net.Error
		switch e := err.(type) {
		case net.Error:
			if !e.Timeout() {
				logs.Error(err)
				conn.Close()
				return nil
			}
		default:
			if err == io.EOF {
				conn.Close()
				return nil
			}
		}
	}
	p := &Packet{}
	p.SetBuffer(b[:n])

	return p
}

// //OnConnection .
// func (RootHandle) OnConnection(ctx context.Context, conn goserver.Conn, next func(context.Context)) {
// 	//todo 连接建立时处理,用于一些建立连接时,需要主动下发数据包的场景,可以在这里开启心跳协程,做登录验证等等
// 	logs.Infof("%s: 客户端建立连接", conn.RemoteAddr())
// }

//OnMessage .
func (RootHandle) OnMessage(ctx context.Context, conn goserver.Conn, p goserver.Packet, next func(context.Context)) {
	logs.Info(ctx.Value("logger"))
	logs.Infof("%s:我好像收到了不知名快递哦", conn.RemoteAddr())
	sendP := &Packet{}
	if p != nil {
		data := p.GetBuffer()
		sendP.SetBuffer(data)
	}
	conn.Write(sendP)
	next(ctx)
}

//OnClose .
func (RootHandle) OnClose(ctx context.Context, state *goserver.ConnState, next func(context.Context)) {
	logs.Infof("客户端断开连接,连接状态:%s", state.String())
}

//OnRecvTimeOut .
func (RootHandle) OnRecvTimeOut(ctx context.Context, conn goserver.Conn, next func(context.Context)) {
	logs.Infof("%s: 服务器接收消息超时", conn.RemoteAddr())
}

//OnHandTimeOut .
func (RootHandle) OnHandTimeOut(ctx context.Context, conn goserver.Conn, next func(context.Context)) {
	logs.Infof("%s: 服务器处理消息超时", conn.RemoteAddr())
}

//OnPanic .
func (RootHandle) OnPanic(ctx context.Context, conn goserver.Conn, err error, next func(context.Context)) {
	logs.Errorf("%s: 服务器发生恐慌,错误信息:%s", conn.RemoteAddr(), err)
}

//OnRecvError .
func (RootHandle) OnRecvError(ctx context.Context, conn goserver.Conn, err error, next func(context.Context)) {
	logs.Errorf("%s: 服务器接收数据异常,错误信息:%s", conn.RemoteAddr(), err)
}