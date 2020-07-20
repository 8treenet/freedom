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
###### 篇幅有限,只展示了2个控制器。使用DDD风格的领域事件，请在聚合或实体内使用DomainEvent。

#### producer
```go
//购买控制器
type ShopController struct {
    Worker freedom.Worker
    /*
        github.com/8treenet/freedom/infra/kafka
        type Producer interface {
            NewMsg(topic string, content []byte, producerName ...string) *Msg
        }
    */
    Producer kafka.Producer    //使用生产端组件
}

/* 
    Get handles the GET: /shop/:id route.
*/
func (s *ShopController) GetBy(id int) string {
    //生产消息
    data, _ := json.Marshal(objects.Goods{
        ID:     id,
        Amount: 10,
    })
    msg := s.Producer.NewMsg(EventSell, data)
    msg.SetHeaders(map[string]string{"x-action": "购买"}).SetWorker(s.Worker).Publish()
    return "ok"
}  
```


#### consumer
```go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        store := &StoreController{}
        initiator.BindController("/store", store)

        //绑定消费事件
        initiator.ListenEvent(EventSell, "StoreController.PostSellGoodsBy")
    })
}

type StoreController struct {
    Worker freedom.Worker
}

/*
    消费事件必须 Post, 参数是事件唯一id
*/
func (s *StoreController) PostSellGoodsBy(eventID string) error {
    var goods objects.Goods
    s.Worker.Ctx().ReadJSON(&goods)
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
	app.Run(addrRunner, *conf.Get().App)
}
```


```go
//更多参照实体教程，本代码只介绍领域事件
package entity

import (
    "strconv"
    "github.com/8treenet/freedom"
    "github.com/8treenet/freedom/example/event-example/domain/dto"
)

type Goods struct {
    freedom.Entity             //继承实体接口
    goodsObj dto.Goods 
}

func (g *Goods) Shopping() {
    /*
        相关购买逻辑。。。
    */

    //触发领域事件 `Goods:Shopping`
    g.DomainEvent("Goods:Shopping", g.goodsObj)
}

func (g *Goods) Identity() string {
    return strconv.Itoa(g.goodsObj.ID)
}
```