package requests

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// HTTPRequest .
type HTTPRequest struct {
	StdRequest      *http.Request
	StdResponse     *http.Response
	Error           error
	Params          url.Values
	url             string
	SingleflightKey string
	ResponseError   error
	stop            bool
	Client          *http.Client
}

// Post .
func (req *HTTPRequest) Post() Request {
	req.StdRequest.Method = "POST"
	return req
}

// Put .
func (req *HTTPRequest) Put() Request {
	req.StdRequest.Method = "PUT"
	return req
}

// Get .
func (req *HTTPRequest) Get() Request {
	req.StdRequest.Method = "GET"
	return req
}

// Delete .
func (req *HTTPRequest) Delete() Request {
	req.StdRequest.Method = "DELETE"
	return req
}

// Head .
func (req *HTTPRequest) Head() Request {
	req.StdRequest.Method = "HEAD"
	return req
}

// WithContext .
func (req *HTTPRequest) WithContext(ctx context.Context) Request {
	req.StdRequest = req.StdRequest.WithContext(ctx)
	return req
}

// SetJSONBody .
func (req *HTTPRequest) SetJSONBody(obj interface{}) Request {
	byts, e := Marshal(obj)
	if e != nil {
		req.Error = e
		return req
	}

	req.StdRequest.Body = ioutil.NopCloser(bytes.NewReader(byts))
	req.StdRequest.ContentLength = int64(len(byts))
	req.StdRequest.Header.Set("Content-Type", "application/json")
	return req
}

// SetBody .
func (req *HTTPRequest) SetBody(byts []byte) Request {
	req.StdRequest.Body = ioutil.NopCloser(bytes.NewReader(byts))
	req.StdRequest.ContentLength = int64(len(byts))
	req.StdRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// ToJSON .
func (req *HTTPRequest) ToJSON(obj interface{}) (r Response) {
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
func (req *HTTPRequest) ToString() (value string, r Response) {
	var body []byte
	r, body = req.singleflightDo()
	if r.Error != nil {
		return
	}
	value = string(body)
	return
}

// ToBytes .
func (req *HTTPRequest) ToBytes() (value []byte, r Response) {
	r, value = req.singleflightDo()
	return
}

// ToXML .
func (req *HTTPRequest) ToXML(v interface{}) (r Response) {
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
func (req *HTTPRequest) SetParam(key string, value interface{}) Request {
	req.Params.Add(key, fmt.Sprint(value))
	return req
}

// SetHeader .
func (req *HTTPRequest) SetHeader(header http.Header) Request {
	req.StdRequest.Header = header
	return req
}

// AddHeader .
func (req *HTTPRequest) AddHeader(key, value string) Request {
	req.StdRequest.Header.Add(key, value)
	return req
}

// URL .
func (req *HTTPRequest) URL() string {
	params := req.Params.Encode()
	if params != "" {
		return req.url + "?" + params
	}
	return req.url
}

type singleflightData struct {
	Res  Response
	Body []byte
}

func (req *HTTPRequest) singleflightDo() (r Response, body []byte) {
	if req.SingleflightKey == "" {
		r.Error = req.prepare()
		if r.Error != nil {
			return
		}
		handle(req)
		r.Error = req.ResponseError
		if r.Error != nil {
			return
		}
		body = req.body()
		req.httpRespone(&r)
		return
	}

	data, _, _ := httpclientGroup.Do(req.SingleflightKey, func() (interface{}, error) {
		var res Response
		res.Error = req.prepare()
		if r.Error != nil {
			return &singleflightData{Res: res}, nil
		}
		handle(req)
		res.Error = req.ResponseError
		if res.Error != nil {
			return &singleflightData{Res: res}, nil
		}

		req.httpRespone(&res)
		return &singleflightData{Res: res, Body: req.body()}, nil
	})

	sfdata := data.(*singleflightData)
	return sfdata.Res, sfdata.Body
}

func (req *HTTPRequest) do() (e error) {
	if req.Error != nil {
		return req.Error
	}
	u, e := url.Parse(req.URL())
	if e != nil {
		return e
	}
	req.StdRequest.URL = u
	req.StdResponse, e = req.Client.Do(req.StdRequest)
	return
}

func (req *HTTPRequest) body() (body []byte) {
	defer req.StdResponse.Body.Close()
	if req.StdResponse.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(req.StdResponse.Body)
		if err != nil {
			return
		}
		body, _ = ioutil.ReadAll(reader)
		return body
	}
	body, _ = ioutil.ReadAll(req.StdResponse.Body)
	return
}

func (req *HTTPRequest) httpRespone(httpRespone *Response) {
	httpRespone.StatusCode = req.StdResponse.StatusCode
	httpRespone.HTTP11 = false
	if req.StdResponse.ProtoMajor == 1 && req.StdResponse.ProtoMinor == 1 {
		httpRespone.HTTP11 = true
	}

	httpRespone.ContentLength = req.StdResponse.ContentLength
	httpRespone.ContentType = req.StdResponse.Header.Get("Content-Type")
	httpRespone.Header = req.StdResponse.Header
}

// Singleflight .
func (req *HTTPRequest) Singleflight(key ...interface{}) Request {
	req.SingleflightKey = fmt.Sprint(key...)
	return req
}

func (req *HTTPRequest) prepare() (e error) {
	if req.Error != nil {
		return req.Error
	}

	u, e := url.Parse(req.URL())
	if e != nil {
		return
	}
	req.StdRequest.URL = u
	return
}

// Next .
func (req *HTTPRequest) Next() {
	req.ResponseError = req.do()
}

// Stop .
func (req *HTTPRequest) Stop(e ...error) {
	req.stop = true
	if len(e) > 0 {
		req.ResponseError = e[0]
		return
	}
	req.ResponseError = errors.New("Middleware stop")
}

// IsStopped .
func (req *HTTPRequest) IsStopped() bool {
	return req.stop
}

// GetRequest .
func (req *HTTPRequest) GetRequest() *http.Request {
	return req.StdRequest
}

// GetRespone .
func (req *HTTPRequest) GetRespone() (*http.Response, error) {
	return req.StdResponse, req.ResponseError
}

// GetStdRequest .
func (req *HTTPRequest) GetStdRequest() interface{} {
	return req.StdRequest
}

// Header .
func (req *HTTPRequest) Header() http.Header {
	return req.StdRequest.Header
}
