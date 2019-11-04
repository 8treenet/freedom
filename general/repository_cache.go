package general

import (
	"github.com/go-redis/redis"
)

// RepositoryCache .
type RepositoryCache struct {
	Runtime Runtime
}

// BeginRequest .
func (repo *RepositoryCache) BeginRequest(rt Runtime) {
	repo.Runtime = rt
}

// Client .
func (repo *RepositoryCache) Client() *redis.Client {
	if globalApp.Redis.client == nil {
		panic("Redis not installed")
	}
	return globalApp.Redis.client
}
