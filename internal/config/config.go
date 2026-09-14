package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Postgres struct {
		DataSource string
	}
	Cache cache.CacheConf
	Kafka struct {
		Brokers []string
		Topic   string
	}
	PaymentRpc zrpc.RpcClientConf `json:",optional"`
}