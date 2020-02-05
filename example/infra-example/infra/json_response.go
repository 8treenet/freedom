package infra

import (
	"github.com/8treenet/extjson"
	"github.com/8treenet/freedom/general"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/hero"
)

func init() {
	//More references github.com/8treenet/extjson
	extjson.SetDefaultOption(extjson.ExtJSONEntityOption{
		NamedStyle:       extjson.NamedStyleLowerCamelCase,
		SliceNotNull:     true, //空数组不返回null, 返回[]
		StructPtrNotNull: true, //nil结构体指针不返回null, 返回{}})
	})
}

type JSONResponse struct {
	// Code automatically sets default 501 if the error is not empty
	Code        int
	Err         error
	contentType string
	content     []byte

	// If not nil then it will fire that as "application/json" or the
	// "ContentType" if not empty.
	Object interface{}

	// If Path is not empty then it will redirect
	// the client to this Path, if Code is >= 300 and < 400
	// then it will use that Code to do the redirection, otherwise
	// StatusFound(302) or StatusSeeOther(303) for post methods will be used.
	// Except when err != nil.
	Path string
}

func (jrep JSONResponse) Dispatch(ctx context.Context) {
	jrep.contentType = "application/json"

	if jrep.Path != "" && jrep.Err == nil {
		// it's not a redirect valid status
		if jrep.Code < 300 || jrep.Code >= 400 {
			if ctx.Method() == "POST" {
				jrep.Code = 303 // StatusSeeOther
			}
			jrep.Code = 302 // StatusFound
		}
		ctx.Redirect(jrep.Path, jrep.Code)
		return
	}

	var repData struct {
		Code int
		Err  string      `json:"error"`
		Data interface{} `json:"data,omitempty"`
	}
	repData.Data = jrep.Object
	repData.Code = jrep.Code
	if jrep.Err != nil {
		repData.Err = jrep.Err.Error()
	}
	if repData.Err != "" && repData.Code == 0 {
		repData.Code = iris.StatusNotImplemented
	}
	loggger := ctx.Values().Get("logger_impl")

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
	if jsonErr != nil {
		return
	}
	hero.DispatchCommon(ctx, jrep.Code, jrep.contentType, jrep.content, nil, jrep.Err, true)
}
