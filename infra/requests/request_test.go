package requests

import (
	"fmt"
	"sync"
	"testing"
)

func TestMiddlewares(t *testing.T) {
	UseMiddleware(NewTestMiddlewares())
	value, rep := NewHttpRequest("http://127.0.0.1:8000/hello").Get().ToString()
	t.Log(value, rep)
}

func NewTestMiddlewares() Handler {
	return func(middle Middleware) {
		//普罗米修斯等耗时拦截器
		middle.Next()
	}
}

func NewTestStopMiddlewares() Handler {
	return func(middle Middleware) {
		fmt.Println("开始")
		req := middle.GetRequest()
		fmt.Println(req)
		//在这里可以拦截
		middle.Stop()
		fmt.Println("结束")
	}
}

func TestH2cPress(t *testing.T) {
	UseMiddleware(NewTestMiddlewares())
	var wait sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wait.Add(1)
		go func() {
			value, _ := NewH2CRequest("http://127.0.0.1:8000/hello").Get().ToString()
			if value != "hello" {
				panic("fuck")
			}
			wait.Done()
		}()
	}
	wait.Wait()
}

func TestH2cSingleflightPress(t *testing.T) {
	UseMiddleware(NewTestMiddlewares())
	var wait sync.WaitGroup
	for i := 0; i < 20000; i++ {
		wait.Add(1)
		go func() {
			value, _ := NewH2CRequest("http://127.0.0.1:8000/hello").Get().Singleflight("fuck").ToString()
			if value != "hello" {
				panic("fuck")
			}
			wait.Done()
		}()
	}
	wait.Wait()
}

func TestH1cPress(t *testing.T) {
	UseMiddleware(NewTestMiddlewares())
	var wait sync.WaitGroup
	for i := 0; i < 300; i++ {
		wait.Add(1)
		go func() {
			value, _ := NewHttpRequest("http://127.0.0.1:8000/hello").Get().ToString()
			if value != "hello" {
				panic("fuck")
			}
			wait.Done()
		}()
	}
	wait.Wait()
}

func TestH1SingleflightPress(t *testing.T) {
	UseMiddleware(NewTestMiddlewares())
	var wait sync.WaitGroup
	for i := 0; i < 20000; i++ {
		wait.Add(1)
		go func() {
			value, _ := NewHttpRequest("http://127.0.0.1:8000/hello").Get().Singleflight("fuck").ToString()
			if value != "hello" {
				panic("fuck")
			}
			wait.Done()
		}()
	}
	wait.Wait()
}
