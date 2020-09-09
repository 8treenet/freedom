package requests

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

// HTTPTraceInfo .
type HTTPTraceInfo struct {
	DNSLookup    time.Duration
	ConnTime     time.Duration
	TCPConnTime  time.Duration
	TLSHandshake time.Duration

	ServerTime   time.Duration
	ResponseTime time.Duration
	TotalTime    time.Duration

	IsConnReused bool

	IsConnWasIdle bool
	ConnIdleTime  time.Duration
}

type requestConnTrace struct {
	getConn              time.Time
	dnsStart             time.Time
	dnsDone              time.Time
	connectDone          time.Time
	tlsHandshakeStart    time.Time
	tlsHandshakeDone     time.Time
	gotConn              time.Time
	gotFirstResponseByte time.Time
	endTime              time.Time
	gotConnInfo          httptrace.GotConnInfo
}

func (trace *requestConnTrace) createContext(ctx context.Context) context.Context {
	return httptrace.WithClientTrace(
		ctx,
		&httptrace.ClientTrace{
			DNSStart: func(_ httptrace.DNSStartInfo) {
				trace.dnsStart = time.Now()
			},
			DNSDone: func(_ httptrace.DNSDoneInfo) {
				trace.dnsDone = time.Now()
			},
			ConnectStart: func(_, _ string) {
				if trace.dnsDone.IsZero() {
					trace.dnsDone = time.Now()
				}
				if trace.dnsStart.IsZero() {
					trace.dnsStart = trace.dnsDone
				}
			},
			ConnectDone: func(net, addr string, err error) {
				trace.connectDone = time.Now()
			},
			GetConn: func(_ string) {
				trace.getConn = time.Now()
			},
			GotConn: func(ci httptrace.GotConnInfo) {
				trace.gotConn = time.Now()
				trace.gotConnInfo = ci
			},
			GotFirstResponseByte: func() {
				trace.gotFirstResponseByte = time.Now()
			},
			TLSHandshakeStart: func() {
				trace.tlsHandshakeStart = time.Now()
			},
			TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
				trace.tlsHandshakeDone = time.Now()
			},
		},
	)
}

func (trace *requestConnTrace) traceInfo() HTTPTraceInfo {
	ti := HTTPTraceInfo{
		DNSLookup:     trace.dnsDone.Sub(trace.dnsStart),
		TLSHandshake:  trace.tlsHandshakeDone.Sub(trace.tlsHandshakeStart),
		ServerTime:    trace.gotFirstResponseByte.Sub(trace.gotConn),
		IsConnReused:  trace.gotConnInfo.Reused,
		IsConnWasIdle: trace.gotConnInfo.WasIdle,
		ConnIdleTime:  trace.gotConnInfo.IdleTime,
	}

	if trace.gotConnInfo.Reused {
		ti.TotalTime = trace.endTime.Sub(trace.getConn)
	} else {
		ti.TotalTime = trace.endTime.Sub(trace.dnsStart)
	}

	if !trace.connectDone.IsZero() {
		ti.TCPConnTime = trace.connectDone.Sub(trace.dnsDone)
	}

	if !trace.gotConn.IsZero() {
		ti.ConnTime = trace.gotConn.Sub(trace.getConn)
	}

	if !trace.gotFirstResponseByte.IsZero() {
		ti.ResponseTime = trace.endTime.Sub(trace.gotFirstResponseByte)
	}

	return ti
}

func (trace *requestConnTrace) done() {
	trace.endTime = time.Now()
}
