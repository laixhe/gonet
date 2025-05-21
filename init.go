package gonet

import (
	"github.com/laixhe/gonet/protocol/gen/config/calibaba"
	"github.com/laixhe/gonet/protocol/gen/config/cgorm"
	"github.com/laixhe/gonet/protocol/gen/config/cmongodb"
	"github.com/laixhe/gonet/protocol/gen/config/credis"
	"github.com/laixhe/gonet/protocol/gen/config/cwechat"
	"github.com/laixhe/gonet/sdk/sdkaliyun/imagesearch"
	"github.com/laixhe/gonet/sdk/sdkaliyun/oss"
	"github.com/laixhe/gonet/sdk/sdkwechat/mini"
	"github.com/laixhe/gonet/sdk/sdkwechat/offiaccount"
	"github.com/laixhe/gonet/sdk/sdkwechat/openplatform"
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
	weChatOpenProgram map[string]*openplatform.SdkWeChatOpenProgram
	weChatOffiaccount map[string]*offiaccount.SdkWeChatOffiaccount
	aliyunOss         map[string]*oss.SdkAliyunOss
	aliyunImageSearch map[string]*imagesearch.SdkAliyunImageSearch
}

var xgonet *GoNet

func init() {
	xgonet = &GoNet{
		redis:             make(map[string]*xredis.RedisClient),                // Redis 客户端
		gorm:              make(map[string]*xgorm.GormClient),                  // Gorm 客户端
		mongo:             make(map[string]*xmongo.MongoClient),                // Mongo 客户端
		weChatMiniProgram: make(map[string]*mini.SdkWeChatMiniProgram),         // 微信小程序客户端
		weChatOpenProgram: make(map[string]*openplatform.SdkWeChatOpenProgram), // 微信开放平台客户端
		weChatOffiaccount: make(map[string]*offiaccount.SdkWeChatOffiaccount),  // 微信公众号客户端
		aliyunOss:         make(map[string]*oss.SdkAliyunOss),                  // 阿里云对象存储
		aliyunImageSearch: make(map[string]*imagesearch.SdkAliyunImageSearch),  // 阿里云图像搜索
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

func InitWeChatOpenProgram(config *cwechat.OpenProgram, isDebug bool, key ...string) error {
	weChatOpenProgram, err := openplatform.Init(config, isDebug)
	if err != nil {
		return err
	}
	if len(key) > 0 {
		xgonet.weChatOpenProgram[key[0]] = weChatOpenProgram
	} else {
		xgonet.weChatOpenProgram[DEFAULT] = weChatOpenProgram
	}
	return nil
}

func WeChatOpenProgram(key ...string) *openplatform.SdkWeChatOpenProgram {
	if len(key) > 0 {
		return xgonet.weChatOpenProgram[key[0]]
	} else {
		return xgonet.weChatOpenProgram[DEFAULT]
	}
}

func InitWeChatOffiaccount(config *cwechat.Offiaccount, isDebug bool, key ...string) error {
	weChatOffiaccount, err := offiaccount.Init(config, isDebug)
	if err != nil {
		return err
	}
	if len(key) > 0 {
		xgonet.weChatOffiaccount[key[0]] = weChatOffiaccount
	} else {
		xgonet.weChatOffiaccount[DEFAULT] = weChatOffiaccount
	}
	return nil
}

func WeChatOffiaccount(key ...string) *offiaccount.SdkWeChatOffiaccount {
	if len(key) > 0 {
		return xgonet.weChatOffiaccount[key[0]]
	} else {
		return xgonet.weChatOffiaccount[DEFAULT]
	}
}

// aliyunOss         map[string]*oss.SdkAliyunOss
// aliyunImageSearch map[string]*imagesearch.SdkAliyunImageSearch
func InitAliyunOss(config *calibaba.Oss, key ...string) error {
	aliyunOss, err := oss.Init(config)
	if err != nil {
		return err
	}
	if len(key) > 0 {
		xgonet.aliyunOss[key[0]] = aliyunOss
	} else {
		xgonet.aliyunOss[DEFAULT] = aliyunOss
	}
	return nil
}

func AliyunOss(key ...string) *oss.SdkAliyunOss {
	if len(key) > 0 {
		return xgonet.aliyunOss[key[0]]
	} else {
		return xgonet.aliyunOss[DEFAULT]
	}
}
func InitAliyunImageSearch(config *calibaba.ImageSearch, key ...string) error {
	aliyunImageSearch, err := imagesearch.Init(config)
	if err != nil {
		return err
	}
	if len(key) > 0 {
		xgonet.aliyunImageSearch[key[0]] = aliyunImageSearch
	} else {
		xgonet.aliyunImageSearch[DEFAULT] = aliyunImageSearch
	}
	return nil
}

func AliyunImageSearch(key ...string) *imagesearch.SdkAliyunImageSearch {
	if len(key) > 0 {
		return xgonet.aliyunImageSearch[key[0]]
	} else {
		return xgonet.aliyunImageSearch[DEFAULT]
	}
}
