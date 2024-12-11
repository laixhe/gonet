package gonet

import (
	"github.com/laixhe/gonet/protocol/gen/config/cgorm"
	"github.com/laixhe/gonet/protocol/gen/config/cmongodb"
	"github.com/laixhe/gonet/protocol/gen/config/credis"
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

func InitRedis(c *credis.Redis) error {
	redis, err := xredis.Init(c)
	if err != nil {
		return err
	}
	cc.redis = redis
	return nil
}

func Redis() *xredis.RedisClient {
	return cc.redis
}

func InitMongo(c *cmongodb.MongoDB) error {
	mongo, err := xmongo.Init(c)
	if err != nil {
		return err
	}
	cc.mongo = mongo
	return nil
}

func Mongo() *xmongo.MongoClient {
	return cc.mongo
}

func InitGorm(c *cgorm.Gorm) error {
	gorm, err := xgorm.Init(c)
	if err != nil {
		return err
	}
	cc.gorm = gorm
	return nil
}

func Gorm() *xgorm.GormClient {
	return cc.gorm
}

func InitI18n(httpHeaderKey string) {
	cc.i18n = xi18n.NewI18n(httpHeaderKey)
}

func I18n() *xi18n.I18n {
	return cc.i18n
}
