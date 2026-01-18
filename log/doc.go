/*
log 是一个功能齐全的日志库，支持日志轮转和多级日志记录。

主要特性:
- 支持日志文件轮转（按时间/大小）
- 支持不同级别的日志记录（Info, Error, Warn, Debug等）
- 支持开发/生产环境切换
- 提供通用日志和访问日志两种类型
- 自动创建日志目录

快速开始:

	package main

	import (
	    "github.com/passerma/pm-log/log"
	)

	func main() {
	    // 使用默认配置，直接开始记录日志
	    log.Info("程序启动")
	    log.Error("发生错误")
	    log.Warnf("警告信息: %s", "测试")
	}

环境变量:

- ENV 或 APP_ENV 或 env - 设置环境，当值为 dev 或 development 时会输出到控制台

高级配置:

	package main

	import (
	    "time"
	    "github.com/passerma/pm-log/log"
	)

	func main() {
	    // 自定义配置
	    config := log.LogConfig{
	        LogDir:        "./custom_logs",
	        MaxAge:        30 * 24 * time.Hour, // 保留30天
	        RotationTime:  24 * time.Hour,      // 每天轮转
	        RotationSize:  10 * 1024 * 1024,   // 10MB
	        EnableConsole: true,                // 同时输出到控制台
	    }

	    log.SetConfig(config)

	    log.Info("使用自定义配置记录日志")
	}
*/
package log
