package goserver

import (
	"context"
	"net"
)

//Dial .
func Dial(network, address string, option ConnOption) (*Conn, error) {
	rwc, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	conn := NewConn(context.Background(), rwc, option)
	return &conn, nil
}

//DialTCP create tcp conn
func DialTCP(address string, option ConnOption) (*Conn, error) {
	return Dial("tcp", address, option)
}

//DialUDP create tcp conn
func DialUDP(address string, option ConnOption) (*Conn, error) {
	return Dial("udp", address, option)
}
