package gcache

import (
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/gcache"
	cache "github.com/8treenet/gcache"
	"github.com/jinzhu/gorm"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindInfra(true, &GCache{})
	})
}

// gCacheConf .
type gCacheConf struct {
	Addr               string `toml:"addr"`
	Password           string `toml:"password"`
	DB                 int    `toml:"db"`
	MaxRetries         int    `toml:"max_retries"`
	PoolSize           int    `toml:"pool_size"`
	ReadTimeout        int    `toml:"read_timeout"`
	WriteTimeout       int    `toml:"write_timeout"`
	IdleTimeout        int    `toml:"idle_timeout"`
	IdleCheckFrequency int    `toml:"idle_check_frequency"`
	MaxConnAge         int    `toml:"max_conn_age"`
	PoolTimeout        int    `toml:"pool_timeout"`
	Expires            int    `toml:"expires"`
}

// GCache .
type GCache struct {
	cache.Plugin
	db *gorm.DB
}

// Booting .
func (g *GCache) Booting(sb freedom.SingleBoot) {
	var cfg gCacheConf
	freedom.Configure(cfg, "infra/gcache.toml", true)
	opt := cache.DefaultOption{}
	opt.Expires = cfg.Expires      //缓存时间，默认60秒。范围 30-900
	opt.Level = gcache.LevelSearch //缓存级别，默认LevelSearch。LevelDisable:关闭缓存，LevelModel:模型缓存， LevelSearch:查询缓存
	ropt := cache.RedisOption{
		Addr:               cfg.Addr,
		Password:           cfg.Password,
		DB:                 cfg.DB,
		MaxRetries:         cfg.MaxRetries,
		PoolSize:           cfg.PoolSize,
		ReadTimeout:        time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:       time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:        time.Duration(cfg.IdleTimeout) * time.Second,
		IdleCheckFrequency: time.Duration(cfg.IdleCheckFrequency) * time.Second,
		MaxConnAge:         time.Duration(cfg.MaxConnAge) * time.Second,
		PoolTimeout:        time.Duration(cfg.PoolTimeout) * time.Second,
	}

	//缓存中间件 注入到Gorm
	g.Plugin = gcache.AttachDB(sb.DB(), &opt, &ropt)
}
