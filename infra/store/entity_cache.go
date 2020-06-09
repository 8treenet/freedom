package store

/*
	缓存组件，实现了一级缓存，二级缓存，防击穿.
*/
import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/8treenet/freedom"
	"github.com/go-redis/redis"
	"golang.org/x/sync/singleflight"
)

type EntityCache interface {
	//获取实体
	GetEntity(freedom.Entity) error
	//删除实体缓存
	Delete(result freedom.Entity, async ...bool) error
	//设置数据源
	SetSource(func(freedom.Entity) error) EntityCache
	//设置前缀
	SetPrefix(string) EntityCache
	//设置缓存时间，默认5分钟
	SetExpiration(time.Duration) EntityCache
	//设置异步反写缓存。默认关闭，缓存未命中读取数据源后的异步反写缓存
	SetAsyncWrite(bool) EntityCache
	//设置防击穿，默认开启
	SetSingleFlight(bool) EntityCache
	//关闭二级缓存. 关闭后只有一级缓存生效
	CloseRedis() EntityCache
}

var _ EntityCache = new(EntityCacheImpl)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(false, func() *EntityCacheImpl {
			return &EntityCacheImpl{}
		})
	})
}

var group singleflight.Group

// EntityCacheImpl .
type EntityCacheImpl struct {
	freedom.Infra
	asyncWrite   bool
	prefix       string
	expiration   time.Duration
	call         func(result freedom.Entity) error
	singleFlight bool
	client       redis.Cmdable
}

// BeginRequest
func (cache *EntityCacheImpl) BeginRequest(worker freedom.Worker) {
	cache.expiration = 5 * time.Minute
	cache.singleFlight = true
	cache.asyncWrite = false
	cache.Infra.BeginRequest(worker)
	cache.client = cache.Redis()
}

// Get 读取实体缓存
func (cache *EntityCacheImpl) GetEntity(result freedom.Entity) error {
	value := reflect.ValueOf(result)
	//一级缓存读取
	name := cache.getName(value.Type()) + ":" + result.Identity()
	ok, err := cache.getStore(name, result)
	if err != nil || ok {
		return err
	}

	//二级缓存读取
	entityBytes, err := cache.getRedis(name)
	if err != nil && err != redis.Nil {
		return err
	}
	if err != redis.Nil {
		err = json.Unmarshal(entityBytes, result)
		if err != nil {
			return err
		}
		cache.setStore(name, entityBytes)
		return nil
	}

	//持久化数据源读取
	entityBytes, err = cache.getCall(name, result)
	if err != nil {
		return err
	}

	//反写缓存
	cache.setStore(name, entityBytes)
	expiration := cache.expiration
	client := cache.client
	if client == nil {
		return nil
	}
	if !cache.asyncWrite {
		return client.Set(name, entityBytes, expiration).Err()
	}
	go func() {
		var err error
		defer func() {
			if perr := recover(); perr != nil {
				err = fmt.Errorf(fmt.Sprint(perr))
			}
			if err != nil {
				freedom.Logger().Errorf("Failed to set entity cache, name:%s err:%v, ", name, err)
			}
		}()
		err = client.Set(name, entityBytes, expiration).Err()
	}()

	return nil
}

// Delete 删除实体缓存
func (cache *EntityCacheImpl) Delete(result freedom.Entity, async ...bool) error {
	name := cache.getName(reflect.ValueOf(result).Type()) + ":" + result.Identity()
	if !cache.Worker.IsDeferRecycle() {
		cache.Worker.Store().Remove(name)
	}
	client := cache.client
	if client == nil {
		return nil
	}
	if len(async) == 0 {
		return client.Del(name).Err()
	}
	go func() {
		var err error
		defer func() {
			if perr := recover(); perr != nil {
				err = fmt.Errorf(fmt.Sprint(perr))
			}
			if err != nil {
				freedom.Logger().Errorf("Failed to delete entity cache, name:%s err:%v, ", name, err)
			}
		}()
		err = client.Del(name).Err()
	}()
	return nil
}

// SetSource 设置数据源
func (cache *EntityCacheImpl) SetSource(call func(result freedom.Entity) error) EntityCache {
	cache.call = call
	return cache
}

// SetAsyncWrite 设置异步写入,默认同步写入缓存。 当缓存未命中读取数据源后是否异步写入缓存
func (cache *EntityCacheImpl) SetAsyncWrite(open bool) EntityCache {
	cache.asyncWrite = open
	return cache
}

// SetPrefix 设置缓存实体前缀
func (cache *EntityCacheImpl) SetPrefix(prefix string) EntityCache {
	cache.prefix = prefix
	return cache
}

// SetExpiration 设置缓存实体时间 默认5分钟
func (cache *EntityCacheImpl) SetExpiration(expiration time.Duration) EntityCache {
	cache.expiration = expiration
	return cache
}

// SetSingleFlight 默认开启
func (cache *EntityCacheImpl) SetSingleFlight(open bool) EntityCache {
	cache.singleFlight = open
	return cache
}

// CloseRedis 关闭二级缓存
func (cache *EntityCacheImpl) CloseRedis() EntityCache {
	cache.client = nil
	return cache
}

func (cache *EntityCacheImpl) getName(entityType reflect.Type) string {
	for entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}
	if cache.prefix != "" {
		return cache.prefix + ":" + entityType.Name()
	}
	return entityType.Name()
}

func (cache *EntityCacheImpl) getStore(name string, result freedom.Entity) (bool, error) {
	if cache.Worker.IsDeferRecycle() {
		return false, nil
	}
	entityStore := cache.Worker.Store().Get(name)
	if entityStore == nil {
		return false, nil
	}
	if err := json.Unmarshal(entityStore.([]byte), result); err != nil {
		return false, err
	}
	return true, nil
}

func (cache *EntityCacheImpl) setStore(name string, stroe []byte) {
	if cache.Worker.IsDeferRecycle() {
		return
	}
	cache.Worker.Store().Set(name, stroe)
}

func (cache *EntityCacheImpl) getRedis(name string) ([]byte, error) {
	if cache.client == nil {
		return nil, redis.Nil
	}
	client := cache.client
	if cache.singleFlight {
		entityData, err, _ := group.Do("cache:"+name, func() (interface{}, error) {
			return client.Get(name).Bytes()
		})
		if err != nil {
			return nil, err
		}
		return entityData.([]byte), err
	}
	return client.Get(name).Bytes()
}

func (cache *EntityCacheImpl) getCall(name string, result freedom.Entity) ([]byte, error) {
	if cache.call == nil {
		return nil, errors.New("Undefined source")
	}
	if cache.singleFlight {
		entityData, err, shared := group.Do("call:"+name, func() (interface{}, error) {
			e := cache.call(result)
			if e != nil {
				return nil, e
			}
			return json.Marshal(result)
		})
		if err != nil {
			return nil, err
		}
		if shared {
			err = json.Unmarshal(entityData.([]byte), result)
		}
		return entityData.([]byte), err
	}
	err := cache.call(result)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
