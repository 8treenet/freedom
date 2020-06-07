package case_test

import (
	"fmt"
	"testing"

	"github.com/8treenet/freedom"
)

// 动态代理对象
type ProxyObj struct {
	freedom.ProxyHandle
}

func (proxy *ProxyObj) Run(step int) string {
	result, _ := proxy.Call("Run", step+2) //加2
	return result[0].(string)
}

// 汽车对象
type Car struct {
}

func (proxy *Car) Run(step int) string {
	return fmt.Sprintf("汽车加速 %d", step)
}

func TestAOP(t *testing.T) {
	proxy := new(ProxyObj)
	freedom.NewProxy(proxy, new(Car))

	//使用动态代理对象
	t.Log(proxy.Run(5))
	t.Log(proxy.Run(10))
}
