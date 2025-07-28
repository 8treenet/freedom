package requests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Unmarshal Deserialization, default to the official JSON processor.
var Unmarshal func(data []byte, v interface{}) error

// Marshal Serialization, the default official JSON processor.
var Marshal func(v interface{}) ([]byte, error)

func init() {
	if Unmarshal == nil {
		Unmarshal = json.Unmarshal
	}
	if Marshal == nil {
		Marshal = json.Marshal
	}
}

// Request The interface definition of the HTTP request package.
type Request interface {
	// HTTP POST method.
	Post() Request
	// HTTP PUT method.
	Put() Request
	// HTTP GET method.
	Get() Request
	// HTTP DELETE method.
	Delete() Request
	// HTTP HEAD method.
	Head() Request
	// HTTP OPTIONS method.
	Options() Request
	// Set up JSON data.
	SetJSONBody(obj interface{}) Request
	// Set the original data.
	SetBody(byts []byte) Request
	// Set up Form data.
	SetFormBody(url.Values) Request
	// Set up File data.
	SetFile(field, file string) Request
	// The data is returned after the request and converted to JSON data.
	ToJSON(obj interface{}) *Response
	// The data is returned after the request and converted into a string.
	ToString() (string, *Response)
	// The data is returned after the request and converted into a []byte.
	ToBytes() ([]byte, *Response)
	// The data is returned after the request and converted to XML data.
	ToXML(v interface{}) *Response
	// Set the request parameters.
	SetQueryParam(key string, value interface{}) Request
	// Set the request parameters.
	SetQueryParams(map[string]interface{}) Request
	URL() string
	Context() context.Context
	WithContext(context.Context) Request
	// idempotent request processing, need to fill in KEY.
	Singleflight(key ...interface{}) Request
	// Set up HTTP Header information.
	SetHeader(header http.Header) Request
	// Set up HTTP Header value.
	SetHeaderValue(key, value string) Request
	// Returns the header of the current request.
	Header() http.Header
	// Returns the Request of the standard library.
	GetStdRequest() *http.Request
	// Add cookie information.
	AddCookie(*http.Cookie) Request
	// Turn on link tracking.
	EnableTrace() Request
	// Set up client.
	SetClient(client Client) Request
}

// NewHTTPRequest Create an HTTP request object.
// The URL of the incoming HTTP.
func NewHTTPRequest(rawurl string) Request {
	result := new(httpRequest)
	req := &http.Request{
		Header: make(http.Header),
	}
	result.StdRequest = req
	result.Params = make(url.Values)
	result.RawURL = rawurl
	result.Client = defaultHTTPClient
	return result
}

// NewH2CRequest Create an HTTP H2C request object.
// The URL of the incoming HTTP.
func NewH2CRequest(rawurl string) Request {
	result := new(httpRequest)
	req := &http.Request{
		Header: make(http.Header),
	}
	result.StdRequest = req
	result.Params = make(url.Values)
	result.RawURL = rawurl
	result.Client = defaultH2CClient
	result.h2c = true
	return result
}
