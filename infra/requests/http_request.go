package requests

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	DefaultHttpClient *http.Client
	httpclientGroup   Group
)

func init() {
	NewHttpClient(10 * time.Second)
}

// Newhttpclient .
func NewHttpClient(rwTimeout time.Duration, connectTimeout ...time.Duration) {
	t := 2 * time.Second
	if len(connectTimeout) > 0 {
		t = connectTimeout[0]
	}

	tran := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   t,
			KeepAlive: 15 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          512,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   100,
	}

	DefaultHttpClient = &http.Client{
		Transport: tran,
		Timeout:   rwTimeout,
	}
}

func NewHttpRequest(rawurl string) Request {
	result := new(HttpRequest)
	req := &http.Request{
		Header: make(http.Header),
	}
	result.resq = req
	result.params = make(url.Values)
	result.url = rawurl
	result.stop = false
	return result
}

// HttpRequest .
type HttpRequest struct {
	resq            *http.Request
	resp            *http.Response
	reqe            error
	params          url.Values
	url             string
	bus             string
	ctx             context.Context
	singleflightKey string
	responseError   error
	stop            bool
}

// Post .
func (req *HttpRequest) Post() Request {
	req.resq.Method = "POST"
	return req
}

// Put .
func (req *HttpRequest) Put() Request {
	req.resq.Method = "PUT"
	return req
}

// Get .
func (req *HttpRequest) Get() Request {
	req.resq.Method = "GET"
	return req
}

// Delete .
func (req *HttpRequest) Delete() Request {
	req.resq.Method = "DELETE"
	return req
}

// Head .
func (req *HttpRequest) Head() Request {
	req.resq.Method = "HEAD"
	return req
}

// SetContext .
func (req *HttpRequest) SetContext(ctx context.Context) Request {
	req.ctx = ctx
	return req
}

// SetJSONBody .
func (req *HttpRequest) SetJSONBody(obj interface{}) Request {
	byts, e := Marshal(obj)
	if e != nil {
		req.reqe = e
		return req
	}

	req.resq.Body = ioutil.NopCloser(bytes.NewReader(byts))
	req.resq.ContentLength = int64(len(byts))
	req.resq.Header.Set("Content-Type", "application/json")
	return req
}

// SetBody .
func (req *HttpRequest) SetBody(byts []byte) Request {
	req.resq.Body = ioutil.NopCloser(bytes.NewReader(byts))
	req.resq.ContentLength = int64(len(byts))
	req.resq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// ToJSON .
func (req *HttpRequest) ToJSON(obj interface{}) (r Response) {
	var body []byte
	r, body = req.singleflightDo()
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
func (req *HttpRequest) ToString() (value string, r Response) {
	var body []byte
	r, body = req.singleflightDo()
	if r.Error != nil {
		return
	}
	value = string(body)
	return
}

// ToBytes .
func (req *HttpRequest) ToBytes() (value []byte, r Response) {
	r, value = req.singleflightDo()
	return
}

// ToXML .
func (req *HttpRequest) ToXML(v interface{}) (r Response) {
	var body []byte
	r, body = req.singleflightDo()
	if r.Error != nil {
		return
	}

	req.httpRespone(&r)
	r.Error = xml.Unmarshal(body, v)
	return
}

// SetParam .
func (req *HttpRequest) SetParam(key string, value interface{}) Request {
	req.params.Add(key, fmt.Sprint(value))
	return req
}

// SetHeader .
func (req *HttpRequest) SetHeader(header http.Header) Request {
	req.resq.Header = header
	return req
}

// AddHeader .
func (req *HttpRequest) AddHeader(key, value string) Request {
	req.resq.Header.Add(key, value)
	return req
}

// URI .
func (req *HttpRequest) URL() string {
	params := req.params.Encode()
	if params != "" {
		return req.url + "?" + params
	}
	return req.url
}

type singleflightData struct {
	Res  Response
	Body []byte
}

func (req *HttpRequest) singleflightDo() (r Response, body []byte) {
	if req.singleflightKey == "" {
		r.Error = req.prepare()
		if r.Error != nil {
			return
		}
		handle(req)
		r.Error = req.responseError
		if r.Error != nil {
			return
		}
		body = req.body()
		req.httpRespone(&r)
		return
	}

	data, _, _ := httpclientGroup.Do(req.singleflightKey, func() (interface{}, error) {
		var res Response
		res.Error = req.prepare()
		if r.Error != nil {
			return &singleflightData{Res: res}, nil
		}
		handle(req)
		res.Error = req.responseError
		if res.Error != nil {
			return &singleflightData{Res: res}, nil
		}

		req.httpRespone(&res)
		return &singleflightData{Res: res, Body: req.body()}, nil
	})

	sfdata := data.(*singleflightData)
	return sfdata.Res, sfdata.Body
}

func (req *HttpRequest) do() (e error) {
	if req.reqe != nil {
		return req.reqe
	}
	u, e := url.Parse(req.URL())
	if e != nil {
		return e
	}
	if req.ctx != nil {
		req.resq = req.resq.WithContext(req.ctx)
	}
	req.resq.URL = u
	req.resp, e = DefaultHttpClient.Do(req.resq)
	return
}

func (req *HttpRequest) body() (body []byte) {
	defer req.resp.Body.Close()
	if req.resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(req.resp.Body)
		if err != nil {
			return
		}
		body, _ = ioutil.ReadAll(reader)
		return body
	}
	body, _ = ioutil.ReadAll(req.resp.Body)
	return
}

func (req *HttpRequest) httpRespone(httpRespone *Response) {
	httpRespone.StatusCode = req.resp.StatusCode
	httpRespone.HTTP11 = false
	if req.resp.ProtoMajor == 1 && req.resp.ProtoMinor == 1 {
		httpRespone.HTTP11 = true
	}

	httpRespone.ContentLength = req.resp.ContentLength
	httpRespone.ContentType = req.resp.Header.Get("Content-Type")
	httpRespone.Header = req.resp.Header
}

func (req *HttpRequest) Singleflight(key ...interface{}) Request {
	req.singleflightKey = fmt.Sprint(key...)
	return req
}

func (req *HttpRequest) prepare() (e error) {
	if req.reqe != nil {
		return req.reqe
	}

	u, e := url.Parse(req.URL())
	if e != nil {
		return
	}
	req.resq.URL = u
	return
}

func (req *HttpRequest) Next() {
	req.responseError = req.do()
}

func (req *HttpRequest) Stop(e ...error) {
	req.stop = true
	if len(e) > 0 {
		req.responseError = e[0]
		return
	}
	req.responseError = errors.New("Middleware stop")
}

func (req *HttpRequest) getStop() bool {
	return req.stop
}

func (req *HttpRequest) GetRequest() *http.Request {
	return req.resq
}

func (req *HttpRequest) GetRespone() (*http.Response, error) {
	return req.resp, req.responseError
}
