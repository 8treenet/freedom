package general

import (
	"time"

	"github.com/8treenet/freedom/general/requests"
	"github.com/kataras/iris"
)

// RepositoryAPI .
type RepositoryAPI struct {
	Runtime Runtime
}

// BeginRequest .
func (repo *RepositoryAPI) BeginRequest(rt Runtime) {
	repo.Runtime = rt
}

// Request .
type Request requests.Request

// NewFastRequest .
func (repo *RepositoryAPI) NewFastRequest(url string) Request {
	req := requests.NewFastRequest(url)
	var traceid string
	repo.Runtime.Store().Get(globalApp.traceIDKey, &traceid)
	req.SetHeader(globalApp.traceIDKey, traceid)
	return req
}

// NewH2CRequest .
func (repo *RepositoryAPI) NewH2CRequest(url string) Request {
	req := requests.NewH2CRequest(url)
	var traceid string
	repo.Runtime.Store().Get(globalApp.traceIDKey, &traceid)
	req.SetHeader(globalApp.traceIDKey, traceid)
	return req
}

func repositoryAPIRun(irisConf iris.Configuration) {
	sec := int64(5)
	if v, ok := irisConf.Other["repository_request_timeout"]; ok {
		sec = v.(int64)
	}
	requests.NewFastClient(time.Duration(sec) * time.Second)
	requests.Newh2cClient(time.Duration(sec) * time.Second)
}
