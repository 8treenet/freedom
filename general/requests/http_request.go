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
	"strconv"
	"time"
)

var (
	httpclient      *http.Client
	httpclientGroup Group
)

func init() {
	NewHttpClient(10 * time.Second)
}

// Newhttpclient .
func NewHttpClient(rwTimeout time.Duration) {
	tran := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
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

	httpclient = &http.Client{
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
func (req *HttpRequest) SetHeader(key, value string) Request {
	req.resq.Header.Set(key, value)
	return req
}

// URI .
func (req *HttpRequest) URI() string {
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
	waitChanErr := make(chan error)
	defer close(waitChanErr)

	go func(waitErr chan error) {
		var err error
		now := time.Now()
		defer func() {
			code := ""
			pro := ""
			if req.resp != nil {
				code = strconv.Itoa(req.resp.StatusCode)
				pro = req.resp.Proto
			}
			host := ""
			if u, err := url.Parse(req.url); err == nil {
				host = u.Host
			}
			if err != nil {
				code = "error"
			}

			PrometheusImpl.HttpClientWithLabelValues(host, code, pro, req.resq.Method, now)
		}()
		req.resp, err = httpclient.Do(req.resq)
		waitErr <- err
	}(waitChanErr)

	if req.ctx == nil {
		if err := <-waitChanErr; err != nil {
			e = err
			return
		}

		code := req.resp.StatusCode
		if code >= 400 && code <= 600 {
			e = fmt.Errorf("The FastRequested URL returned error: %d", code)
			return
		}
		return
	}

	select {
	case <-time.After(httpclient.Timeout):
		e = errors.New("Timeout")
		return
	case <-req.ctx.Done():
		e = req.ctx.Err()
		return
	case err := <-waitChanErr:
		if err != nil {
			e = err
			return
		}
		code := req.resp.StatusCode
		if code >= 400 && code <= 600 {
			e = fmt.Errorf("The FastRequested URL returned error: %d", code)
			return
		}
		return
	}
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
	if httpRespone.Header == nil {
		httpRespone.Header = make(map[string]string)
	}
	for key, values := range req.resp.Header {
		if len(values) < 1 {
			continue
		}
		httpRespone.Header[key] = values[0]
	}
}

func (req *HttpRequest) Singleflight(key ...interface{}) Request {
	req.singleflightKey = fmt.Sprint(key...)
	return req
}

func (req *HttpRequest) prepare() (e error) {
	if req.reqe != nil {
		return req.reqe
	}

	u, e := url.Parse(req.URI())
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
