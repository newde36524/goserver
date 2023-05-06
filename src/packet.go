package goserver

//Packet 协议包内容
type Packet interface {
	SetBuffer(frame []byte)     // 设置客户端上传的数据帧
	GetBuffer() []byte          // 获取客户端上传的数据帧
	Serialize() ([]byte, error) // 获取服务端解析后的数据帧
}

//P .
type P []byte

//SetBuffer .
func (p *P) SetBuffer(frame []byte) {
	*p = frame
}

//GetBuffer .
func (p P) GetBuffer() []byte {
	return p
}

//Serialize .
func (p P) Serialize() ([]byte, error) {
	return p, nil
}
