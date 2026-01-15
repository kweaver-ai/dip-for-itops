package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Log = zap.SugaredLogger

type LogCfg struct {
	Filepath    string `yaml:"filepath"`    //日志文件路径
	Level       string `yaml:"level"`       // 日志级别 info warning error
	MaxSize     int    `yaml:"max_size"`    // 每个日志文件最大空间(单位：MB)
	MaxAge      int    `yaml:"max_age"`     // 文件最多保留多少天
	MaxBackups  int    `yaml:"max_backups"` // 文件最多保留多少备份
	Compress    bool   `yaml:"compress"`    //是否压缩
	Development bool   `yaml:"development"` //是否开启开发模式，开启后日志会同步打印到标准输出，同时会打印更详细的堆栈信息
}

var (
	defaultLogFilePath = "/opt/itops-alert-analysis/log/itops-alert-analysis.log"

	Logger *Log
)

func SetDefaultLog(logConf *LogCfg) {
	Logger = NewLogger(logConf)
}

// 初始化日志对象
func NewLogger(logConf *LogCfg) *Log {
	// 日志文件hook
	hook := &lumberjack.Logger{
		Filename:   logConf.Filepath,
		LocalTime:  true,               // 日志文件名的时间格式为本地时间
		MaxAge:     logConf.MaxAge,     // 文件保留的最长时间，单位为天
		MaxBackups: logConf.MaxBackups, // 旧文件保留的最大个数
		MaxSize:    logConf.MaxSize,    // 单个文件最大长度，单位是M
		Compress:   logConf.Compress,   // 是否压缩归档的日志文件
	}

	// 日志格式设定
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "linenum",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	atomicLevel, err := zap.ParseAtomicLevel(logConf.Level)

	if err != nil {
		// 初始化时 Logger 可能还未设置，直接使用默认级别
		atomicLevel = zap.NewAtomicLevel()
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),                                       // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(hook)), // 打印到控制台和文件
		atomicLevel, // 日志级别
	)

	// 设置初始化字段
	filed := zap.Fields(zap.String("serviceName", "itops-alert-analysis"))
	// 构造日志
	var logger *zap.SugaredLogger
	// 跳过一层调用栈，显示实际调用日志的位置而不是 log.go
	callerSkip := zap.AddCallerSkip(1)
	caller := zap.AddCaller()

	if logConf.Development {
		// 开启开发模式，堆栈跟踪
		development := zap.Development()
		logger = zap.New(core, caller, callerSkip, development, filed).Sugar()
	} else {
		logger = zap.New(core, caller, callerSkip, filed).Sugar()
	}

	return logger
}

func init() {
	lf := LogCfg{
		Filepath:    defaultLogFilePath,
		Development: true,
		Level:       "info",
		MaxAge:      100,
		MaxBackups:  20,
		MaxSize:     100,
	}
	Logger = NewLogger(&lf)
}

// 便捷方法 - 直接使用全局 Logger

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	Logger.Debugf(template, args...)
}
func Debugw(msg string, keysAndValues ...interface{}) {
	Logger.Debugw(msg, keysAndValues...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	Logger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	Logger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	Logger.Errorf(template, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	Logger.Fatalf(template, args...)
}

func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}
