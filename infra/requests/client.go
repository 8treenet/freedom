package requests

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/sync/singleflight"
)

func init() {
	InitH2CClient(10 * time.Second)
	InitHTTPClient(10 * time.Second)
}

var (
	// defaultH2CClient .
	defaultH2CClient Client
	// defaultHTTPClient .
	defaultHTTPClient Client

	h2cclientGroup  singleflight.Group
	httpclientGroup singleflight.Group
)

// SetHTTPClient .
func SetHTTPClient(client Client) {
	defaultHTTPClient = client
}

// SetH2CClient .
func SetH2CClient(client Client) {
	defaultH2CClient = client
}

// Client .
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientImpl .
type ClientImpl struct {
	*http.Client
}

// Do .
func (client *ClientImpl) Do(req *http.Request) (*http.Response, error) {
	return client.Client.Do(req)
}

// InitHTTPClient .
func InitHTTPClient(rwTimeout time.Duration, connectTimeout ...time.Duration) {
	sec := 2 * time.Second
	if len(connectTimeout) > 0 {
		sec = connectTimeout[0]
	}
	defaultHTTPClient = NewHTTPClient(rwTimeout, sec)
}

// InitH2CClient .
func InitH2CClient(rwTimeout time.Duration, connectTimeout ...time.Duration) {
	sec := 2 * time.Second
	if len(connectTimeout) > 0 {
		sec = connectTimeout[0]
	}
	defaultH2CClient = NewH2CClient(rwTimeout, sec)
}

// NewHTTPClient .
func NewHTTPClient(rwTimeout time.Duration, connectTimeout time.Duration) *ClientImpl {
	tran := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   connectTimeout,
			KeepAlive: 15 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          512,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   100,
	}

	return &ClientImpl{Client: &http.Client{
		Transport: tran,
		Timeout:   rwTimeout,
	}}
}

// NewH2CClient .
func NewH2CClient(rwTimeout time.Duration, connectTimeout time.Duration) *ClientImpl {
	tran := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			fun := timeoutDialer(connectTimeout)
			return fun(network, addr)
		},
	}

	return &ClientImpl{Client: &http.Client{
		Transport: tran,
		Timeout:   rwTimeout,
	}}
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
