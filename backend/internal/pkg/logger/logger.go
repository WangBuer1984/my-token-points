package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// InitLogger 初始化日志
func InitLogger(level string) *logrus.Logger {
	log := logrus.New()

	// 设置输出
	log.SetOutput(os.Stdout)

	// 设置格式
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 设置日志级别
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	return log
}

