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

// H2CRequest .
type H2CRequest struct {
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
func (h2cReq *H2CRequest) Post() Request {
	h2cReq.StdRequest.Method = "POST"
	return h2cReq
}

// Put .
func (h2cReq *H2CRequest) Put() Request {
	h2cReq.StdRequest.Method = "PUT"
	return h2cReq
}

// Get .
func (h2cReq *H2CRequest) Get() Request {
	h2cReq.StdRequest.Method = "GET"
	return h2cReq
}

// Delete .
func (h2cReq *H2CRequest) Delete() Request {
	h2cReq.StdRequest.Method = "DELETE"
	return h2cReq
}

// Head .
func (h2cReq *H2CRequest) Head() Request {
	h2cReq.StdRequest.Method = "HEAD"
	return h2cReq
}

// WithContext .
func (h2cReq *H2CRequest) WithContext(ctx context.Context) Request {
	h2cReq.StdRequest = h2cReq.StdRequest.WithContext(ctx)
	return h2cReq
}

// SetJSONBody .
func (h2cReq *H2CRequest) SetJSONBody(obj interface{}) Request {
	byts, e := Marshal(obj)
	if e != nil {
		h2cReq.Error = e
		return h2cReq
	}

	h2cReq.StdRequest.Body = ioutil.NopCloser(bytes.NewReader(byts))
	h2cReq.StdRequest.ContentLength = int64(len(byts))
	h2cReq.StdRequest.Header.Set("Content-Type", "application/json")
	return h2cReq
}

// SetBody .
func (h2cReq *H2CRequest) SetBody(byts []byte) Request {
	h2cReq.StdRequest.Body = ioutil.NopCloser(bytes.NewReader(byts))
	h2cReq.StdRequest.ContentLength = int64(len(byts))
	h2cReq.StdRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
	h2cReq.Params.Add(key, fmt.Sprint(value))
	return h2cReq
}

// SetHeader .
func (h2cReq *H2CRequest) SetHeader(header http.Header) Request {
	h2cReq.StdRequest.Header = header
	return h2cReq
}

// AddHeader .
func (h2cReq *H2CRequest) AddHeader(key, value string) Request {
	h2cReq.StdRequest.Header.Add(key, value)
	return h2cReq
}

// URL .
func (h2cReq *H2CRequest) URL() string {
	params := h2cReq.Params.Encode()
	if params != "" {
		return h2cReq.url + "?" + params
	}
	return h2cReq.url
}

func (h2cReq *H2CRequest) singleflightDo() (r Response, body []byte) {
	if h2cReq.SingleflightKey == "" {
		r.Error = h2cReq.prepare()
		if r.Error != nil {
			return
		}
		handle(h2cReq)
		r.Error = h2cReq.ResponseError
		if r.Error != nil {
			return
		}
		body = h2cReq.body()
		h2cReq.httpRespone(&r)
		return
	}

	data, _, _ := h2cclientGroup.Do(h2cReq.SingleflightKey, func() (interface{}, error) {
		var res Response
		res.Error = h2cReq.prepare()
		if r.Error != nil {
			return &singleflightData{Res: res}, nil
		}
		handle(h2cReq)
		res.Error = h2cReq.ResponseError
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
	if h2cReq.Error != nil {
		return h2cReq.Error
	}
	u, e := url.Parse(h2cReq.URL())
	if e != nil {
		return e
	}
	h2cReq.StdRequest.URL = u
	h2cReq.StdResponse, e = h2cReq.Client.Do(h2cReq.StdRequest)
	return
}

func (h2cReq *H2CRequest) body() (body []byte) {
	defer h2cReq.StdResponse.Body.Close()
	if h2cReq.StdResponse.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(h2cReq.StdResponse.Body)
		if err != nil {
			return
		}
		body, _ = ioutil.ReadAll(reader)
		return body
	}
	body, _ = ioutil.ReadAll(h2cReq.StdResponse.Body)
	return
}

func (h2cReq *H2CRequest) httpRespone(httpRespone *Response) {
	httpRespone.StatusCode = h2cReq.StdResponse.StatusCode
	httpRespone.HTTP11 = false
	if h2cReq.StdResponse.ProtoMajor == 1 && h2cReq.StdResponse.ProtoMinor == 1 {
		httpRespone.HTTP11 = true
	}

	httpRespone.ContentLength = h2cReq.StdResponse.ContentLength
	httpRespone.ContentType = h2cReq.StdResponse.Header.Get("Content-Type")
	httpRespone.Header = h2cReq.StdResponse.Header
}

// Singleflight .
func (h2cReq *H2CRequest) Singleflight(key ...interface{}) Request {
	h2cReq.SingleflightKey = fmt.Sprint(key...)
	return h2cReq
}

func (h2cReq *H2CRequest) prepare() (e error) {
	if h2cReq.Error != nil {
		return h2cReq.Error
	}

	u, e := url.Parse(h2cReq.URL())
	if e != nil {
		return
	}
	h2cReq.StdRequest.URL = u
	return
}

// Next .
func (h2cReq *H2CRequest) Next() {
	h2cReq.ResponseError = h2cReq.do()
}

// Stop .
func (h2cReq *H2CRequest) Stop(e ...error) {
	h2cReq.stop = true
	if len(e) > 0 {
		h2cReq.ResponseError = e[0]
		return
	}
	h2cReq.ResponseError = errors.New("Middleware stop")
}

// IsStopped .
func (h2cReq *H2CRequest) IsStopped() bool {
	return h2cReq.stop
}

// GetRequest .
func (h2cReq *H2CRequest) GetRequest() *http.Request {
	return h2cReq.StdRequest
}

// GetRespone .
func (h2cReq *H2CRequest) GetRespone() (*http.Response, error) {
	return h2cReq.StdResponse, h2cReq.ResponseError
}

// GetStdRequest .
func (h2cReq *H2CRequest) GetStdRequest() interface{} {
	return h2cReq.StdRequest
}

// Header .
func (h2cReq *H2CRequest) Header() http.Header {
	return h2cReq.StdRequest.Header
}
