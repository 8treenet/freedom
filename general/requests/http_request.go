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
	"strings"
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
	result.params = make(map[string]interface{})
	result.url = rawurl
	return result
}

// HttpRequest .
type HttpRequest struct {
	resq            *http.Request
	resp            *http.Response
	reqe            error
	params          map[string]interface{}
	url             string
	bus             string
	ctx             context.Context
	singleflightKey string
}

// Post .
func (hr *HttpRequest) Post() Request {
	hr.resq.Method = "POST"
	return hr
}

// Put .
func (hr *HttpRequest) Put() Request {
	hr.resq.Method = "PUT"
	return hr
}

// Get .
func (hr *HttpRequest) Get() Request {
	hr.resq.Method = "GET"
	return hr
}

// Delete .
func (hr *HttpRequest) Delete() Request {
	hr.resq.Method = "DELETE"
	return hr
}

// Head .
func (hr *HttpRequest) Head() Request {
	hr.resq.Method = "HEAD"
	return hr
}

// SetContext .
func (hr *HttpRequest) SetContext(ctx context.Context) Request {
	hr.ctx = ctx
	return hr
}

// SetJSONBody .
func (hr *HttpRequest) SetJSONBody(obj interface{}) Request {
	byts, e := Marshal(obj)
	if e != nil {
		hr.reqe = e
		return hr
	}

	hr.resq.Body = ioutil.NopCloser(bytes.NewReader(byts))
	hr.resq.ContentLength = int64(len(byts))
	hr.resq.Header.Set("Content-Type", "application/json")
	return hr
}

// SetBody .
func (hr *HttpRequest) SetBody(byts []byte) Request {
	hr.resq.Body = ioutil.NopCloser(bytes.NewReader(byts))
	hr.resq.ContentLength = int64(len(byts))
	hr.resq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return hr
}

// ToJSON .
func (hr *HttpRequest) ToJSON(obj interface{}) (r Response) {
	var body []byte
	r, body = hr.singleflightDo()
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
func (hr *HttpRequest) ToString() (value string, r Response) {
	var body []byte
	r, body = hr.singleflightDo()
	if r.Error != nil {
		return
	}
	value = string(body)
	return
}

// ToBytes .
func (hr *HttpRequest) ToBytes() (value []byte, r Response) {
	r, value = hr.singleflightDo()
	return
}

// ToXML .
func (hr *HttpRequest) ToXML(v interface{}) (r Response) {
	var body []byte
	r, body = hr.singleflightDo()
	if r.Error != nil {
		return
	}

	hr.httpRespone(&r)
	r.Error = xml.Unmarshal(body, v)
	return
}

// SetParam .
func (hr *HttpRequest) SetParam(key string, value interface{}) Request {
	hr.params[key] = value
	return hr
}

// SetHeader .
func (hr *HttpRequest) SetHeader(key, value string) Request {
	hr.resq.Header.Set(key, value)
	return hr
}

// URI .
func (hr *HttpRequest) URI() string {
	result := hr.url
	if len(hr.params) > 0 {
		uris := strings.Split(result, "?")
		uri := ""
		if len(uris) > 1 {
			uri = uris[1]
		}
		index := 0
		for key, value := range hr.params {
			if index > 0 {
				uri += "&"
			}
			uri += key + "=" + fmt.Sprint(value)
			index++
		}
		result = uris[0] + "?" + uri
	}
	return result
}

type singleflightData struct {
	Res  Response
	Body []byte
}

func (hr *HttpRequest) singleflightDo() (r Response, body []byte) {
	if hr.singleflightKey == "" {
		r.Error = hr.do()
		if r.Error != nil {
			return
		}
		body = hr.body()
		hr.httpRespone(&r)
		return
	}

	data, _, _ := httpclientGroup.Do(hr.singleflightKey, func() (interface{}, error) {
		var res Response
		res.Error = hr.do()
		if res.Error != nil {
			return &singleflightData{Res: res}, nil
		}

		hr.httpRespone(&res)
		return &singleflightData{Res: res, Body: hr.body()}, nil
	})

	sfdata := data.(*singleflightData)
	return sfdata.Res, sfdata.Body
}

func (hr *HttpRequest) do() (e error) {
	if hr.reqe != nil {
		return hr.reqe
	}

	u, e := url.Parse(hr.URI())
	if e != nil {
		return
	}
	hr.resq.URL = u

	waitChanErr := make(chan error)
	defer close(waitChanErr)

	go func(waitErr chan error) {
		var err error
		now := time.Now()
		defer func() {
			code := ""
			pro := ""
			if hr.resp != nil {
				code = strconv.Itoa(hr.resp.StatusCode)
				pro = hr.resp.Proto
			}
			host := ""
			if u, err := url.Parse(hr.url); err == nil {
				host = u.Host
			}
			if err != nil {
				code = "error"
			}

			PrometheusImpl.HttpClientWithLabelValues(host, code, pro, hr.resq.Method, now)
		}()
		hr.resp, err = httpclient.Do(hr.resq)
		waitErr <- err
	}(waitChanErr)

	if hr.ctx == nil {
		if err := <-waitChanErr; err != nil {
			e = err
			return
		}

		code := hr.resp.StatusCode
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
	case <-hr.ctx.Done():
		e = hr.ctx.Err()
		return
	case err := <-waitChanErr:
		if err != nil {
			e = err
			return
		}
		code := hr.resp.StatusCode
		if code >= 400 && code <= 600 {
			e = fmt.Errorf("The FastRequested URL returned error: %d", code)
			return
		}
		return
	}
}

func (hr *HttpRequest) body() (body []byte) {
	defer hr.resp.Body.Close()
	if hr.resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(hr.resp.Body)
		if err != nil {
			return
		}
		body, _ = ioutil.ReadAll(reader)
		return body
	}
	body, _ = ioutil.ReadAll(hr.resp.Body)
	return
}

func (hr *HttpRequest) httpRespone(httpRespone *Response) {
	httpRespone.StatusCode = hr.resp.StatusCode
	httpRespone.HTTP11 = false
	if hr.resp.ProtoMajor == 1 && hr.resp.ProtoMinor == 1 {
		httpRespone.HTTP11 = true
	}

	httpRespone.ContentLength = hr.resp.ContentLength
	httpRespone.ContentType = hr.resp.Header.Get("Content-Type")
	if httpRespone.Header == nil {
		httpRespone.Header = make(map[string]string)
	}
	for key, values := range hr.resp.Header {
		if len(values) < 1 {
			continue
		}
		httpRespone.Header[key] = values[0]
	}
}

func (hr *HttpRequest) Singleflight(key ...interface{}) Request {
	hr.singleflightKey = fmt.Sprint(key...)
	return hr
}
