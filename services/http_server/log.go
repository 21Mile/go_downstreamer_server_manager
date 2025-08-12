package http_server

import (
	"fmt"
	"log"
	"os"
	"time"
)

// 自定义日志文件
var httpLogger *log.Logger
var logFile *os.File

// 初始化日志文件
func initGrpcLogger(logPath string) error {
	// 确保日志目录存在
	err := os.MkdirAll(logPath, 0755)
	if err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 打开或创建日志文件
	filename := fmt.Sprintf("%s/http_server_%s.log", logPath, time.Now().Format("20060102_150405"))
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %v", err)
	}

	// 创建自定义logger
	httpLogger = log.New(file, "", log.LstdFlags|log.Lshortfile)
	logFile = file

	// 同时输出到控制台和文件（可选）
	// multiWriter := io.MultiWriter(os.Stdout, file)
	// grpcLogger = log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)

	return nil
}

// 关闭日志文件
func closeGrpcLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
