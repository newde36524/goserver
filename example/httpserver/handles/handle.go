package handles

import (
	"bufio"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	srv "github.com/newde36524/goserver"

	"github.com/issue9/logs"
)

//RootHandle tcpserver使用示例,回复相同的内容
type RootHandle struct {
	srv.Handle
}

//ReadPacket .
func (RootHandle) ReadPacket(conn *srv.Conn, next func()) srv.Packet {
	return nil
}

//OnConnection .
func (RootHandle) OnConnection(conn *srv.Conn, next func()) {
	//todo 连接建立时处理,用于一些建立连接时,需要主动下发数据包的场景,可以在这里开启心跳协程,做登录验证等等
	logs.Infof("%s: 对方好像对你很感兴趣呦", conn.RemoteAddr())
	if req, err := http.ReadRequest(bufio.NewReader(conn.Raw())); err == nil {
		r := Request{
			Method:           req.Method,
			URL:              req.URL,
			Proto:            req.Proto,
			ProtoMajor:       req.ProtoMajor,
			ProtoMinor:       req.ProtoMinor,
			Header:           req.Header,
			Body:             req.Body,
			ContentLength:    req.ContentLength,
			TransferEncoding: req.TransferEncoding,
			Close:            req.Close,
			Host:             req.Host,
			Form:             req.Form,
			PostForm:         req.PostForm,
			MultipartForm:    req.MultipartForm,
			Trailer:          req.Trailer,
			RemoteAddr:       req.RemoteAddr,
			RequestURI:       req.RequestURI,
		}
		if v, err := json.Marshal(r); err == nil {
			logs.Infof("%s", string(v))
		} else {
			logs.Error(err)
		}
		req.ParseForm()
		r = Request{
			Method:           req.Method,
			URL:              req.URL,
			Proto:            req.Proto,
			ProtoMajor:       req.ProtoMajor,
			ProtoMinor:       req.ProtoMinor,
			Header:           req.Header,
			Body:             req.Body,
			ContentLength:    req.ContentLength,
			TransferEncoding: req.TransferEncoding,
			Close:            req.Close,
			Host:             req.Host,
			Form:             req.Form,
			PostForm:         req.PostForm,
			MultipartForm:    req.MultipartForm,
			Trailer:          req.Trailer,
			RemoteAddr:       req.RemoteAddr,
			RequestURI:       req.RequestURI,
		}
		if v, err := json.Marshal(r); err == nil {
			logs.Infof("%s", string(v))
		} else {
			logs.Error(err)
		}
		logs.Infof("%#v", req)
	} else {
		logs.Error(err)
	}
	p := &srv.P{}
	p.SetBuffer([]byte("嘤嘤嘤"))
	conn.Write(p)
	conn.Close()
}

//OnMessage .
func (RootHandle) OnMessage(conn *srv.Conn, p srv.Packet, next func()) {
	next()
}

//OnClose .
func (RootHandle) OnClose(state *srv.ConnState, next func()) {
	next()
}

//OnRecvTimeOut .
func (RootHandle) OnRecvTimeOut(conn *srv.Conn, next func()) {
	next()
}

//OnHandTimeOut .
func (RootHandle) OnHandTimeOut(conn *srv.Conn, next func()) {
	next()
}

//OnPanic .
func (RootHandle) OnPanic(conn *srv.Conn, err error, next func()) {
	next()
}

//OnRecvError .
func (RootHandle) OnRecvError(conn *srv.Conn, err error, next func()) {
	next()
}

type Request struct {
	Method           string
	URL              *url.URL
	Proto            string // "HTTP/1.0"
	ProtoMajor       int    // 1
	ProtoMinor       int    // 0
	Header           http.Header
	Body             io.ReadCloser
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Host             string
	Form             url.Values
	PostForm         url.Values
	MultipartForm    *multipart.Form
	Trailer          http.Header
	RemoteAddr       string
	RequestURI       string
}
