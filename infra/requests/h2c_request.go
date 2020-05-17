package requests

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/http2"
)

var (
	DefaultH2CClient *http.Client
	h2cclientGroup   Group
)

func init() {
	NewH2cClient(10 * time.Second)
}

// Newh2cClient .
func NewH2cClient(rwTimeout time.Duration, connectTimeout ...time.Duration) {
	tran := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			t := 2 * time.Second
			if len(connectTimeout) > 0 {
				t = connectTimeout[0]
			}
			fun := timeoutDialer(t)
			return fun(network, addr)
		},
	}

	DefaultH2CClient = &http.Client{
		Transport: tran,
		Timeout:   rwTimeout,
	}
}

// timeoutDialer returns functions of connection dialer with timeout settings for http.Transport Dial field.
func timeoutDialer(cTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		return conn, err
	}
}

func NewH2CRequest(rawurl string) Request {
	result := new(H2CRequest)
	req := &http.Request{
		Header: make(http.Header),
	}
	result.resq = req
	result.params = make(url.Values)
	result.url = rawurl
	result.stop = false
	return result
}

// H2CRequest .
type H2CRequest struct {
	resq            *http.Request
	resp            *http.Response
	reqe            error
	params          url.Values
	url             string
	ctx             context.Context
	singleflightKey string
	responseError   error
	stop            bool
}

// Post .
func (h2cReq *H2CRequest) Post() Request {
	h2cReq.resq.Method = "POST"
	return h2cReq
}

// Put .
func (h2cReq *H2CRequest) Put() Request {
	h2cReq.resq.Method = "PUT"
	return h2cReq
}

// Get .
func (h2cReq *H2CRequest) Get() Request {
	h2cReq.resq.Method = "GET"
	return h2cReq
}

// Delete .
func (h2cReq *H2CRequest) Delete() Request {
	h2cReq.resq.Method = "DELETE"
	return h2cReq
}

// Head .
func (h2cReq *H2CRequest) Head() Request {
	h2cReq.resq.Method = "HEAD"
	return h2cReq
}

// SetContext .
func (h2cReq *H2CRequest) SetContext(ctx context.Context) Request {
	h2cReq.ctx = ctx
	return h2cReq
}

// SetJSONBody .
func (h2cReq *H2CRequest) SetJSONBody(obj interface{}) Request {
	byts, e := Marshal(obj)
	if e != nil {
		h2cReq.reqe = e
		return h2cReq
	}

	h2cReq.resq.Body = ioutil.NopCloser(bytes.NewReader(byts))
	h2cReq.resq.ContentLength = int64(len(byts))
	h2cReq.resq.Header.Set("Content-Type", "application/json")
	return h2cReq
}

// SetBody .
func (h2cReq *H2CRequest) SetBody(byts []byte) Request {
	h2cReq.resq.Body = ioutil.NopCloser(bytes.NewReader(byts))
	h2cReq.resq.ContentLength = int64(len(byts))
	h2cReq.resq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return h2cReq
}

// ToJSON .
func (h2cReq *H2CRequest) ToJSON(obj interface{}) (r Response) {
	var body []byte
	r, body = h2cReq.singleflightDo()
	if r.Error != nil {
		return
	}
	r.Error = Unmarshal(body, obj)
	if r.Error != nil {
		r.Error = fmt.Errorf("%s, body:%s", r.Error.Error(), string(body))
	}
	return
}

// ToString .
func (h2cReq *H2CRequest) ToString() (value string, r Response) {
	var body []byte
	r, body = h2cReq.singleflightDo()
	if r.Error != nil {
		return
	}
	value = string(body)
	return
}

// ToBytes .
func (h2cReq *H2CRequest) ToBytes() (value []byte, r Response) {
	r, value = h2cReq.singleflightDo()
	return
}

// ToXML .
func (h2cReq *H2CRequest) ToXML(v interface{}) (r Response) {
	var body []byte
	r, body = h2cReq.singleflightDo()
	if r.Error != nil {
		return
	}

	h2cReq.httpRespone(&r)
	r.Error = xml.Unmarshal(body, v)
	return
}

// SetParam .
func (h2cReq *H2CRequest) SetParam(key string, value interface{}) Request {
	h2cReq.params.Add(key, fmt.Sprint(value))
	return h2cReq
}

