package requests

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	fclient *fasthttp.Client
)

func init() {
	NewFastClient(10 * time.Second)
}

// NewFastClient .
func NewFastClient(rwTimeout time.Duration) {
	fclient = &fasthttp.Client{ReadTimeout: rwTimeout, WriteTimeout: rwTimeout, Name: "ht/1.0.0(fasthttp)"}
}

// NewFastRequest .
func NewFastRequest(uri, bus string) Request {
	result := new(FastRequest)
	result.resq = fasthttp.AcquireRequest()
	result.resp = fasthttp.AcquireResponse()
	result.params = make(map[string]interface{})
	result.url = uri
	result.bus = bus
	result.skipBus = false
	return result
}

// FastRequest .
type FastRequest struct {
	resq    *fasthttp.Request
	resp    *fasthttp.Response
	reqe    error
	params  map[string]interface{}
	url     string
	bus     string
	skipBus bool
}

// Post .
func (fr *FastRequest) Post() Request {
	fr.resq.Header.SetMethod("POST")
	return fr
}

// Put .
func (fr *FastRequest) Put() Request {
	fr.resq.Header.SetMethod("PUT")
	return fr
}

// Get .
func (fr *FastRequest) Get() Request {
	fr.resq.Header.SetMethod("GET")
	return fr
}

// Delete .
func (fr *FastRequest) Delete() Request {
	fr.resq.Header.SetMethod("DELETE")
	return fr
}

// SetJSONBody .
func (fr *FastRequest) SetJSONBody(obj interface{}) Request {
	byts, e := json.Marshal(obj)
	if e != nil {
		fr.reqe = e
		return fr
	}
	fr.resq.SetBody(byts)
	fr.resq.Header.SetContentLength(len(byts))
	fr.resq.Header.SetContentType("application/json")
	return fr
}

// SetBody .
func (fr *FastRequest) SetBody(byts []byte) Request {
	fr.resq.SetBody(byts)
	fr.resq.Header.SetContentLength(len(byts))
	fr.resq.Header.SetContentType("application/x-www-form-urlencoded")
	return fr
}

// SetParam .
func (fr *FastRequest) SetParam(key string, value interface{}) Request {
	fr.params[key] = value
	return fr
}

// SkipBus .
func (fr *FastRequest) SkipBus() Request {
	fr.skipBus = true
	return fr
}

// ToJSON .
func (fr *FastRequest) ToJSON(obj interface{}) (r Response) {
	defer func() {
		fasthttp.ReleaseRequest(fr.resq)
		fasthttp.ReleaseResponse(fr.resp)
		fr.resq = nil
		fr.resq = nil
	}()
	r.Error = fr.do()
	fr.httpRespone(&r)
	if r.Error != nil {
		return
	}
	r.Error = json.Unmarshal(fr.resp.Body(), obj)
	return
}

// ToString .
func (fr *FastRequest) ToString() (value string, r Response) {
	defer func() {
		fasthttp.ReleaseRequest(fr.resq)
		fasthttp.ReleaseResponse(fr.resp)
		fr.resq = nil
		fr.resq = nil
	}()
	r.Error = fr.do()
	fr.httpRespone(&r)
	if r.Error != nil {
		return
	}

	value = string(fr.resp.Body())
	return
}

// ToBytes .
func (fr *FastRequest) ToBytes() (value []byte, r Response) {
	defer func() {
		fasthttp.ReleaseRequest(fr.resq)
		fasthttp.ReleaseResponse(fr.resp)
		fr.resq = nil
		fr.resq = nil
	}()
	r.Error = fr.do()
	fr.httpRespone(&r)
	if r.Error != nil {
		return
	}

	value = fr.resp.Body()
	return
}

// ToXML .
func (fr *FastRequest) ToXML(v interface{}) (r Response) {
	defer func() {
		fasthttp.ReleaseRequest(fr.resq)
		fasthttp.ReleaseResponse(fr.resp)
		fr.resq = nil
		fr.resq = nil
	}()
	r.Error = fr.do()
	fr.httpRespone(&r)
	if r.Error != nil {
		return
	}
	r.Error = xml.Unmarshal(fr.resp.Body(), v)
	return
}

// SetHeader .
func (fr *FastRequest) SetHeader(key, value string) Request {
	fr.resq.Header.Set(key, value)
	return fr
}

// URI .
func (fr *FastRequest) URI() string {
	result := fr.url
	if len(fr.params) > 0 {
		uris := strings.Split(result, "?")
		uri := ""
		if len(uris) > 1 {
			uri = uris[1]
		}
		index := 0
		for key, value := range fr.params {
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

func (fr *FastRequest) do() error {
	if fr.reqe != nil {
		return fr.reqe
	}

	if !fr.skipBus {
		fr.SetHeader("Freedom-Bus", fr.bus)
	}

	fr.resq.SetRequestURI(fr.URI())
	if err := fclient.Do(fr.resq, fr.resp); err != nil {
		return err
	}

	code := fr.resp.Header.StatusCode()
	if code >= 400 {
		return fmt.Errorf("The FastRequested URL returned error: %d", code)
	}
	return nil
}

func (fr *FastRequest) httpRespone(httpRespone *Response) {
	httpRespone.StatusCode = fr.resp.Header.StatusCode()
	httpRespone.HTTP11 = fr.resp.Header.IsHTTP11()
	httpRespone.ContentLength = int64(fr.resp.Header.ContentLength())
	httpRespone.ContentType = string(fr.resp.Header.ContentType())
	httpRespone.Header = make(map[string]string)

	fr.resp.Header.VisitAll(func(key, value []byte) {
		httpRespone.Header[string(key)] = string(value)
	})
}
