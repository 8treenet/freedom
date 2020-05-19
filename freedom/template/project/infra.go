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
		"github.com/8treenet/freedom/infra/requests"
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
		freedom.Prepare(func(initiator freedom.Initiator) {
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
	"strconv"

	"github.com/8treenet/extjson"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/hero"
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
		Code  int          %sjson:"code"%s
		Error string       %sjson:"error"%s
		Data  interface{}  %sjson:"data,omitempty"%s
	}

	repData.Data = jrep.Object
	repData.Code = jrep.Code
	if jrep.Error != nil {
		repData.Error = jrep.Error.Error()
	}
	if repData.Error != "" && repData.Code == 0 {
		repData.Code = 501
	}
	ctx.Values().Set("code", strconv.Itoa(repData.Code))

	jrep.content, _ = extjson.Marshal(repData)
	ctx.Values().Set("response", string(jrep.content))
	hero.DispatchCommon(ctx, 0, jrep.contentType, jrep.content, nil, nil, true)
}

`
	return fmt.Sprintf(result, "`", "`", "`", "`", "`", "`")
}
