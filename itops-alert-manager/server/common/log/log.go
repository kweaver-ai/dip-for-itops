package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Field = zapcore.Field
type LogCfg struct {
	FilePath    string `mapstructure:"filePath"`    //日志文件路径
	Level       string `mapstructure:"logLevel"`    // 日志级别 info warning error
	MaxSize     int    `mapstructure:"maxSize"`     // 每个日志文件最大空间(单位：MB)
	MaxAge      int    `mapstructure:"maxAge"`      // 文件最多保留多少天
	MaxBackups  int    `mapstructure:"maxBackups"`  // 文件最多保留多少备份
	Compress    bool   `mapstructure:"compress"`    //是否压缩
	Development bool   `mapstructure:"development"` //是否开启开发模式，开启后日志会同步打印到标准输出，同时会打印更详细的堆栈信息
}
type Logger interface {
	Debug(msg string, fields ...Field)
	Debugf(msg string, v ...interface{})
	Debugw(msg string, keysAndVals ...interface{})

	Info(msg string, fields ...Field)
	Infof(msg string, v ...interface{})
	Infow(msg string, keysAndVals ...interface{})

	Warn(msg string, fields ...Field)
	Warnf(msg string, v ...interface{})
	Warnw(msg string, keysAndVals ...interface{})

	Error(msg string, fields ...Field)
	Errorf(msg string, v ...interface{})
	Errorw(msg string, keysAndVals ...interface{})
}

type zapLogger struct {
	logger *zap.Logger
}

var defaultLogger Logger

func InitLogger(cfg LogCfg) {
	opts := Options{
		Name:        "default",
		FilePath:    cfg.FilePath,
		Level:       cfg.Level,
		MaxSize:     cfg.MaxSize,
		MaxBackups:  cfg.MaxBackups,
		MaxAge:      cfg.MaxAge,
		Compress:    cfg.Compress,
		Development: cfg.Development,
	}
	defaultLogger = NewLogger(opts)
}

func NewLogger(opts Options) Logger {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "linenum",
		FunctionKey:    "function",
		StacktraceKey:  "tacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	hook := lumberjack.Logger{
		Filename:   opts.FilePath,
		MaxSize:    opts.MaxSize,
		MaxBackups: opts.MaxBackups,
		MaxAge:     opts.MaxAge,
		Compress:   opts.Compress,
	}
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(opts.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}
	atomicLevel := zap.NewAtomicLevelAt(zapLevel)
	writeSyncers := []zapcore.WriteSyncer{zapcore.AddSync(&hook)}
	if opts.Development {
		writeSyncers = append(writeSyncers, zapcore.AddSync(os.Stdout))
	}
	//core :=  zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(writeSyncers...), atomicLevel)
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(writeSyncers...), atomicLevel)
	//l := zap.New(core, zap.Fields(zap.String("servicename", opts.Name)))
	l := zap.New(core)

	if opts.Development {
		// 开启文件及行号
		l = l.WithOptions(zap.Development())
	}
	if opts.AddCaller {
		// 开启开发模式，堆栈跟踪
		l = l.WithOptions(zap.AddCaller())
	}
	logger := &zapLogger{logger: l}
	return logger
}

func Debug(msg string, fields ...Field) {
	defaultLogger.Debug(msg, fields...)
}
func Debugf(msg string, v ...interface{}) {
	defaultLogger.Debugf(msg, v...)
}
func Debugw(msg string, keysAndVals ...interface{}) {
	defaultLogger.Debugw(msg, keysAndVals...)
}

func Info(msg string, fields ...Field) {
	defaultLogger.Info(msg, fields...)
}
func Infof(msg string, v ...interface{}) {
	defaultLogger.Infof(msg, v...)
}
func Infow(msg string, keysAndVals ...interface{}) {
	defaultLogger.Infow(msg, keysAndVals...)
}

func Warn(msg string, fields ...Field) {
	defaultLogger.Warn(msg, fields...)
}
func Warnf(msg string, v ...interface{}) {
	defaultLogger.Warnf(msg, v...)
}
func Warnw(msg string, keysAndVals ...interface{}) {
	defaultLogger.Warnw(msg, keysAndVals...)
}
func Error(msg string, fields ...Field) {
	defaultLogger.Error(msg, fields...)
}
func Errorf(msg string, v ...interface{}) {
	defaultLogger.Errorf(msg, v...)
}
func Errorw(msg string, keysAndVals ...interface{}) {
	defaultLogger.Errorw(msg, keysAndVals...)
}

func (z zapLogger) Debug(msg string, fields ...Field) {
	z.logger.Debug(msg, fields...)
}

func (z zapLogger) Debugf(msg string, v ...interface{}) {
	z.logger.Sugar().Debugf(msg, v...)
}

func (z zapLogger) Debugw(msg string, keysAndVals ...interface{}) {
	z.logger.Sugar().Debugw(msg, keysAndVals...)
}

func (z zapLogger) Info(msg string, fields ...Field) {
	z.logger.Info(msg, fields...)
}

func (z zapLogger) Infof(msg string, v ...interface{}) {
	z.logger.Sugar().Infof(msg, v...)
}

func (z zapLogger) Infow(msg string, keysAndVals ...interface{}) {
	z.logger.Sugar().Infow(msg, keysAndVals...)
	z.logger.Sugar().Info()
}

func (z zapLogger) Warn(msg string, fields ...Field) {
	z.logger.Warn(msg, fields...)
}

func (z zapLogger) Warnf(msg string, v ...interface{}) {
	z.logger.Sugar().Warnf(msg, v...)
}

func (z zapLogger) Warnw(msg string, keysAndVals ...interface{}) {
	z.logger.Sugar().Warnw(msg, keysAndVals...)
}

func (z zapLogger) Error(msg string, fields ...Field) {
	z.logger.Error(msg, fields...)
}

func (z zapLogger) Errorf(msg string, v ...interface{}) {
	z.logger.Sugar().Errorf(msg, v...)
}

func (z zapLogger) Errorw(msg string, keysAndVals ...interface{}) {
	z.logger.Sugar().Errorw(msg, keysAndVals...)
}
