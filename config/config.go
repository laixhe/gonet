package config

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// splitConfigFile 通过文件路径获取目录、文件名、扩展名
// 例如 "config/app.yaml" -> dir="config", fileName="app", extName="yaml"
func splitConfigFile(configFile string) (dir, fileName, extName string) {
	dir = filepath.Dir(configFile)
	base := filepath.Base(configFile)
	ext := filepath.Ext(base)
	extName = strings.TrimPrefix(ext, ".")
	fileName = strings.TrimSuffix(base, ext)
	return
}

// initViper 初始化并加载配置文件到 loadData（必须为指针类型）
func initViper(configFile string, loadData any) error {
	dir, fileName, extName := splitConfigFile(configFile)
	v := viper.New()
	v.SetConfigName(fileName)
	v.AddConfigPath(dir)
	v.SetConfigType(extName)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	return v.Unmarshal(loadData)
}

// Init 加载配置文件到 loadData（必须为指针类型）。
//
// 支持 yaml / json / toml 等 viper 支持的格式。
//
// 使用示例：
//
//	type AppConfig struct {
//	    Server ServerConf `mapstructure:"server"`
//	}
//	var cfg AppConfig
//	if err := config.Init("config/app.yaml", &cfg); err != nil {
//	    log.Fatal(err)
//	}
func Init(configFile string, loadData any) error {
	return initViper(configFile, loadData)
}
