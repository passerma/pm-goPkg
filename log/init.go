package log

import (
	"os"
	"strings"
)

func isDEV() bool {
	// 支持多种环境变量名称
	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("APP_ENV")
	}
	if env == "" {
		env = os.Getenv("env")
	}
	return strings.ToLower(env) == "dev" || strings.ToLower(env) == "development"
}

// Package log 是一个功能齐全的日志库，支持日志轮转和多级日志记录。
//
// 主要特性:
// - 支持日志文件轮转（按时间/大小）
// - 支持不同级别的日志记录（Info, Error, Warn, Debug等）
// - 支持开发/生产环境切换
// - 提供通用日志和访问日志两种类型
// - 自动创建日志目录

// init log 包初始化
func init() {
	initLog()
	initAccessLog()
}
