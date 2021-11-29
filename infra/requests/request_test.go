package requests_test

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/8treenet/freedom/infra/requests"
)

func newRequest(url string) requests.Request {
	//return requests.NewH2CRequest(url)
	return requests.NewHTTPRequest(url)
}

func NewMiddleware() requests.Handler {
	return func(middle requests.Middleware) {
		fmt.Println("begin")
		middle.Next()
		fmt.Println("end")
	}
}
func NewEnableTraceMiddleware() requests.Handler {
	return func(middle requests.Middleware) {
		fmt.Println("begin EnableTrace")
		middle.EnableTraceFromMiddleware()
		middle.Next()
		fmt.Println("end EnableTrace", middle.GetRespone().TraceInfo())
	}
}

func TestGet(t *testing.T) {
	//添加中间件
	requests.InstallMiddleware(NewMiddleware())
	requests.InstallMiddleware(NewEnableTraceMiddleware())

	value, rep := newRequest("http://127.0.0.1:8000/hello").Get().ToString()
	t.Log(value, rep.Error)

	var data interface{}
	rep = newRequest("http://127.0.0.1:8000").Get().ToJSON(&data)
	t.Log(data, rep.Error)

	bytes, rep := newRequest("http://127.0.0.1:8000").Get().ToBytes()
	t.Log(string(bytes), rep.Error)
}

func TestQueryParamGet(t *testing.T) {
	var repData struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
		Data  struct {
			Name  string
			Token string
			ID    int64
			IP    []int64
		} `json:"data"`
	}

	rep := newRequest("http://127.0.0.1:8000/user/8treenet").Get().SetQueryParam("token", "ALFAJSJD13").SetQueryParam("id", 123444).ToJSON(&repData)
	t.Log(repData, rep.Error)

	rep = newRequest("http://127.0.0.1:8000/user/8treenet").Get().SetQueryParams(map[string]interface{}{"token": "ALFAJSJD13", "id": 123321}).ToJSON(&repData)
	t.Log(repData, rep.Error)

	ip := []int{192, 168, 1, 1} //支持slice
	resultString, rep := newRequest("http://127.0.0.1:8000/user/8treenet").Get().SetQueryParam("ip", ip).SetQueryParam("token", "ALFAJSJD13").SetQueryParam("id", 123444).ToString()
	t.Log(resultString, rep.Error)

	param := map[string]interface{}{
		"token": "ALFAJSJD13",
		"id":    123321,
		"ip":    []int{225, 231, 11, 110}, //支持slice
	}
	rep = newRequest("http://127.0.0.1:8000/user/8treenet").Get().SetQueryParams(param).ToJSON(&repData)
	t.Log(repData, rep.Error)
}

func TestPost(t *testing.T) {
	requests.InstallMiddleware(NewMiddleware())

	var postBody struct {
		UserName     string `json:"userName"`
		UserPassword string `json:"userPassword"`
	}
	postBody.UserName = "8treenet"
	postBody.UserPassword = "LALSDL13A15PfDAWE"

	var repData struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
		Data  string `json:"data"`
	}

	rep := newRequest("http://127.0.0.1:8000/hello").Post().SetJSONBody(postBody).ToJSON(&repData)
	t.Log(repData, rep.Error)

	rep = newRequest("http://127.0.0.1:8000/hello").Put().ToJSON(&repData)
	t.Log(repData, rep.Error)
}

func TestFeature(t *testing.T) {
	requests.InstallMiddleware(NewMiddleware())
	resultString, rep := newRequest("http://127.0.0.1:8000/hello").Get().EnableTrace().ToString()
	t.Log(resultString, rep.Error, rep.TraceInfo())

	req := newRequest("http://127.0.0.1:8000/hello").Get()
	ctx, cancel := context.WithTimeout(req.Context(), time.Millisecond*200)
	defer cancel()
	resultString, rep = req.WithContext(ctx).ToString()
	t.Log(resultString, rep.Error)
}

func TestSingleflight(t *testing.T) {
	requests.InstallMiddleware(NewMiddleware())
	requests.InstallMiddleware(NewEnableTraceMiddleware())

	group := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		group.Add(1)
		go func() {
			resultString, rep := newRequest("http://127.0.0.1:8000/hello").Get().Singleflight("xxxxxxx").ToString()
			t.Log(resultString, rep.Error)
			group.Done()
		}()
	}
	group.Wait()
}

func TestForm(t *testing.T) {
	req := newRequest("http://localhost:8000/form").Post()
	value := map[string]interface{}{}
	data := url.Values{}

	data.Set("userName", "8treenet")
	data.Set("mail", "4932004@qq.com")
	data.Add("myData", "123")
	data.Add("myData", "456")
	req = req.SetFormBody(data)
	haha := req.ToJSON(&value)
	t.Log(haha, value)
}
