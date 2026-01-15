package config

import (
	"os"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	//配置文件信息
	cfgPath string = "./config/"
	cfgName string = "config"
	cfgType string = "yaml"
	//服务版本路径
	versionPath string = "./VERSION"

	gCfg *GlobalCfg
	vp   *viper.Viper
)

const (
	DBServer                = "db-server"
	UniQueryServiceName     = "uniquery"
	UniQueryServiceNamePort = "default"

	ReleaseMode string = "release"
	DebugMode   string = "debug"
)

type GlobalCfg struct {
	App        AppCfg     `mapstructure:"app"`
	Log        log.LogCfg `mapstructure:"log"`
	Mysql      MysqlCfg
	HttpServer HttpServerCfg `mapstructure:"server"`
	RestAPI    RestAPI
}

// application config
type AppCfg struct {
	Mode    string `mapstructure:"mode"`    // 启动模式 : release，debug
	Version string `mapstructure:"version"` // 应用版本
}

// http server config
type HttpServerCfg struct {
	RunMode      string        `mapstructure:"runMode"`
	Addr         int           `mapstructure:"httpPort"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
}

// db config
type MysqlCfg struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// RestAPI
type RestAPI struct {
	UniQueryDomain string
}

func Get() *GlobalCfg {
	return gCfg
}

// 初始化配置
func InitPremise() {
	vp = viper.New()
	vp.AddConfigPath(cfgPath)
	vp.SetConfigName(cfgName)
	vp.SetConfigType(cfgType)
	loadSetting(vp)
	vp.WatchConfig()
	vp.OnConfigChange(func(e fsnotify.Event) {
		loadSetting(vp)
	})
}
func loadSetting(vp *viper.Viper) {
	if err := vp.ReadInConfig(); err != nil {
		panic(err.Error())
	}
	if err := vp.Unmarshal(&gCfg); err != nil {
		panic(err.Error())
	}
	gCfg.App.Version, _ = parseVersion(versionPath)
	initResetApi()
	setRunMode()
	log.InitLogger(gCfg.Log)
}

func parseVersion(versionPath string) (string, error) {
	b, err := os.ReadFile(versionPath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
func setRunMode() {
	switch gCfg.App.Mode {
	case ReleaseMode:
		gCfg.Log.Development = false
		gCfg.HttpServer.RunMode = gin.ReleaseMode
	default:
		gCfg.Log.Development = true
		gCfg.HttpServer.RunMode = gin.DebugMode
	}
}

func initResetApi() {
	//指标模型url
	gCfg.RestAPI.UniQueryDomain = "http://mdl-uniquery-svc:13011"
	//gCfg.RestAPI.UniQueryDomain = "http://192.168.201.15"
}
