package timer

import (
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/domain"
)

// FixedTime
func FixedTime(app freedom.Application) {
	/*
		展示非控制器使用领域服务的示例
		接口 Application.CallService(fun interface{}, worker ...freedom.Worker)
		CallService可以自定义Wokrer
	*/
	go func() {
		time.Sleep(1 * time.Second) //延迟，等待程序Application.Run
		t := time.NewTimer(1 * time.Second)
		for range t.C {
			app.CallService(func(goodsService *domain.Goods) {
				goodsService.AddStock(1, 1)
			})
			t.Reset(6 * time.Second)
		}
	}()
}
