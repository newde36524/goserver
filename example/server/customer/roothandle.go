package customer

import (
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
func (*RootHandle) ReadPacket(ctx goserver.ReadContext) goserver.Packet {
	//todo 定义读取数据帧的规则
	defer ctx.Next()
	b := make([]byte, 1024)
	n, err := ctx.Conn().Read(b)
	if err != nil {
		logs.Errorf("%#v", err)
		//当客户端连接强制中断时,在wsl中err被识别为io.EOF  而在linux和windows中识别为net.Error
		switch e := err.(type) {
		case net.Error:
			if !e.Timeout() {
				logs.Error(err)
				ctx.Conn().Close()
				return nil
			}
		default:
			if err == io.EOF {
				ctx.Conn().Close()
				return nil
			}
		}
	}
	logs.Info("ReadPacket")
	p := goserver.P(b[:n])
	return &p
}

//OnConnection .
func (*RootHandle) OnConnection(ctx goserver.ConnectionContext) {
	//todo 连接建立时处理,用于一些建立连接时,需要主动下发数据包的场景,可以在这里开启心跳协程,做登录验证等等
	defer ctx.Next()
	logs.Infof("%s: 客户端建立连接", ctx.Conn().RemoteAddr())
}

//OnMessage .
func (*RootHandle) OnMessage(ctx goserver.MessageContext) {
	defer ctx.Next()
	logs.Infof("%s:获取客户端信息: %s", ctx.Conn().RemoteAddr(), string(ctx.Packet().GetBuffer()))
	ctx.Conn().Write(ctx.Packet())
}

//OnClose .
func (*RootHandle) OnClose(ctx goserver.CloseContext) {
	defer ctx.Next()
	logs.Infof("客户端断开连接,连接状态:%s", ctx.State().String())
}

//OnRecvTimeOut .
func (*RootHandle) OnRecvTimeOut(ctx goserver.RecvTimeOutContext) {
	defer ctx.Next()
	logs.Infof("%s: 服务器接收消息超时", ctx.Conn().RemoteAddr())
	ctx.Conn().Close()
}

//OnPanic .
func (*RootHandle) OnPanic(ctx goserver.PanicContext) {
	defer ctx.Next()
	logs.Errorf("%s: 服务器发生恐慌,错误信息:%s", ctx.Conn().RemoteAddr(), ctx.Error())
}
