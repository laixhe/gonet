package gonet

import (
	"github.com/laixhe/gonet/proto/gen/config/cgorm"
	"github.com/laixhe/gonet/proto/gen/config/cmongodb"
	"github.com/laixhe/gonet/proto/gen/config/credis"
	"github.com/laixhe/gonet/xgorm"
	"github.com/laixhe/gonet/xi18n"
	"github.com/laixhe/gonet/xmongo"
	"github.com/laixhe/gonet/xredis"
)

type GoNet struct {
	redis *xredis.RedisClient
	mongo *xmongo.MongoClient
	gorm  *xgorm.GormClient
	i18n  *xi18n.I18n
}

var cc *GoNet

func init() {
	cc = &GoNet{}
}

func RedisInit(c *credis.Redis) error {
	redis, err := xredis.Init(c)
	if err != nil {
		return err
	}
	cc.redis = redis
	return nil
}

func RedisClient() *xredis.RedisClient {
	return cc.redis
}

func MongoInit(c *cmongodb.MongoDB) error {
	mongo, err := xmongo.Init(c)
	if err != nil {
		return err
	}
	cc.mongo = mongo
	return nil
}

func MongoClient() *xmongo.MongoClient {
	return cc.mongo
}

func GormInit(c *cgorm.Gorm) error {
	gorm, err := xgorm.Init(c)
	if err != nil {
		return err
	}
	cc.gorm = gorm
	return nil
}

func GormClient() *xgorm.GormClient {
	return cc.gorm
}

func I18nInit(httpHeaderKey string) {
	cc.i18n = xi18n.NewI18n(httpHeaderKey)
}

func I18n() *xi18n.I18n {
	return cc.i18n
}
