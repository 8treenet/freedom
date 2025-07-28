package requests

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// ErrDefaultMiddleware The middleware stops and the request returns the default error.
var ErrDefaultMiddleware = errors.New("middleware stop")

// httpRequest .
type httpRequest struct {
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
func (req *httpRequest) Post() Request {
	req.StdRequest.Method = "POST"
	return req
}

// Put .
func (req *httpRequest) Put() Request {
	req.StdRequest.Method = "PUT"
	return req
}

// Get .
func (req *httpRequest) Get() Request {
	req.StdRequest.Method = "GET"
	return req
}

// Delete .
func (req *httpRequest) Delete() Request {
	req.StdRequest.Method = "DELETE"
	return req
}

// Head .
func (req *httpRequest) Head() Request {
	req.StdRequest.Method = "HEAD"
	return req
}

// Options .
func (req *httpRequest) Options() Request {
	req.StdRequest.Method = "OPTIONS"
	return req
}

// AddCookie .
func (req *httpRequest) AddCookie(c *http.Cookie) Request {
	req.StdRequest.AddCookie(c)
	return req
}

// SetJSONBody .
func (req *httpRequest) SetJSONBody(obj interface{}) Request {
	byts, e := Marshal(obj)
	if e != nil {
		req.Response.Error = e
		return req
	}

	req.StdRequest.Body = io.NopCloser(bytes.NewReader(byts))
	req.StdRequest.ContentLength = int64(len(byts))
	req.StdRequest.Header.Set("Content-Type", "application/json")
	return req
}

// SetBody .
func (req *httpRequest) SetBody(byts []byte) Request {
	req.StdRequest.Body = io.NopCloser(bytes.NewReader(byts))
	req.StdRequest.ContentLength = int64(len(byts))
	req.StdRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// SetFormBody .
func (req *httpRequest) SetFormBody(v url.Values) Request {
	reader := strings.NewReader(v.Encode())
	req.StdRequest.Body = io.NopCloser(reader)
	req.StdRequest.ContentLength = int64(reader.Len())
	req.StdRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Post()
	return req
}

// SetFile .
func (req *httpRequest) SetFile(field, file string) Request {
	f, err := os.Open(file)
	if err != nil {
		req.Response.Error = err
		return req
	}
	defer f.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, e := writer.CreateFormFile(field, filepath.Base(file))
	if e != nil {
		req.Response.Error = e
		return req
	}

	_, err = io.Copy(part, f)
	if err != nil {
		req.Response.Error = err
		return req
	}
	if err = writer.Close(); err != nil {
		req.Response.Error = e
		return req
	}

	req.StdRequest.Body = io.NopCloser(bytes.NewReader(body.Bytes()))
	req.StdRequest.ContentLength = int64(body.Len())
	req.StdRequest.Header.Set("Content-Type", writer.FormDataContentType())
	req.Post()
	return req
}

// ToJSON .
func (req *httpRequest) ToJSON(obj interface{}) *Response {
	var body []byte
	body = req.singleflightDo(true)
	if req.Response.Error != nil {
		return &req.Response
	}
	req.Response.Error = Unmarshal(body, obj)
	if req.Response.Error != nil {
		req.Response.Error = fmt.Errorf("ToJSON %s data:%s, %w", req.GetStdRequest().URL.String(), string(body), req.Response.Error)
	}
	return &req.Response
}

// ToString .
func (req *httpRequest) ToString() (string, *Response) {
	var body []byte
	body = req.singleflightDo(true)
	if req.Response.Error != nil {
		return "", &req.Response
	}
	value := string(body)
	return value, &req.Response
}

// ToBytes .
func (req *httpRequest) ToBytes() ([]byte, *Response) {
	value := req.singleflightDo(false)
	return value, &req.Response
}

// ToXML .
func (req *httpRequest) ToXML(v interface{}) *Response {
	var body []byte
	body = req.singleflightDo(true)
	if req.Response.Error != nil {
		return &req.Response
	}

	req.Response.Error = xml.Unmarshal(body, v)
	if req.Response.Error != nil {
		req.Response.Error = fmt.Errorf("ToXML %s data:%s, %w", req.GetStdRequest().URL.String(), string(body), req.Response.Error)
	}
	return &req.Response
}

// SetQueryParam .
func (req *httpRequest) SetQueryParam(key string, value interface{}) Request {
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
func (req *httpRequest) SetQueryParams(m map[string]interface{}) Request {
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
func (req *httpRequest) SetHeader(header http.Header) Request {
	req.StdRequest.Header = header
	return req
}

// SetHeaderValue .
func (req *httpRequest) SetHeaderValue(key, value string) Request {
	req.StdRequest.Header.Set(key, value)
	return req
}

// URL .
func (req *httpRequest) URL() string {
	params := req.Params.Encode()
	if params != "" {
		return req.RawURL + "?" + params
	}
	return req.RawURL
}

// SetClient .
func (req *httpRequest) SetClient(client Client) Request {
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

func (req *httpRequest) singleflightDo(noCopy bool) (body []byte) {
	if req.SingleflightKey == "" {
		if req.Response.Error = req.prepare(); req.Response.Error != nil {
			return
		}
		req.Next()
		if req.Response.Error != nil {
			return
		}
		body = req.responeBody
		return
	}

	data, _, shared := httpclientGroup.Do(req.SingleflightKey, func() (interface{}, error) {
		if req.Response.Error = req.prepare(); req.Response.Error != nil {
			return &singleflightData{Res: *req.Response.Clone(), Body: []byte{}}, nil
		}
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

func (req *httpRequest) do() {
	if req.Response.Error != nil {
		return
	}
	req.Response.stdResponse, req.Response.Error = req.Client.Do(req.StdRequest)
	if req.Response.Error != nil {
		return
	}
	req.readBody()
	req.fillingRespone()
	return
}

func (req *httpRequest) readBody() {
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

func (req *httpRequest) fillingRespone() {
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
func (req *httpRequest) Singleflight(key ...interface{}) Request {
	req.SingleflightKey = fmt.Sprint(key...)
	return req
}

func (req *httpRequest) prepare() (e error) {
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
func (req *httpRequest) Next() {
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
func (req *httpRequest) Stop(e ...error) {
	req.stop = true
	if len(e) > 0 {
		req.Response.Error = e[0]
	} else {
		req.Response.Error = ErrDefaultMiddleware
	}

	req.Response.Error = fmt.Errorf("%s %s: %w", req.StdRequest.Method, req.StdRequest.URL.String(), req.Response.Error)
}

// IsStopped .
func (req *httpRequest) IsStopped() bool {
	return req.stop
}

// GetRequest .
func (req *httpRequest) GetRequest() *http.Request {
	return req.StdRequest
}

// GetRespone .
func (req *httpRequest) GetRespone() *Response {
	return &req.Response
}

// GetStdRequest .
func (req *httpRequest) GetStdRequest() *http.Request {
	return req.StdRequest
}

// Header .
func (req *httpRequest) Header() http.Header {
	return req.StdRequest.Header
}

// GetResponeBody .
func (req *httpRequest) GetResponeBody() []byte {
	return req.responeBody
}

// Context .
func (req *httpRequest) Context() context.Context {
	return req.StdRequest.Context()
}

// EnableTrace .
func (req *httpRequest) EnableTrace() Request {
	if req.connTrace != nil {
		return req
	}
	req.connTrace = &requestConnTrace{}
	req.StdRequest = req.StdRequest.WithContext(req.connTrace.createContext(req.StdRequest.Context()))
	return req
}

// EnableTraceFromMiddleware .
func (req *httpRequest) EnableTraceFromMiddleware() {
	req.EnableTrace()
}

// WithContext .
func (req *httpRequest) WithContext(ctx context.Context) Request {
	req.StdRequest = req.StdRequest.WithContext(ctx)
	return req
}

// WithContextFromMiddleware .
func (req *httpRequest) WithContextFromMiddleware(ctx context.Context) {
	req.WithContext(ctx)
}

// SetClientFromMiddleware .
func (req *httpRequest) SetClientFromMiddleware(client Client) {
	req.SetClient(client)
}

// IsH2C .
func (req *httpRequest) IsH2C() bool {
	return req.h2c
}
