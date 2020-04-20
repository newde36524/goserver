package goserver

import "errors"

var (
	//errorFdNotfound .
	errorFdNotfound = errors.New("can not get fd")
	//errRecvTimeOutNotSet .
	errRecvTimeOutNotSet = errors.New("recvTimeOut option not set")
	//errSendTimeOutNotSet .
	errSendTimeOutNotSet = errors.New("sendTimeOut option not set")
	//errParallelSize .
	errParallelSize = errors.New("the parallelSize value must be greater than or equal to 1")
	//errReadPacket .
	errReadPacket = errors.New("read the first of packet from handle pipe only")
	//ErrRecvTimeOutNotSet .
	ErrRecvTimeOutNotSet = errors.New("recvTimeOut option not set")
	//ErrSendTimeOutNotSet .
	ErrSendTimeOutNotSet = errors.New("sendTimeOut option not set")
	//ErrParallelSize .
	ErrParallelSize = errors.New("the parallelSize value must be greater than or equal to 1")
)
