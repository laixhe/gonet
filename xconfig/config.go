package xconfig

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// splitConfigFile 通过文件路径获取目录、文件名、扩展名
func splitConfigFile(configFile string) (dir string, fileName string, extName string, err error) {
	if len(configFile) == 0 {
		err = errors.New(configFile + " is empty")
		return
	}
	configFiles := strings.Split(configFile, "/")
	lens := len(configFiles) - 1
	if lens == 0 {
		dir = "."
	} else {
		dir = strings.Join(configFiles[:lens], "/")
	}
	files := strings.Split(configFiles[lens], ".")
	if len(files) <= 1 {
		err = errors.New(configFile + " file name is empty")
		return
	}
	fileName = files[0]
	extName = files[1]
	return
}

// initViper 初始化配置文件
// configFile 配置文件
// isEnv      是否获取环境变量环境
// loadData   装载的数据结构(指针类型)
func initViper(configFile string, isEnv bool, loadData interface{}) error {
	dir, fileName, extName, err := splitConfigFile(configFile)
	if err != nil {
		return err
	}

	v := viper.New()
	// 设置配置文件的名字
	v.SetConfigName(fileName)
	// 添加配置文件所在的路径
	v.AddConfigPath(dir)
	// 设置配置文件类型
	v.SetConfigType(extName)
	if err = v.ReadInConfig(); err != nil {
		return err
	}

	// 优先替换环境变量
	if isEnv {
		envConfigs := ListEnvConfig()
		for k := range envConfigs {
			envConfig := envConfigs[k]
			env := os.Getenv(envConfig.env)
			if env != "" {
				v.Set(envConfig.configKey, env) // 替换
			}
		}
	}

	if err = v.Unmarshal(loadData); err != nil {
		return err
	}
	return nil
}

// Init 初始化配置
// configFile 配置文件
// isEnv      是否获取环境变量环境
// loadData   装载的数据结构(指针类型)
func Init(configFile string, isEnv bool, loadData interface{}) {
	if _, err := os.Stat(configFile); err != nil {
		panic("配置文件不存在或其他错误：" + err.Error())
	}
	if err := initViper(configFile, isEnv, loadData); err != nil {
		panic(err)
	}
}
