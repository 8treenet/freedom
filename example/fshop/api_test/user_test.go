package api_test

import (
	"testing"

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

// 获取用户信息
func TestGetBy(t *testing.T) {
	str, err := requests.NewH2CRequest("http://127.0.0.1:8000/user/1").Get().ToString()
	t.Log(str, err)
}

// 创建用户
func TestPostBy(t *testing.T) {
	var PostUserReq struct {
		Name     string
		Password string
	}

	var GetUserRes struct {
		Code  int
		Error string
		Data  struct {
			Id    int
			Name  string
			Money int
		}
	}

	PostUserReq.Name = "freedom"
	PostUserReq.Password = "123321"
	err := requests.NewH2CRequest("http://127.0.0.1:8000/user").Post().SetJSONBody(PostUserReq).ToJSON(&GetUserRes)
	t.Log(GetUserRes, err)
}

// 修改密码
func TestPut(t *testing.T) {
	var PostUserReq struct {
		ID          int
		NewPassword string
		OldPassword string
	}

	var Resp struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}

	PostUserReq.ID = 3
	PostUserReq.NewPassword = "4561232"
	PostUserReq.OldPassword = "123321"
	err := requests.NewH2CRequest("http://127.0.0.1:8000/user").Put().SetJSONBody(PostUserReq).ToJSON(&Resp)
	t.Log(Resp, err)
}
