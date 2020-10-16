# freedom
### Message Event
---

## Overview
- kafka
- infrastructure components
- consumer
- producer
- proxy
###### The example in this article uses the kafka message queue, which can implement infra components by itself, such as orchestration services, timing tasks, rockets, etc.
###### Due to limited space, only 2 controllers are shown here. To use DDD style domain events, please use DomainEvent in aggregates or entities.

#### producer
```go
// Shop Controller
type ShopController struct {
    Worker freedom.Worker
    /*
        github.com/8treenet/freedom/infra/kafka
        type Producer interface {
            NewMsg(topic string, content []byte, producerName ...string) *Msg
        }
    */
    Producer kafka.Producer    // Using Producer
}

/* 
    Get handles the GET: /shop/:id route.
*/
func (s *ShopController) GetBy(id int) string {
    // Product Msg
    data, _ := json.Marshal(objects.Goods{
        ID:     id,
        Amount: 10,
    })
    msg := s.Producer.NewMsg(EventSell, data)
    msg.SetHeaders(map[string]string{"x-action": "Buy"}).SetWorker(s.Worker).Publish()
    return "ok"
}  
```


#### consumer
```go
func init() {
    freedom.Prepare(func(initiator freedom.Initiator) {
        store := &StoreController{}
        initiator.BindController("/store", store)

        // Binding Shop Event
        initiator.ListenEvent(EventSell, "StoreController.PostSellGoodsBy")
    })
}

type StoreController struct {
    Worker freedom.Worker
}

/*
    The event method starts with Post, and the parameter is the event id.
*/
func (s *StoreController) PostSellGoodsBy(eventID string) error {
    var goods objects.Goods
    s.Worker.Ctx().ReadJSON(&goods)
    return nil
}
```

#### Config File
```sh
# conf/infra/kafka.toml
[consumer]
# Whether to open kafaka
open = true
# Proxy address, the address of this service
proxy_addr = "http://127.0.0.1:8000"
# Proxy protocol, whether HTTP2
proxy_http2 = false

[producer]
# Whether to open kafaka
open = true

# Support multiple instances
[[producer_clients]]
servers = [":9092"]
# Multi-instance production message selection, default array 0th instance.
# name = "North China II"
# msg := s.Producer.NewMsg(EventSell, data).SelectClient("North China II")

[[consumer_clients]]
servers = [":9092"]
group_id = "freedom"
```


#### Domain Event & Message Middleware
```go
/*
   1) First install the infrastructure of domain events. Freedom has implemented kafka. You can define others by yourself, but domain events only support unique installations.
   2) After installation, the entity can be triggered directly using the DomanEvent function.
*/
func main() {
	// Obtain and install the kafka infrastructure for domain events
    app.InstallDomainEventInfra(kafka.GetDomainEventInfra())
    // Install the message middleware.
    kafka.InstallMiddleware(newProducerMiddleware())
    app.Run(addrRunner, *conf.Get().App)
}

// Customize a message log middleware.
func newProducerMiddleware() kafka.ProducerHandler {
    return func(msg *kafka.Msg) {
        now := time.Now()
        msg.Next()
        diff := time.Now().Sub(now)
        
        if err := msg.GetExecution(); err != nil {
            freedom.Logger().Error(string(msg.Content), freedom.LogFields{
                "topic":    msg.GetTopic(),
                "duration": diff.Milliseconds(),
                "error":    err.Error(),
                "title":    "producer",
            })
            return
        }
        freedom.Logger().Info(string(msg.Content), freedom.LogFields{
            "topic":    msg.GetTopic(),
            "duration": diff.Milliseconds(),
            "title":    "producer",
        })
    }
}
```


```go
// For more reference to the entity tutorial, this code only introduces domain events
package entity

import (
    "strconv"
    "github.com/8treenet/freedom"
    "github.com/8treenet/freedom/example/event-example/domain/dto"
)

type Goods struct {
    freedom.Entity  // Inherited entity interface
    goodsObj dto.Goods 
}

func (g *Goods) Shopping() {
    /*
        Related shoping logic. . .
    */

    // Trigger domain event, `Goods:Shopping`
    g.DomainEvent("Goods:Shopping", g.goodsObj)
}

func (g *Goods) Identity() string {
    return strconv.Itoa(g.goodsObj.ID)
}
```