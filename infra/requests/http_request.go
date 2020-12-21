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
	"reflect"
	"strings"
)

// HTTPRequest .
type HTTPRequest struct {
	StdRequest      *http.Request
	Response        Response
	Params          url.Values
	RawURL          string
	SingleflightKey string
	stop            bool
	Client          Client
	nextIndex       int
	responeBody     []byte
	connTrace       *requestConnTrace
	h2c             bool
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

// Options .
func (req *HTTPRequest) Options() Request {
	req.StdRequest.Method = "OPTIONS"
	return req
}

// AddCookie .
func (req *HTTPRequest) AddCookie(c *http.Cookie) Request {
	req.StdRequest.AddCookie(c)
	return req
}

// SetJSONBody .
func (req *HTTPRequest) SetJSONBody(obj interface{}) Request {
	byts, e := Marshal(obj)
	if e != nil {
		req.Response.Error = e
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
func (req *HTTPRequest) ToJSON(obj interface{}) *Response {
	var body []byte
	body = req.singleflightDo(true)
	if req.Response.Error != nil {
		return &req.Response
	}
	req.Response.Error = Unmarshal(body, obj)
	if req.Response.Error != nil {
		req.Response.Error = fmt.Errorf("%s, body:%s", req.Response.Error.Error(), string(body))
	}
	return &req.Response
}

// ToString .
func (req *HTTPRequest) ToString() (string, *Response) {
	var body []byte
	body = req.singleflightDo(true)
	if req.Response.Error != nil {
		return "", &req.Response
	}
	value := string(body)
	return value, &req.Response
}

// ToBytes .
func (req *HTTPRequest) ToBytes() ([]byte, *Response) {
	value := req.singleflightDo(false)
	return value, &req.Response
}

// ToXML .
func (req *HTTPRequest) ToXML(v interface{}) *Response {
	var body []byte
	body = req.singleflightDo(true)
	if req.Response.Error != nil {
		return &req.Response
	}

	req.Response.Error = xml.Unmarshal(body, v)
	return &req.Response
}

// SetQueryParam .
func (req *HTTPRequest) SetQueryParam(key string, value interface{}) Request {
	reflectValue := reflect.ValueOf(value)
	if reflectValue.Kind() != reflect.Slice && reflectValue.Kind() != reflect.Array {
		req.Params.Add(key, fmt.Sprint(value))
		return req
	}

	for i := 0; i < reflectValue.Len(); i++ {
		req.Params.Add(key, fmt.Sprint(reflectValue.Index(i).Interface()))
	}
	return req
}

// SetQueryParams .
func (req *HTTPRequest) SetQueryParams(m map[string]interface{}) Request {
	for k, v := range m {
		reflectValue := reflect.ValueOf(v)
		if reflectValue.Kind() != reflect.Slice && reflectValue.Kind() != reflect.Array {
			req.Params.Add(k, fmt.Sprint(v))
			continue
		}
		for i := 0; i < reflectValue.Len(); i++ {
			req.Params.Add(k, fmt.Sprint(reflectValue.Index(i).Interface()))
		}
	}
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
		return req.RawURL + "?" + params
	}
	return req.RawURL
}

// SetClient .
func (req *HTTPRequest) SetClient(client Client) Request {
	if client == nil {
		panic("This is an undefined client")
	}
	req.Client = client
	return req
}

type singleflightData struct {
	Res  Response
	Body []byte
}

func (req *HTTPRequest) singleflightDo(noCopy bool) (body []byte) {
	if req.SingleflightKey == "" {
		req.Next()
		if req.Response.Error != nil {
			return
		}
		body = req.responeBody
		return
	}

	data, _, shared := httpclientGroup.Do(req.SingleflightKey, func() (interface{}, error) {
		req.Next()
		if req.Response.Error != nil {
			return &singleflightData{Res: *req.Response.Clone(), Body: []byte{}}, nil
		}
		if noCopy {
			return &singleflightData{Res: *req.Response.Clone(), Body: req.responeBody}, nil
		}

		bodyBuffer := make([]byte, len(req.responeBody))
		copy(bodyBuffer, req.responeBody)
		return &singleflightData{Res: *req.Response.Clone(), Body: bodyBuffer}, nil
	})
	if !shared {
		return req.responeBody
	}

	sfdata := data.(*singleflightData)
	req.Response = *sfdata.Res.Clone()
	if noCopy {
		return sfdata.Body
	}
	copyData := make([]byte, len(sfdata.Body))
	copy(copyData, sfdata.Body)
	return copyData
}

func (req *HTTPRequest) do() {
	if req.Response.Error != nil {
		return
	}
	if req.Response.Error = req.prepare(); req.Response.Error != nil {
		return
	}
	u, e := url.Parse(req.URL())
	if e != nil {
		req.Response.Error = e
		return
	}
	req.StdRequest.URL = u
	req.Response.stdResponse, req.Response.Error = req.Client.Do(req.StdRequest)
	if req.Response.Error != nil {
		return
	}
	req.readBody()
	req.fillingRespone()
	return
}

func (req *HTTPRequest) readBody() {
	defer func() {
		req.Response.stdResponse.Body.Close()
		if req.connTrace != nil {
			req.connTrace.done()
		}
	}()

	if strings.EqualFold(req.Response.Header.Get("Content-Encoding"), "gzip") && req.Response.ContentLength != 0 {
		reader, err := gzip.NewReader(req.Response.stdResponse.Body)
		if err != nil {
			req.Response.Error = err
			return
		}
		defer reader.Close()
		req.responeBody, req.Response.Error = ioutil.ReadAll(reader)
		return
	}

	req.responeBody, req.Response.Error = ioutil.ReadAll(req.Response.stdResponse.Body)
	return
}

func (req *HTTPRequest) fillingRespone() {
	if req.Response.Error != nil {
		return

	}
	std := req.Response.stdResponse
	req.Response.Status = std.Status
	req.Response.StatusCode = std.StatusCode
	req.Response.Proto = std.Proto
	req.Response.ProtoMajor = std.ProtoMajor
	req.Response.ProtoMinor = std.ProtoMinor
	req.Response.Header = std.Header
	req.Response.ContentLength = std.ContentLength
	req.Response.Uncompressed = std.Uncompressed
	req.Response.HTTP11 = false
	if req.Response.ProtoMajor == 1 && req.Response.ProtoMinor == 1 {
		req.Response.HTTP11 = true
	}
	req.Response.ContentType = std.Header.Get("Content-Type")
	if req.connTrace != nil {
		req.Response.traceInfo = req.connTrace.traceInfo()
	}
}

// Singleflight .
func (req *HTTPRequest) Singleflight(key ...interface{}) Request {
	req.SingleflightKey = fmt.Sprint(key...)
	return req
}

func (req *HTTPRequest) prepare() (e error) {
	if req.Response.Error != nil {
		return req.Response.Error
	}

	u, e := url.Parse(req.URL())
	if e != nil {
		req.Response.Error = e
		return
	}
	req.StdRequest.URL = u
	return
}

// Next .
func (req *HTTPRequest) Next() {
	if len(middlewares) == 0 {
		req.do()
		return
	}
	if req.IsStopped() {
		return
	}

	if req.nextIndex == len(middlewares) {
		req.do()
		return
	}

	req.nextIndex = req.nextIndex + 1
	middlewares[req.nextIndex-1](req)
}

// Stop .
func (req *HTTPRequest) Stop(e ...error) {
	req.stop = true
	if len(e) > 0 {
		req.Response.Error = e[0]
		return
	}
	req.Response.Error = errors.New("Middleware stop")
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
func (req *HTTPRequest) GetRespone() *Response {
	return &req.Response
}

// GetStdRequest .
func (req *HTTPRequest) GetStdRequest() *http.Request {
	return req.StdRequest
}

// Header .
func (req *HTTPRequest) Header() http.Header {
	return req.StdRequest.Header
}

// GetResponeBody .
func (req *HTTPRequest) GetResponeBody() []byte {
	return req.responeBody
}

// Context .
func (req *HTTPRequest) Context() context.Context {
	return req.StdRequest.Context()
}

// EnableTrace .
func (req *HTTPRequest) EnableTrace() Request {
	if req.connTrace != nil {
		return req
	}
	req.connTrace = &requestConnTrace{}
	req.StdRequest = req.StdRequest.WithContext(req.connTrace.createContext(req.StdRequest.Context()))
	return req
}

// EnableTraceFromMiddleware .
func (req *HTTPRequest) EnableTraceFromMiddleware() {
	req.EnableTrace()
}

// WithContext .
func (req *HTTPRequest) WithContext(ctx context.Context) Request {
	req.StdRequest = req.StdRequest.WithContext(ctx)
	return req
}

// WithContextFromMiddleware .
func (req *HTTPRequest) WithContextFromMiddleware(ctx context.Context) {
	req.WithContext(ctx)
}

// SetClientFromMiddleware .
func (req *HTTPRequest) SetClientFromMiddleware(client Client) {
	req.SetClient(client)
}

// IsH2C .
func (req *HTTPRequest) IsH2C() bool {
	return req.h2c
}
