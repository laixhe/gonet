package gonet

import (
	"github.com/laixhe/gonet/protocol/gen/config/cgorm"
	"github.com/laixhe/gonet/protocol/gen/config/cmongodb"
	"github.com/laixhe/gonet/protocol/gen/config/credis"
	"github.com/laixhe/gonet/protocol/gen/config/cwechat"
	"github.com/laixhe/gonet/sdk/sdkwechat/mini"
	"github.com/laixhe/gonet/xgorm"
	"github.com/laixhe/gonet/xi18n"
	"github.com/laixhe/gonet/xmongo"
	"github.com/laixhe/gonet/xredis"
)

const DEFAULT = "default"

type GoNet struct {
	i18n              *xi18n.I18n
	redis             map[string]*xredis.RedisClient
	gorm              map[string]*xgorm.GormClient
	mongo             map[string]*xmongo.MongoClient
	weChatMiniProgram map[string]*mini.SdkWeChatMiniProgram
}

var xgonet *GoNet

func init() {
	xgonet = &GoNet{
		redis:             make(map[string]*xredis.RedisClient),
		gorm:              make(map[string]*xgorm.GormClient),
		mongo:             make(map[string]*xmongo.MongoClient),
		weChatMiniProgram: make(map[string]*mini.SdkWeChatMiniProgram),
	}
}

func InitI18n(httpHeaderKey string) {
	xgonet.i18n = xi18n.NewI18n(httpHeaderKey)
}

func I18n() *xi18n.I18n {
	return xgonet.i18n
}

func InitRedis(c *credis.Redis, key ...string) error {
	redis, err := xredis.Init(c)
	if err != nil {
		return err
	}
	if len(key) > 0 {
		xgonet.redis[key[0]] = redis
	} else {
		xgonet.redis[DEFAULT] = redis
	}
	return nil
}

func Redis(key ...string) *xredis.RedisClient {
	if len(key) > 0 {
		return xgonet.redis[key[0]]
	} else {
		return xgonet.redis[DEFAULT]
	}
}

func InitGorm(c *cgorm.Gorm, key ...string) error {
	gorm, err := xgorm.Init(c)
	if err != nil {
		return err
	}
	if len(key) > 0 {
		xgonet.gorm[key[0]] = gorm
	} else {
		xgonet.gorm[DEFAULT] = gorm
	}
	return nil
}

func Gorm(key ...string) *xgorm.GormClient {
	if len(key) > 0 {
		return xgonet.gorm[key[0]]
	} else {
		return xgonet.gorm[DEFAULT]
	}
}

func InitMongo(c *cmongodb.MongoDB, key ...string) error {
	mongo, err := xmongo.Init(c)
	if err != nil {
		return err
	}
	if len(key) > 0 {
		xgonet.mongo[key[0]] = mongo
	} else {
		xgonet.mongo[DEFAULT] = mongo
	}
	return nil
}

func Mongo(key ...string) *xmongo.MongoClient {
	if len(key) > 0 {
		return xgonet.mongo[key[0]]
	} else {
		return xgonet.mongo[DEFAULT]
	}
}

func InitWeChatMiniProgram(config *cwechat.MiniProgram, isDebug bool, key ...string) error {
	weChatMiniProgram, err := mini.Init(config, isDebug)
	if err != nil {
		return err
	}
	if len(key) > 0 {
		xgonet.weChatMiniProgram[key[0]] = weChatMiniProgram
	} else {
		xgonet.weChatMiniProgram[DEFAULT] = weChatMiniProgram
	}
	return nil
}

func WeChatMiniProgram(key ...string) *mini.SdkWeChatMiniProgram {
	if len(key) > 0 {
		return xgonet.weChatMiniProgram[key[0]]
	} else {
		return xgonet.weChatMiniProgram[DEFAULT]
	}
}
