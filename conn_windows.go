// +build windows

package goserver

import (
	"context"
	"net"
	"time"
)

//Conn net.Conn proxy object
type Conn struct {
	rwc     net.Conn        //row connection
	option  ConnOption      //connection option object
	state   *ConnState      //connection state
	context context.Context //global context
	cancel  func()          //global context cancel function
	isDebug bool            //is open inner debug message flag
	pipe    Pipe            //connection handle pipeline
}

//NewConn return a wrap of raw conn
func NewConn(ctx context.Context, rwc net.Conn, option ConnOption, hs []Handle) Conn {
	result := Conn{
		rwc:    rwc,
		option: option,
		state: &ConnState{
			ActiveTime: time.Now(),
			RemoteAddr: rwc.RemoteAddr().String(),
		},
	}
	result.valid()
	result.context, result.cancel = context.WithCancel(ctx)
	result.pipe = NewPipe(result.context, hs)
	return result
}
