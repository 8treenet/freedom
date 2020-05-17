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
