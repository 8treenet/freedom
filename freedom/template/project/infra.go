package project

import "fmt"

func init() {
	content["/infra/response.go"] = jsonResponseTemplate()
	content["/infra/request.go"] = jsonRequestTemplate()
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
			initiator.BindInfra(false, func() *Request {
				return &Request{}
			})
			initiator.InjectController(func(ctx freedom.Context) (com *Request) {
				initiator.GetInfra(ctx, &com)
				return
			})
		})
	}
	
	// Request .
	type Request struct {
		freedom.Infra
	}
	
	// BeginRequest .
	func (req *Request) BeginRequest(worker freedom.Worker) {
		req.Infra.BeginRequest(worker)
	}
	
	// ReadJSON .
	func (req *Request) ReadJSON(obj interface{}) error {
		rawData, err := ioutil.ReadAll(req.Worker.IrisContext().Request().Body)
		if err != nil {
			return err
		}
		err = extjson.Unmarshal(rawData, obj)
		if err != nil {
			return err
		}
	
		return validate.Struct(obj)
	}
	
	// ReadQuery .
	func (req *Request) ReadQuery(obj interface{}) error {
		if err := req.Worker.IrisContext().ReadQuery(obj); err != nil {
			return err
		}
		return validate.Struct(obj)
	}
	
	// ReadForm .
	func (req *Request) ReadForm(obj interface{}) error {
		if err := req.Worker.IrisContext().ReadForm(obj); err != nil {
			return err
		}
		return validate.Struct(obj)
	}	
	`
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
