package goserver

import "errors"

var (
	//errorFdNotfound .
	errorFdNotfound = errors.New("can not get fd")
	//errRecvTimeOutNotSet .
	errRecvTimeOutNotSet = errors.New("recvTimeOut option not set")
	//errSendTimeOutNotSet .
	errSendTimeOutNotSet = errors.New("sendTimeOut option not set")
	//errHandTimeOutOutNotSet .
	errHandTimeOutOutNotSet = errors.New("handTimeOut option not set")
	//errMaxWaitCountByHandTimeOut .
	errMaxWaitCountByHandTimeOut = errors.New("the maxWaitCountByHandTimeOut value must be greater than or equal to 1")
	//errParallelSize .
	errParallelSize = errors.New("the parallelSize value must be greater than or equal to 1")
	//errReadPacket .
	errReadPacket = errors.New("read the first of packet from handle pipe only")
)
