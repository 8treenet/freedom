package requests

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/sync/singleflight"
)

var (
	// DefaultH2CClient .
	DefaultH2CClient *http.Client
	h2cclientGroup   singleflight.Group
	// DefaultHTTPClient .
	DefaultHTTPClient *http.Client
	httpclientGroup   singleflight.Group
)

func init() {
	InitH2cClient(10 * time.Second)
	InitHTTPClient(10 * time.Second)
}

// InitHTTPClient .
func InitHTTPClient(rwTimeout time.Duration, connectTimeout ...time.Duration) {
	t := 2 * time.Second
	if len(connectTimeout) > 0 {
		t = connectTimeout[0]
	}

	tran := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   t,
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

	DefaultHTTPClient = &http.Client{
		Transport: tran,
		Timeout:   rwTimeout,
	}
}

// InitH2cClient .
func InitH2cClient(rwTimeout time.Duration, connectTimeout ...time.Duration) {
	tran := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			t := 2 * time.Second
			if len(connectTimeout) > 0 {
				t = connectTimeout[0]
			}
			fun := timeoutDialer(t)
			return fun(network, addr)
		},
	}

	DefaultH2CClient = &http.Client{
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
