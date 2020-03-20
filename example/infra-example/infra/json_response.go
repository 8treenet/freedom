package infra

import (
	"strconv"

	"github.com/8treenet/extjson"
	"github.com/8treenet/freedom/general"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/hero"
)

type JSONResponse struct {
	Code        int
	Error       error
	contentType string
	content     []byte
	Object      interface{}
}

func (jrep JSONResponse) Dispatch(ctx context.Context) {
	jrep.contentType = "application/json"
	var repData struct {
		Code  int         `json:"code"`
		Error string      `json:"error"`
		Data  interface{} `json:"data,omitempty"`
	}
	repData.Data = jrep.Object
	repData.Code = jrep.Code
	if jrep.Error != nil {
		repData.Error = jrep.Error.Error()
	}
	if repData.Error != "" && repData.Code == 0 {
		repData.Code = iris.StatusNotImplemented
	}
	loggger := ctx.Values().Get("logger_impl")
	ctx.Values().Set("code", strconv.Itoa(repData.Code))

	var jsonErr error
	jrep.content, jsonErr = extjson.Marshal(repData)
	if loggger != nil {
		log := loggger.(general.Logger)
		if jsonErr != nil {
			log.Info("JSONResponse json error:", jsonErr, "from data", repData)
			return
		}
		log.Info("JSONResponse:", string(jrep.content))
	}
	hero.DispatchCommon(ctx, 0, jrep.contentType, jrep.content, nil, nil, true)
}
