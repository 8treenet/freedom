package config

import "github.com/8treenet/freedom"

func newDBConf() *DBConf {
	result := &DBConf{}
	freedom.Configure(result, "db.toml", false)
	return result
}

// DBCacheConf .
type DBCacheConf struct {
	RedisConf
	Expires int `toml:"expires"`
}

// DBConf .
type DBConf struct {
	Addr            string      `toml:"addr"`
	MaxOpenConns    int         `toml:"max_open_conns"`
	MaxIdleConns    int         `toml:"max_idle_conns"`
	ConnMaxLifeTime int         `toml:"conn_max_life_time"`
	Cache           DBCacheConf `toml:"cache"`
}
