package requests

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

var (
	h2cclient *http.Client
)

func init() {
	Newh2cClient(10 * time.Second)
}

// Newh2cClient .
func Newh2cClient(rwTimeout time.Duration) {
	tran := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			fun := timeoutDialer(3 * time.Second)
			return fun(network, addr)
		},
	}

	h2cclient = &http.Client{
		Transport: tran,
		Timeout:   rwTimeout,
	}
}

// timeoutDialer returns functions of connection dialer with timeout settings for http.Transport Dial field.
func timeoutDialer(cTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		return conn, err
	}
}

func NewH2CRequest(rawurl, bus string) Request {
	result := new(H2CRequest)
	req := &http.Request{
		Header: make(http.Header),
	}
	result.resq = req
	result.params = make(map[string]interface{})
	result.url = rawurl
	result.bus = bus
	result.skipBus = false
	return result
}

// H2CRequest .
type H2CRequest struct {
	resq    *http.Request
	resp    *http.Response
	reqe    error
	params  map[string]interface{}
	url     string
	bus     string
	skipBus bool
}

// Post .
func (hr *H2CRequest) Post() Request {
	hr.resq.Method = "POST"
	return hr
}

// Put .
func (hr *H2CRequest) Put() Request {
	hr.resq.Method = "PUT"
	return hr
}

// Get .
func (hr *H2CRequest) Get() Request {
	hr.resq.Method = "GET"
	return hr
}

// // Delete .
func (hr *H2CRequest) Delete() Request {
	hr.resq.Method = "DELETE"
	return hr
}

// SetJSONBody .
func (hr *H2CRequest) SetJSONBody(obj interface{}) Request {
	byts, e := json.Marshal(obj)
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
func (hr *H2CRequest) SetBody(byts []byte) Request {
	hr.resq.Body = ioutil.NopCloser(bytes.NewReader(byts))
	hr.resq.ContentLength = int64(len(byts))
	hr.resq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return hr
}

// SkipBus .
func (hr *H2CRequest) SkipBus() Request {
	hr.skipBus = true
	return hr
}

// ToJSON .
func (hr *H2CRequest) ToJSON(obj interface{}) (r Response) {
	r.Error = hr.do()
	hr.httpRespone(&r)
	if r.Error != nil {
		return
	}
	r.Error = json.Unmarshal(hr.body(), obj)
	return
}

// ToString .
func (hr *H2CRequest) ToString() (value string, r Response) {
	r.Error = hr.do()
	hr.httpRespone(&r)
	if r.Error != nil {
		return
	}
	value = string(hr.body())
	return
}

// ToBytes .
func (hr *H2CRequest) ToBytes() (value []byte, r Response) {
	r.Error = hr.do()
	hr.httpRespone(&r)
	if r.Error != nil {
		return
	}
	value = hr.body()
	return
}

// ToXML .
func (hr *H2CRequest) ToXML(v interface{}) (r Response) {
	r.Error = hr.do()
	hr.httpRespone(&r)
	if r.Error != nil {
		return
	}
	r.Error = xml.Unmarshal(hr.body(), v)
	return
}

// SetParam .
func (hr *H2CRequest) SetParam(key string, value interface{}) Request {
	hr.params[key] = value
	return hr
}

// SetHeader .
func (hr *H2CRequest) SetHeader(key, value string) Request {
	hr.resq.Header.Set(key, value)
	return hr
}

// URI .
func (hr *H2CRequest) URI() string {
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

func (hr *H2CRequest) do() (e error) {
	if hr.reqe != nil {
		return hr.reqe
	}

	u, e := url.Parse(hr.URI())
	if e != nil {
		return
	}
	hr.resq.URL = u
	if !hr.skipBus {
		hr.SetHeader("Freedom-Bus", hr.bus)
	}

	hr.resp, e = h2cclient.Do(hr.resq)
	if e != nil {
		return
	}

	code := hr.resp.StatusCode
	if code >= 400 {
		return fmt.Errorf("The FastRequested URL returned error: %d", code)
	}
	return
}

func (hr *H2CRequest) body() (body []byte) {
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

func (hr *H2CRequest) httpRespone(httpRespone *Response) {
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
