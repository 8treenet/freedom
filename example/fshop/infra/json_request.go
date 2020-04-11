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
}
