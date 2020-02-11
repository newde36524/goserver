package goserver

import (
	"io"
)

//BytePool .
type ByteBuffer struct {
	io.ReadWriter
	w   int
	r   int
	buf []byte
	rd  io.ReadWriter
}

func NewByteBuffer(rd io.ReadWriter) ByteBuffer {
	return ByteBuffer{
		rd: rd,
	}
}

func (b ByteBuffer) Write(p []byte) (n int, err error) {
	n, err = b.Write(p)
	b.w += n
	return
}
func (b ByteBuffer) Read(p []byte) (n int, err error) {
	n, err = b.rd.Read(b.buf)
	if err != nil {
		return
	}
	b.r += n

	return
}
func (b ByteBuffer) SumPacket() {}
