package goserver

import (
	"github.com/issue9/logs"
)

//Logger 日志接口
type Logger interface {
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Trace(v ...interface{})
	Tracef(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
}

//DefaultLogger .
type DefaultLogger struct {
	Logger
}

//NewDefaultLogger .
func NewDefaultLogger() (result DefaultLogger, err error) {
	result = DefaultLogger{}
	return
}

//Info .
func (DefaultLogger) Info(v ...interface{}) {
	logs.Info(v...)
}

//Infof .
func (DefaultLogger) Infof(format string, v ...interface{}) {
	logs.Infof(format, v...)
}

//Debug .
func (DefaultLogger) Debug(v ...interface{}) {
	logs.Debug(v...)
}

//Debugf .
func (DefaultLogger) Debugf(format string, v ...interface{}) {
	logs.Debugf(format, v...)
}

//Trace .
func (DefaultLogger) Trace(v ...interface{}) {
	logs.Trace(v...)
}

//Tracef .
func (DefaultLogger) Tracef(format string, v ...interface{}) {
	logs.Tracef(format, v...)
}

//Warn .
func (DefaultLogger) Warn(v ...interface{}) {
	logs.Warn(v...)
}

//Warnf .
func (DefaultLogger) Warnf(format string, v ...interface{}) {
	logs.Warnf(format, v...)
}

//Error .
func (DefaultLogger) Error(v ...interface{}) {
	logs.Error(v...)
}

//Errorf .
func (DefaultLogger) Errorf(format string, v ...interface{}) {
	logs.Errorf(format, v...)
}
