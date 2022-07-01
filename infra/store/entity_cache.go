package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/8treenet/freedom"
	"github.com/go-redis/redis"
)

// EntityCache The entity cache component.
// The first and second level caches are implemented.
// The first-level cache uses the requested memory.
// Can prevent breakdown.
type EntityCache interface {
	//Gets the entity.
	GetEntity(freedom.Entity) error
	//Delete the entity.
	Delete(result freedom.Entity, async ...bool) error
	//Set up the data source.
	SetSource(func(freedom.Entity) error) EntityCache
	//Set the prefix for cache KEY.
	SetPrefix(string) EntityCache
	//Set the time of life, The default is 5 minutes.
	SetExpiration(time.Duration) EntityCache
	// Turn asynchronous writes on or off
	// The default is to close.
	// Cache misses read the database.
	SetAsyncWrite(bool) EntityCache
	//Turn off redis and only request memory takes effect.
	CloseRedis() EntityCache
}

var _ EntityCache = (*EntityCacheImpl)(nil)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(false, func() *EntityCacheImpl {
			return &EntityCacheImpl{}
		})
	})
}

// EntityCacheImpl .
type EntityCacheImpl struct {
	freedom.Infra
	asyncWrite bool
	prefix     string
	expiration time.Duration
	call       func(result freedom.Entity) error
	client     redis.Cmdable
}

// BeginRequest Polymorphic method, subclasses can override overrides overrides.
// The request is triggered after entry.
func (cache *EntityCacheImpl) BeginRequest(worker freedom.Worker) {
	cache.expiration = 5 * time.Minute
	cache.asyncWrite = false
	cache.Infra.BeginRequest(worker)
	cache.client = cache.Redis()
	cache.prefix = "ec:"
}

// GetEntity Gets the entity.
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

// Delete the entity.
func (cache *EntityCacheImpl) Delete(result freedom.Entity, async ...bool) error {
	name := cache.getName(reflect.ValueOf(result).Type()) + ":" + result.Identity()
	if !cache.Worker().IsDeferRecycle() {
		cache.Worker().Store().Remove(name)
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

// SetSource Set up the data source.
func (cache *EntityCacheImpl) SetSource(call func(result freedom.Entity) error) EntityCache {
	cache.call = call
	return cache
}

// SetAsyncWrite .
// Turn asynchronous writes on or off
// The default is to close.
// Cache misses read the database.
func (cache *EntityCacheImpl) SetAsyncWrite(open bool) EntityCache {
	cache.asyncWrite = open
	return cache
}

// SetPrefix Set the prefix for cache KEY.
func (cache *EntityCacheImpl) SetPrefix(prefix string) EntityCache {
	cache.prefix = prefix
	return cache
}

// SetExpiration Set the time of life, The default is 5 minutes.
func (cache *EntityCacheImpl) SetExpiration(expiration time.Duration) EntityCache {
	cache.expiration = expiration
	return cache
}

// CloseRedis Turn off redis and only request memory takes effect.
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
	if cache.Worker().IsDeferRecycle() {
		return false, nil
	}
	entityStore := cache.Worker().Store().Get(name)
	if entityStore == nil {
		return false, nil
	}
	if err := json.Unmarshal(entityStore.([]byte), result); err != nil {
		return false, err
	}
	return true, nil
}

func (cache *EntityCacheImpl) setStore(name string, stroe []byte) {
	if cache.Worker().IsDeferRecycle() {
		return
	}
	cache.Worker().Store().Set(name, stroe)
}

func (cache *EntityCacheImpl) getRedis(name string) ([]byte, error) {
	if cache.client == nil {
		return nil, redis.Nil
	}
	client := cache.client
	return client.Get(name).Bytes()
}

func (cache *EntityCacheImpl) getCall(name string, result freedom.Entity) ([]byte, error) {
	if cache.call == nil {
		return nil, errors.New("Undefined source")
	}
	err := cache.call(result)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
