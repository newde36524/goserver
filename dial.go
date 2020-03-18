package goserver

import (
	"context"
	"net"
)

//Dial .
func Dial(network, address string, opt ConnOption) (*Conn, error) {
	rwc, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	conn := NewConn(context.Background(), rwc, opt)
	return conn, nil
}

//DialTCP create tcp conn
func DialTCP(address string, opt ConnOption) (*Conn, error) {
	return Dial("tcp", address, opt)
}

//DialUDP create udp conn
func DialUDP(address string, opt ConnOption) (*Conn, error) {
	return Dial("udp", address, opt)
}
