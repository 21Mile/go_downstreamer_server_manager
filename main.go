package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/chzyer/readline"
	// 假设这些包存在并且接口和你原来的一样
)

// ---------- 全局变量和 I/O 控制 ----------
var (
	mConfig *Config
	rl      *readline.Instance

	// 打印锁，防止 monitor 与命令输出的竞争
	printMu sync.Mutex
)

func main() {
	mConfig = ParseConfig("./config.yaml")
	mConfig.Log.LogPath, _ = filepath.Abs(mConfig.Log.LogPath)
	manager := NewServerManager()

	// readline 初始化
	var err error
	rl, err = readline.NewEx(&readline.Config{
		Prompt:       "> ",
		HistoryLimit: 1000,
	})
	if err != nil {
		log.Fatalf("readline init error: %v", err)
	}
	defer rl.Close()

	// 把 log 输出重定向到 rl.Stderr() 避免打断当前输入行
	log.SetOutput(rl.Stderr())

	// 启动配置中的服务器
	startConfiguredServers(mConfig, manager)

	// 处理退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动监控（后台持续打印，但不会“打断”输入，因为每次打印后我们会调用 rl.Refresh()）
	go monitorServers(manager)

	// 启动命令处理（阻塞在 Readline）
	go handleCommands(manager, quit)

	<-quit
	// 退出清理
	printMu.Lock()
	fmt.Fprintln(rl.Stdout(), "\nShutting down all servers...")
	printMu.Unlock()
	manager.StopAll()
	printMu.Lock()
	fmt.Fprintln(rl.Stdout(), "All servers stopped. Exiting...")
	printMu.Unlock()
}

// 启动配置中的服务器（示例，保持你原有逻辑）
func startConfiguredServers(config *Config, manager *ServerManager) {
	for _, port := range config.GRPC.Ports {
		addr := strconv.Itoa(port)
		if err := manager.StartServer("grpc", addr); err != nil {
			log.Printf("Failed to start GRPC server on %s: %v", addr, err)
		}
	}
	for _, addr := range config.HTTP.Addrs {
		if err := manager.StartServer("http", addr); err != nil {
			log.Printf("Failed to start HTTP server on %s: %v", addr, err)
		}
	}
	for _, port := range config.TCP.Ports {
		addr := strconv.Itoa(port)
		if err := manager.StartServer("tcp", addr); err != nil {
			log.Printf("Failed to start TCP server on %s: %v", addr, err)
		}
	}
}

// 监控循环：持续打印状态并自增 span_time；打印后调用 rl.Refresh() 保持当前输入行不被破坏
func monitorServers(manager *ServerManager) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	span_time := 0

	for range ticker.C {
		// 先构造并打印（加锁）
		printMu.Lock()
		// 使用 readline 提供的 ClearScreen 来保持整洁（随后调用 rl.Refresh 恢复 prompt）
		readline.ClearScreen(rl.Stdout())
		displayServers(manager.GetServers(), span_time)
		printMu.Unlock()

		// 重新绘制 prompt + 当前行（这一步非常关键，可以让用户的输入保持在屏幕底部不会被“丢失”）
		rl.Refresh()

		span_time++
	}
}

// 将服务器状态输出到 rl.Stdout()
// 这里保持表格样式（每次 ClearScreen + 重绘）
func displayServers(servers []*Server, span_time int) {
	w := rl.Stdout()
	fmt.Fprintf(w, "The service has been running continuously for %v seconds.\n", span_time)
	fmt.Fprintln(w, "┌───────────────┬───────────────────┬───────────────┐")
	fmt.Fprintln(w, "│     Type      │   Address/Port    │     Status    │")
	fmt.Fprintln(w, "├───────────────┼───────────────────┼───────────────┤")

	for _, s := range servers {
		fmt.Fprintf(w, "│ %-13s │ %-17s │ %-13s │\n",
			s.Type, s.Address, s.Status)
	}

	fmt.Fprintln(w, "└───────────────┴───────────────────┴───────────────┘")
	fmt.Fprintln(w, "Enter commands: start [type] [address], stop [type] [address]")
}

// 命令处理循环：阻塞读取用户输入并处理
func handleCommands(manager *ServerManager, quit chan<- os.Signal) {
	for {
		// 直接读取一行；不再暂停监控（监控会一直绘制但不会破坏当前编辑，因为我们在 monitor 里调用了 rl.Refresh()）
		line, err := rl.Readline()
		if err != nil {
			printMu.Lock()
			fmt.Fprintln(rl.Stdout(), "\nRead error:", err)
			printMu.Unlock()
			return
		}

		cmd := strings.Fields(line)
		if len(cmd) > 0 {
			processCommand(cmd, manager, quit)
		}
	}
}

func processCommand(cmd []string, manager *ServerManager, quit chan<- os.Signal) {
	// 为了避免输出冲突，所有命令回复也用 printMu 锁
	switch cmd[0] {
	case "start":
		if len(cmd) != 3 {
			printMu.Lock()
			fmt.Fprintln(rl.Stdout(), "Usage: start <type> <address>")
			printMu.Unlock()
		} else {
			if err := manager.StartServer(cmd[1], cmd[2]); err != nil {
				printMu.Lock()
				fmt.Fprintf(rl.Stdout(), "Error: %v\n", err)
				printMu.Unlock()
			}
		}
	case "stop":
		if len(cmd) != 3 {
			printMu.Lock()
			fmt.Fprintln(rl.Stdout(), "Usage: stop <type> <address>")
			printMu.Unlock()
		} else {
			if err := manager.StopServer(cmd[1], cmd[2]); err != nil {
				printMu.Lock()
				fmt.Fprintf(rl.Stdout(), "Error: %v\n", err)
				printMu.Unlock()
			}
		}
	case "exit", "quit":
		quit <- syscall.SIGTERM
	default:
		printMu.Lock()
		fmt.Fprintln(rl.Stdout(), "Unknown command. Available: start, stop, exit")
		printMu.Unlock()
	}
}
