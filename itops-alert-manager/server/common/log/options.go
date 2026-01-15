package log

import "go.uber.org/zap/zapcore"

type Option func(options *Options)

type Level zapcore.Level

type Options struct {
	Name        string
	FilePath    string
	Level       string
	MaxSize     int
	MaxBackups  int
	MaxAge      int
	Compress    bool
	AddCaller   bool
	Development bool
}
