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
	"time"

	"golang.org/x/net/http2"
)

var (
	h2cclient *http.Client
)

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

func NewH2CRequest(rawurl string) Request {
	result := new(H2CRequest)
	u, err := url.Parse(rawurl)
	if err != nil {
		result.reqe = err
		return result
	}
	req := &http.Request{
		URL:    u,
		Header: make(http.Header),
	}
	result.resq = req
	return result
}

// H2CRequest .
type H2CRequest struct {
	resq *http.Request
	resp *http.Response
	reqe error
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

// ToJSON .
func (hr *H2CRequest) ToJSON(obj interface{}, httpRespones ...*Response) (e error) {
	e = hr.do()
	if e != nil {
		return
	}
	if len(httpRespones) > 0 {
		hr.httpRespone(httpRespones[0])
	}

	return json.Unmarshal(hr.body(), obj)
}

// ToString .
func (hr *H2CRequest) ToString(httpRespones ...*Response) (value string, e error) {
	e = hr.do()
	if e != nil {
		return
	}

	if len(httpRespones) > 0 {
		hr.httpRespone(httpRespones[0])
	}
	value = string(hr.body())
	return
}

// ToBytes .
func (hr *H2CRequest) ToBytes(httpRespones ...*Response) (value []byte, e error) {
	e = hr.do()
	if e != nil {
		return
	}
	if len(httpRespones) > 0 {
		hr.httpRespone(httpRespones[0])
	}
	value = hr.body()
	return
}

// ToXML .
func (hr *H2CRequest) ToXML(v interface{}, httpRespones ...*Response) (e error) {

	e = hr.do()
	if e != nil {
		return
	}
	if len(httpRespones) > 0 {
		hr.httpRespone(httpRespones[0])
	}
	return xml.Unmarshal(hr.body(), v)
}

// SetHeader .
func (hr *H2CRequest) SetHeader(key, value string) Request {
	hr.resq.Header.Set(key, value)
	return hr
}

func (hr *H2CRequest) do() (e error) {
	if hr.reqe != nil {
		return hr.reqe
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
