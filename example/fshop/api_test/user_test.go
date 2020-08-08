package api_test

import (
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

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
	PostUserReq.NewPassword = "123321"
	PostUserReq.OldPassword = "123321"
	err := requests.NewH2CRequest("http://127.0.0.1:8000/user").Put().SetJSONBody(PostUserReq).ToJSON(&Resp)
	t.Log(Resp, err)
}
