package requests

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
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
	fclient = &fasthttp.Client{ReadTimeout: rwTimeout, WriteTimeout: rwTimeout, Name: "freedom/1.0.0(fasthttp)"}
}

// NewFastRequest .
func NewFastRequest(uri string, bus ...string) Request {
	result := new(FastRequest)
	result.resq = &fasthttp.Request{}
	result.resp = &fasthttp.Response{}
	result.params = make(map[string]interface{})
	result.url = uri
	result.bus = ""
	result.skipBus = false
	if len(bus) > 0 {
		result.bus = bus[0]
	}
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
	ctx     context.Context
}

// SetContext .
func (fr *FastRequest) SetContext(ctx context.Context) Request {
	fr.ctx = ctx
	return fr
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
	r.Error = fr.do()
	if r.Error != nil {
		return
	}

	fr.httpRespone(&r)
	body := fr.resp.Body()
	r.Error = json.Unmarshal(body, obj)
	if r.Error != nil {
		r.Error = fmt.Errorf("%s, body:%s", r.Error.Error(), string(body))
	}
	return
}

// ToString .
func (fr *FastRequest) ToString() (value string, r Response) {
	r.Error = fr.do()
	if r.Error != nil {
		return
	}

	fr.httpRespone(&r)
	value = string(fr.resp.Body())
	return
}

// ToBytes .
func (fr *FastRequest) ToBytes() (value []byte, r Response) {
	r.Error = fr.do()
	if r.Error != nil {
		return
	}

	fr.httpRespone(&r)
	value = fr.resp.Body()
	return
}

// ToXML .
func (fr *FastRequest) ToXML(v interface{}) (r Response) {
	r.Error = fr.do()
	if r.Error != nil {
		return
	}

	fr.httpRespone(&r)
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

	if !fr.skipBus && fr.bus != "" {
		fr.SetHeader("x-freedom-bus", fr.bus)
	}

	waitChanErr := make(chan error)
	defer close(waitChanErr)
	go func(waitErr chan error) {
		fr.resq.SetRequestURI(fr.URI())
		waitErr <- fclient.Do(fr.resq, fr.resp)
	}(waitChanErr)

	if fr.ctx == nil {
		if err := <-waitChanErr; err != nil {
			return err
		}
		code := fr.resp.Header.StatusCode()
		if code >= 400 && code <= 600 {
			return fmt.Errorf("The FastRequested URL returned error: %d", code)
		}
		return nil
	}

	select {
	case <-time.After(fclient.ReadTimeout):
		return errors.New("Timeout")
	case <-fr.ctx.Done():
		return fr.ctx.Err()
	case err := <-waitChanErr:
		if err != nil {
			return err
		}
		code := fr.resp.Header.StatusCode()
		if code >= 400 && code <= 600 {
			return fmt.Errorf("The FastRequested URL returned error: %d", code)
		}
		return nil
	}
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
