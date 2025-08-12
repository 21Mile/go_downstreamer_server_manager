package main

// 一个main包里面只要有一个main函数即可
import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// Config 配置结构体
type Config struct {
	Base BaseConfig `yaml:"base"`
	HTTP HTTPConfig `yaml:"http"`
	GRPC GRPCConfig `yaml:"grpc"`
	TCP  TCPConfig  `yaml:"tcp"`
	Log  LogConfig  `yaml:"log"`
}

// BaseConfig 基础配置
type BaseConfig struct {
	DebugMode    string `yaml:"debug_mode"`
	TimeLocation string `yaml:"time_location"`
}

// HTTPConfig HTTP配置
type HTTPConfig struct {
	Addrs []string `yaml:"addrs"`
}

// GRPCConfig GRPC配置
type GRPCConfig struct {
	StreamingCount int   `yaml"streamingCount"`
	Ports          []int `yaml:"ports"`
}

// TCPConfig TCP配置
type TCPConfig struct {
	Ports []int `yaml:"ports"`
}

// LogConfig 日志配置
type LogConfig struct {
	LogLevel      string `yaml:"log_level"`
	FileWriterOn  bool   `yaml:"file_writer_on"`
	LogPath       string `yaml:"log_path"`
	ConsoleWriter bool   `yaml:"console_writer"`
	Color         bool   `yaml:"color"`
}

func ParseConfig(filename string) *Config {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 解析YAML
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("解析YAML失败: %v", err)
	}

	// 打印解析结果
	// fmt.Printf("配置解析结果:\n")
	fmt.Printf("Base配置:\n")
	fmt.Printf("  Debug模式: %s\n", config.Base.DebugMode)
	fmt.Printf("  时区: %s\n", config.Base.TimeLocation)

	fmt.Printf("\nHTTP配置:\n")
	for i, addr := range config.HTTP.Addrs {
		fmt.Printf("  地址%d: %s\n", i+1, addr)
	}

	fmt.Printf("\nGRPC配置:\n")
	for i, port := range config.GRPC.Ports {
		fmt.Printf("  地址%d: %v\n", i+1, port)
	}

	fmt.Printf("\nTCP配置:\n")
	for i, port := range config.TCP.Ports {
		fmt.Printf("  地址%d: %v\n", i+1, port)
	}

	// fmt.Printf("\n日志配置:\n")
	// fmt.Printf("  日志级别: %s\n", config.Log.LogLevel)
	// fmt.Printf("  写入文件: %t\n", config.Log.FileWriterOn)
	// fmt.Printf("  日志路径: %s\n", config.Log.LogPath)
	// fmt.Printf("  控制台输出: %t\n", config.Log.ConsoleWriter)
	// fmt.Printf("  彩色输出: %t\n", config.Log.Color)
	return &config
}
