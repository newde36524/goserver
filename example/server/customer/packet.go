package customer

import (
	srv "github.com/newde36524/goserver"
)

//Packet .
type Packet struct {
	srv.Packet
	data []byte
}

//SetBuffer .
func (p *Packet) SetBuffer(frame []byte) {
	//todo 解析数据包，并可根据需要在结构中定义多个字段存储
	p.data = frame
}

//GetBuffer .
func (p *Packet) GetBuffer() []byte {
	//todo 解析数据包，并可根据需要在结构中定义多个字段存储
	return p.data
}

//Serialize .
func (p *Packet) Serialize() ([]byte, error) {
	//todo 数据帧的业务处理逻辑
	return p.data, nil
}