// SetHeader .
func (h2cReq *H2CRequest) SetHeader(key, value string) Request {
	h2cReq.resq.Header.Add(key, value)
	return h2cReq
}

// URI .
func (h2cReq *H2CRequest) URL() string {
	params := h2cReq.params.Encode()
	if params != "" {
		return h2cReq.url + "?" + params
	}
	return h2cReq.url
}

func (h2cReq *H2CRequest) singleflightDo() (r Response, body []byte) {
	if h2cReq.singleflightKey == "" {
		r.Error = h2cReq.prepare()
		if r.Error != nil {
			return
		}
		handle(h2cReq)
		r.Error = h2cReq.responseError
		if r.Error != nil {
			return
		}
		body = h2cReq.body()
		h2cReq.httpRespone(&r)
		return
	}

	data, _, _ := h2cclientGroup.Do(h2cReq.singleflightKey, func() (interface{}, error) {
		var res Response
		res.Error = h2cReq.prepare()
		if r.Error != nil {
			return &singleflightData{Res: res}, nil
		}
		handle(h2cReq)
		res.Error = h2cReq.responseError
		if res.Error != nil {
			return &singleflightData{Res: res}, nil
		}

		h2cReq.httpRespone(&res)
		return &singleflightData{Res: res, Body: h2cReq.body()}, nil
	})

	sfdata := data.(*singleflightData)
	return sfdata.Res, sfdata.Body
}

func (h2cReq *H2CRequest) do() (e error) {
	if h2cReq.reqe != nil {
		return h2cReq.reqe
	}
	u, e := url.Parse(h2cReq.URL())
	if e != nil {
		return e
	}
	if h2cReq.ctx != nil {
		h2cReq.resq = h2cReq.resq.WithContext(h2cReq.ctx)
	}
	h2cReq.resq.URL = u
	h2cReq.resp, e = DefaultH2CClient.Do(h2cReq.resq)
	return
}

func (h2cReq *H2CRequest) body() (body []byte) {
	defer h2cReq.resp.Body.Close()
	if h2cReq.resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(h2cReq.resp.Body)
		if err != nil {
			return
		}
		body, _ = ioutil.ReadAll(reader)
		return body
	}
	body, _ = ioutil.ReadAll(h2cReq.resp.Body)
	return
}

func (h2cReq *H2CRequest) httpRespone(httpRespone *Response) {
	httpRespone.StatusCode = h2cReq.resp.StatusCode
	httpRespone.HTTP11 = false
	if h2cReq.resp.ProtoMajor == 1 && h2cReq.resp.ProtoMinor == 1 {
		httpRespone.HTTP11 = true
	}

	httpRespone.ContentLength = h2cReq.resp.ContentLength
	httpRespone.ContentType = h2cReq.resp.Header.Get("Content-Type")
	if httpRespone.Header == nil {
		httpRespone.Header = make(map[string]string)
	}
	for key, values := range h2cReq.resp.Header {
		if len(values) < 1 {
			continue
		}
		httpRespone.Header[key] = values[0]
	}
}

func (h2cReq *H2CRequest) Singleflight(key ...interface{}) Request {
	h2cReq.singleflightKey = fmt.Sprint(key...)
	return h2cReq
}

func (h2cReq *H2CRequest) prepare() (e error) {
	if h2cReq.reqe != nil {
		return h2cReq.reqe
	}

	u, e := url.Parse(h2cReq.URL())
	if e != nil {
		return
	}
	h2cReq.resq.URL = u
	return
}

func (h2cReq *H2CRequest) Next() {
	h2cReq.responseError = h2cReq.do()
}

func (h2cReq *H2CRequest) Stop(e ...error) {
	h2cReq.stop = true
	if len(e) > 0 {
		h2cReq.responseError = e[0]
		return
	}
	h2cReq.responseError = errors.New("Middleware stop")
}

func (h2cReq *H2CRequest) getStop() bool {
	return h2cReq.stop
}

func (h2cReq *H2CRequest) GetRequest() *http.Request {
	return h2cReq.resq
}

func (h2cReq *H2CRequest) GetRespone() (*http.Response, error) {
	return h2cReq.resp, h2cReq.responseError
}
