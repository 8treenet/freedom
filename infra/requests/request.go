package requests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Unmarshal .
var Unmarshal func(data []byte, v interface{}) error

// Marshal .
var Marshal func(v interface{}) ([]byte, error)

func init() {
	if Unmarshal == nil {
		Unmarshal = json.Unmarshal
	}
	if Marshal == nil {
		Marshal = json.Marshal
	}
}

// Request .
type Request interface {
	Post() Request
	Put() Request
	Get() Request
	Delete() Request
	Head() Request
	SetJSONBody(obj interface{}) Request
	SetBody(byts []byte) Request
	ToJSON(obj interface{}) Response
	ToString() (string, Response)
	ToBytes() ([]byte, Response)
	ToXML(v interface{}) Response
	SetParam(key string, value interface{}) Request
	URL() string
	WithContext(context.Context) Request
	Singleflight(key ...interface{}) Request
	SetHeader(header http.Header) Request
	AddHeader(key, value string) Request
	Header() http.Header
	GetStdRequest() interface{}
}

// Response .
type Response struct {
	Error         error
	Header        http.Header
	ContentLength int64
	ContentType   string
	StatusCode    int
	HTTP11        bool
}

// NewHTTPRequest .
func NewHTTPRequest(rawurl string) Request {
	result := new(HTTPRequest)
	req := &http.Request{
		Header: make(http.Header),
	}
	result.StdRequest = req
	result.Params = make(url.Values)
	result.url = rawurl
	result.stop = false
	result.Client = DefaultHTTPClient
	return result
}

// NewH2CRequest .
func NewH2CRequest(rawurl string) Request {
	result := new(H2CRequest)
	req := &http.Request{
		Header: make(http.Header),
	}
	result.StdRequest = req
	result.Params = make(url.Values)
	result.url = rawurl
	result.stop = false
	result.Client = DefaultH2CClient
	return result
}
