# freedom
### 消息事件
---

## Overview
- kafka
- 基础设施组件
- consumer
- producer
- 代理机制
###### 本篇的示例使用kafka消息队列，可自行实现infra组件，如编排服务、定时任务、rocket等。
###### 篇幅有限,只展示了2个控制器。使用DDD风格的领域事件，请在聚合内使用producer。

#### producer
```go
import (
	"github.com/8treenet/freedom/infra/kafka"
)

const (
	EventSell = "event-sell"
)

//购买控制器
type ShopController struct {
    Runtime freedom.Runtime
    Producer *kafka.Producer    //使用生产端组件
}

/* 
    Get handles the GET: /shop/:id route.
    Producer.NewMsg(Topic string, Content []byte) : 创建消息
    Msg.SetHeaders(map[string]inteface{}) : 设置k/v 头
    Msg.SetRuntime(Runtime) : 设置请求上下文，设置后请求的上下文会透传到消费端
    msg.Publish() : 发布
*/
func (s *ShopController) GetBy(id int) string {
    data, _ := json.Marshal(objects.Goods{
        ID:     id,
        Amount: 10,
    })
    msg := s.Producer.NewMsg(EventSell, data)
    msg.SetHeaders(map[string]string{"x-action": "购买"}).SetRuntime(s.Runtime).Publish()
    return "ok"
}  
```


#### consumer
```go
import (
	"github.com/8treenet/freedom/infra/kafka"
)

func init() {
    freedom.Booting(func(initiator freedom.Initiator) {
        store := &StoreController{}
        initiator.BindController("/store", store)

        //绑定事件
        initiator.ListenEvent(EventSell, store.PostSellGoodsBy)
    })
}

//库存控制器
type StoreController struct {
    Runtime freedom.Runtime
}

/*
    PostSellGoodsBy 事件方法为 Post开头, 参数是事件唯一id
*/
func (s *StoreController) PostSellGoodsBy(eventID string) error {
    //rawData, err := ioutil.ReadAll(s.Runtime.Ctx().Request().Body)
    var goods objects.Goods
    s.Runtime.Ctx().ReadJSON(&goods)

    action := s.Runtime.Ctx().GetHeader("x-action")
    s.Runtime.Logger().Infof("消耗商品ID:%d, %d件, 行为:%s, 消息key:%s", goods.ID, goods.Amount, action, eventID)
    return nil
}
```

#### 配置文件
```sh
# conf/infra/kafka.toml
[consumer]
# 是否开启
open = true
# 代理地址，本服务的地址
proxy_addr = "http://127.0.0.1:8000"
# 代理协议，是否HTTP2
proxy_http2 = false

[producer]
# 是否开启
open = true

#支持多实例
[[producer_clients]]
servers = [":9092"]
# 多实例的生产消息选择, 默认数组第0实例
# name = "华北二"
# msg := s.Producer.NewMsg(EventSell, data).SelectClient("华北二")

[[consumer_clients]]
servers = [":9092"]
group_id = "freedom"
```


#### 领域事件风格
```go
/*
    1.先要安装领域事件的基础设施，freedom 已经实现了kafka。可自行定义其他，但领域事件只支持唯一的安装。
    2.安装以后 entity 可以直接使用 DomanEvent 方法触发。
*/
func main() {
	//获取领域事件的kafka基础设施并安装
	app.InstallDomainEventInfra(kafka.GetDomainEventInfra())
	app.Run(addrRunner, *config.Get().App)
}
```


```go
//更多参照实体教程，本代码只介绍领域事件
package entity

import (
    "strconv"
    "github.com/8treenet/freedom"
    "github.com/8treenet/freedom/example/event-example/application/objects"
)

type Goods struct {
    freedom.Entity             //继承实体接口
    goodsObj objects.Goods     //商品值对象
}

/*
    DomainEvent(fun interface{}, object interface{}, header ...map[string]string)
    fun : 为触发事件的方法。 基于强类型的原则，框架已经做了方法和字符串Topic的映射,`实体名字:方法名`
    object : 结构体数据,会做json转换
    header : k/v 附加数据
*/
func (g *Goods) Shop() {
    /*
        相关购买逻辑。。。
    */

    //触发领域事件 `Goods:Shop`
    g.DomainEvent(g.Shop, g.goodsObj)
}

func (g *Goods) Identity() string {
    return strconv.Itoa(g.goodsObj.ID)
}
```