package project

import "fmt"

func init() {
	content["/infra/json_response.go"] = jsonResponseTemplate()
	content["/infra/json_request.go"] = jsonRequestTemplate()
	content["/infra/json_init.go"] = jsonInitTemplate()
}

func jsonInitTemplate() string {
	return `
	package infra
	import (
		"github.com/8treenet/extjson"
		"github.com/8treenet/freedom/general/requests"
	)
	
	func init() {
		//More references github.com/8treenet/extjson
		extjson.SetDefaultOption(extjson.ExtJSONEntityOption{
			NamedStyle:       extjson.NamedStyleLowerCamelCase,
			SliceNotNull:     true, //空数组不返回null, 返回[]
			StructPtrNotNull: true, //nil结构体指针不返回null, 返回{}})
		})
		requests.Unmarshal = extjson.Unmarshal
		requests.Marshal = extjson.Marshal
	}
	`
}
func jsonRequestTemplate() string {
	return `
	package infra

	import (
		"io/ioutil"
	
		"github.com/8treenet/extjson"
		"github.com/8treenet/freedom"
		"gopkg.in/go-playground/validator.v9"
	)
	
	var validate *validator.Validate
	
	func init() {
		validate = validator.New()
		freedom.Booting(func(initiator freedom.Initiator) {
			initiator.BindInfra(false, func() *JSONRequest {
				return &JSONRequest{}
			})
			initiator.InjectController(func(ctx freedom.Context) (com *JSONRequest) {
				initiator.GetInfra(ctx, &com)
				return
			})
		})
	}
	
	// JSONRequest .
	type JSONRequest struct {
		freedom.Infra
	}
	
	// BeginRequest .
	func (req *JSONRequest) BeginRequest(rt freedom.Runtime) {
		req.Infra.BeginRequest(rt)
	}
	
	// ReadBodyJSON .
	func (req *JSONRequest) ReadBodyJSON(obj interface{}) error {
		rawData, err := ioutil.ReadAll(req.Runtime.Ctx().Request().Body)
		if err != nil {
			return err
		}
		err = extjson.Unmarshal(rawData, obj)
		if err != nil {
			return err
		}
	
		return validate.Struct(obj)
	}
	
	// ReadQueryJSON .
	func (req *JSONRequest) ReadQueryJSON(obj interface{}) error {
		return nil
	}`
}
func jsonResponseTemplate() string {
	result := `
	package infra

	import (
		"github.com/8treenet/extjson"
		"github.com/8treenet/freedom/general"
		"github.com/kataras/iris"
		"github.com/kataras/iris/context"
		"github.com/kataras/iris/hero"
	)
	
	type JSONResponse struct {
		// Code automatically sets default 501 if the error is not empty
		Code        int
		Error       error
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
	
		if jrep.Path != "" && jrep.Error == nil {
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
			Code  int
			Error string       %sjson:"error"%s
			Data  interface{}  %sjson:"data,omitempty"%s
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
		hero.DispatchCommon(ctx, jrep.Code, jrep.contentType, jrep.content, nil, nil, true)
	}`

	return fmt.Sprintf(result, "`", "`", "`", "`")
}
