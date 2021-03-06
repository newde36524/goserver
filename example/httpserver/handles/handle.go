package handles

import (
	"bufio"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/newde36524/goserver"

	"github.com/issue9/logs"
)

//RootHandle .
type RootHandle struct {
	goserver.BaseHandle
}

//ReadPacket .
func (RootHandle) ReadPacket(ctx goserver.ReadContext) goserver.Packet {
	if req, err := http.ReadRequest(bufio.NewReader(ctx.Conn().Raw())); err == nil {
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
		// res, err := http.ReadResponse(bytes.NewReader([]byte("嘤嘤嘤")), req)
		// if err != nil {
		// 	logs.Error(err)
		// }
		// // res.
		// logs.Infof("%#v", req)
	} else {
		logs.Error(err)
	}
	resp := appendResp(nil, "200 OK", "", "o huo~~~~~")
	p := &goserver.P{}
	p.SetBuffer(resp)
	ctx.Conn().Write(p)
	ctx.Conn().Close()
	return nil
}

func appendResp(b []byte, status, head, body string) []byte {
	b = append(b, "HTTP/1.1"...)
	b = append(b, ' ')
	b = append(b, status...)
	b = append(b, '\r', '\n')
	b = append(b, "Server: HttpServer\r\n"...)
	b = append(b, "Date: "...)
	b = time.Now().AppendFormat(b, "Mon, 02 Jan 2006 15:04:05 GMT")
	b = append(b, '\r', '\n')
	if len(body) > 0 {
		b = append(b, "Content-Length: "...)
		b = strconv.AppendInt(b, int64(len(body)), 10)
		b = append(b, '\r', '\n')
	}
	b = append(b, head...)
	b = append(b, '\r', '\n')
	if len(body) > 0 {
		b = append(b, body...)
	}
	return b
}

//Request .
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
