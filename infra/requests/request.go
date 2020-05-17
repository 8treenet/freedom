package requests

import (
	"context"
	"encoding/json"
)

var Unmarshal func(data []byte, v interface{}) error
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
	SetHeader(key, value string) Request
	SetParam(key string, value interface{}) Request
	URL() string
	SetContext(context.Context) Request
	Singleflight(key ...interface{}) Request
}

// Response .
type Response struct {
	Error         error
	Header        map[string]string
	ContentLength int64
	ContentType   string
	StatusCode    int
	HTTP11        bool
}
