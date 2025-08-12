package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"

	"github.com/21Mile/go_downstreamer_server/services/grpc_server"
	"github.com/21Mile/go_downstreamer_server/services/http_server"
	"github.com/21Mile/go_downstreamer_server/services/tcp_server"
	"google.golang.org/grpc"
)

type Server struct {
	Type    string
	Address string
	Status  string
	Stop    func() error
	mu      sync.Mutex
}

type ServerManager struct {
	servers map[string]*Server
	mu      sync.Mutex
}

func NewServerManager() *ServerManager {
	return &ServerManager{
		servers: make(map[string]*Server),
	}
}

// 启动服务器（支持动态添加）

func (m *ServerManager) StartServer(typ, address string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", typ, address)
	if _, exists := m.servers[key]; exists {
		if m.servers[key].Status == "running" {
			return fmt.Errorf("server %s is already running", key)
		} else {
			//否则重启服务：直接删除信息，后后续流程会自动重启服务
			delete(m.servers, key)
		}

	}

	var stopFunc func() error
	var err error

	switch typ {
	case "grpc":
		port, _ := strconv.Atoi(address)
		var s *grpc.Server
		s, err = grpc_server.Run_grpc_server(&port, &mConfig.GRPC.StreamingCount, &mConfig.Log.LogPath)
		if err == nil {
			stopFunc = func() error {
				s.GracefulStop()
				return nil
			}
		}
	case "http":
		var s *http_server.RealServer
		s, err = http_server.Run_http_server(&address, &mConfig.Log.LogPath)
		if err == nil {
			stopFunc = s.Stop
		}
	case "tcp":
		port, _ := strconv.Atoi(address)
		var tcpServer *tcp_server.TcpServer
		tcpServer, err = tcp_server.Run_tcp_server(port, &mConfig.Log.LogPath)
		if err == nil {
			stopFunc = func() error {
				return tcpServer.Close()
			}
		}
	default:
		return fmt.Errorf("unsupported server type: %s", typ)
	}

	if err != nil {
		return fmt.Errorf("failed to start %s server: %w", typ, err)
	}

	server := &Server{
		Type:    typ,
		Address: address,
		Status:  "running",
		Stop:    stopFunc,
	}
	m.servers[key] = server
	return nil
}

// 停止服务器

func (m *ServerManager) StopServer(typ, address string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", typ, address)
	server, exists := m.servers[key]
	if !exists {
		return fmt.Errorf("server %s not found", key)
	}

	if server.Status != "running" {
		return nil
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	if err := server.Stop(); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}
	server.Status = "stopped"
	return nil
}

// 获取所有服务器状态 ->返回经过排序的拷贝
func (m *ServerManager) GetServers() []*Server {
	m.mu.Lock()
	defer m.mu.Unlock()

	servers := make([]*Server, 0, len(m.servers))
	for _, s := range m.servers {
		s.mu.Lock()
		servers = append(servers, &Server{
			Type:    s.Type,
			Address: s.Address,
			Status:  s.Status,
		})
		s.mu.Unlock()
	}

	// 排序：先按 Address，再按 Type（可选再按 Status 保证完全确定性）
	sort.Slice(servers, func(i, j int) bool {
		a, b := servers[i], servers[j]
		if a.Address != b.Address {
			return a.Address < b.Address
		}
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		return a.Status < b.Status
	})

	return servers
}

// 关闭所有服务器
// 关闭所有服务器（直接对真实对象调用 Stop）
func (m *ServerManager) StopAll() {
	// 拷贝真实指针，避免持锁期间做耗时操作
	m.mu.Lock()
	servers := make([]*Server, 0, len(m.servers))
	for _, s := range m.servers {
		servers = append(servers, s)
	}
	m.mu.Unlock()

	for _, s := range servers {
		s.mu.Lock()
		if s.Status == "running" && s.Stop != nil {
			if err := s.Stop(); err != nil {
				log.Printf("server stop err: %s %s, %v", s.Type, s.Address, err)
			} else {
				s.Status = "stopped"
			}
		}
		s.mu.Unlock()
	}
}
