package redisx

import (
	"context"
	"errors"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/laixhe/gonet/logx"
	"github.com/laixhe/gonet/proto/gen/config/credis"
)

// Redisx 客户端
type Redisx struct {
	client redis.Cmdable
}

// Ping 判断服务是否可用
func (r *Redisx) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return err
	}
	return nil
}

// Client get redis client
func (r *Redisx) Client() redis.Cmdable {
	return r.client
}

var db *Redisx

// DB get redisx
func DB() *Redisx {
	return db
}

// Client get redis client
func Client() redis.Cmdable {
	return db.client
}

// Ping 判断服务是否可用
func Ping() error {
	return db.Ping()
}

// initSingle 单机
func initSingle(c *credis.Redis) redis.Cmdable {
	options := &redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       int(c.DbNum),
	}
	if c.PoolSize > 0 {
		options.PoolSize = int(c.PoolSize)
	}
	if c.MinIdleConn > 0 {
		options.MinIdleConns = int(c.MinIdleConn)
	}
	return redis.NewClient(options)
}

// initSentinel 哨兵主从
//func initSentinel(c *credis.Redis) redis.Cmdable {
//	options := &redis.FailoverOptions{
//		MasterName:    c.Master,
//		SentinelAddrs: strings.Split(c.Addr, ","),
//		DB:            int(c.DbNum),
//		Password:      c.Password,
//	}
//	if c.PoolSize > 0 {
//		options.PoolSize = int(c.PoolSize)
//	}
//	if c.MinIdleConn > 0 {
//		options.MinIdleConns = int(c.MinIdleConn)
//	}
//	return redis.NewFailoverClient(options)
//}

// initCluster 分布式集群
func initCluster(c *credis.Redis) redis.Cmdable {
	options := &redis.ClusterOptions{
		Addrs:    strings.Split(c.Addr, ","),
		Password: c.Password,
	}
	if c.PoolSize > 0 {
		options.PoolSize = int(c.PoolSize)
	}
	if c.MinIdleConn > 0 {
		options.MinIdleConns = int(c.MinIdleConn)
	}
	return redis.NewClusterClient(options)
}

// Connect 连接数据库
func Connect(c *credis.Redis) (*Redisx, error) {
	r := &Redisx{}

	addrs := strings.Split(c.Addr, ",")
	if len(addrs) == 1 {
		r.client = initSingle(c) // 单机
	} else {
		r.client = initCluster(c) // 分布式集群
	}
	err := r.Ping()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func Init(c *credis.Redis) {
	if c == nil {
		panic(errors.New("redis config is nil"))
	}
	if c.Addr == "" {
		panic(errors.New("redis config addr is nil"))
	}
	logx.Debugf("redis config=%v", c)
	logx.Debug("redis init...")

	var err error
	db, err = Connect(c)
	if err != nil {
		panic(err)
	}

	logx.Debug("redis init ok...")
}
