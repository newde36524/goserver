package handles

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/newde36524/goserver"

	"github.com/issue9/logs"
)

//RootHandle tcpserver使用示例,回复相同的内容
type RootHandle struct {
	goserver.BaseHandle
}

//ReadPacket .
func (RootHandle) ReadPacket(ctx context.Context, conn goserver.Conn, next func(context.Context)) goserver.Packet {
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
		res, err := http.ReadResponse(bytes.NewReader([]byte("嘤嘤嘤")), req)
		if err != nil {
			logs.Error(err)
		}
		// res.
		logs.Infof("%#v", req)
	} else {
		logs.Error(err)
	}
	p := &goserver.P{}
	p.SetBuffer()
	conn.Write(p)
	conn.Close()
	return nil
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
