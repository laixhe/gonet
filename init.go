package gonet

import (
	"github.com/laixhe/gonet/gormx"
	"github.com/laixhe/gonet/i18nx"
	"github.com/laixhe/gonet/mongox"
	"github.com/laixhe/gonet/proto/gen/config/cgorm"
	"github.com/laixhe/gonet/proto/gen/config/cmongodb"
	"github.com/laixhe/gonet/proto/gen/config/credis"
	"github.com/laixhe/gonet/redisx"
)

type GoNet struct {
	redis *redisx.RedisClient
	mongo *mongox.MongoClient
	gorm  *gormx.GormClient
	i18n  *i18nx.I18n
}

var cc *GoNet

func init() {
	cc = &GoNet{}
}

func RedisInit(c *credis.Redis) error {
	redis, err := redisx.Init(c)
	if err != nil {
		return err
	}
	cc.redis = redis
	return nil
}

func RedisClient() *redisx.RedisClient {
	return cc.redis
}

func MongoInit(c *cmongodb.MongoDB) error {
	mongo, err := mongox.Init(c)
	if err != nil {
		return err
	}
	cc.mongo = mongo
	return nil
}

func MongoClient() *mongox.MongoClient {
	return cc.mongo
}

func GormInit(c *cgorm.Gorm) error {
	gorm, err := gormx.Init(c)
	if err != nil {
		return err
	}
	cc.gorm = gorm
	return nil
}

func GormClient() *gormx.GormClient {
	return cc.gorm
}

func I18nInit(httpHeaderKey string) {
	cc.i18n = i18nx.NewI18n(httpHeaderKey)
}

func I18n() *i18nx.I18n {
	return cc.i18n
}
