package http_server

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func Run_http_server(addr, logPath *string) (*RealServer, error) {
	// 初始化日志
	err := initGrpcLogger(*logPath)
	if err != nil {
		httpLogger.Fatalf("初始化日志失败: %v", err)
	}
	defer closeGrpcLogger()

	// 记录服务器启动日志
	httpLogger.Printf("开始启动http服务器，addr: %v, 日志路径: %s\n", *addr, *logPath)
	rs1 := &RealServer{Addr: *addr}
	// 协程处理
	go func() {
		if err := rs1.Run(); err != nil && err != http.ErrServerClosed {
			httpLogger.Println("HTTP server failed:%v", rs1.Addr)
		}
	}()

	return rs1, nil

}

type RealServer struct {
	Addr   string
	server *http.Server
}

func (r *RealServer) Run() error {
	httpLogger.Println("Starting httpserver at " + r.Addr)
	mux := http.NewServeMux()
	mux.HandleFunc("/", r.HelloHandler) //没有匹配的路径会默认匹配到这里
	mux.HandleFunc("/base/error", r.ErrorHandler)
	mux.HandleFunc("/timeout", r.TimeoutHandler)
	r.server = &http.Server{
		Addr:         r.Addr,
		WriteTimeout: time.Second * 3,
		Handler:      mux,
	}
	// 暂时不用zkp节点
	// go func() {
	// 	//注册zk节点
	// 	zkManager := zookeeper.NewZkManager([]string{"127.0.0.1:2181"})
	// 	err := zkManager.GetConnect()
	// 	if err != nil {
	// 		httpLogger.Printf(" connect zookeeper error: %s ", err)
	// 	}
	// 	defer zkManager.Close()
	// 	err = zkManager.RegistServerPath("/real_server", r.Addr)
	// 	if err != nil {
	// 		httpLogger.Printf(" regist zookeeper node error: %s ", err)
	// 	}
	// 	zlist, err := zkManager.GetServerListByPath("/real_server")
	// 	httpLogger.Println(zlist)
	// 	httpLogger.Fatal(server.ListenAndServe())
	// }()
	if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		httpLogger.Printf("HTTP serve failed : %v", r.Addr)
		return err
	}
	return nil
}

func (r *RealServer) Stop() error {
	if err := r.server.Close(); err != nil {
		httpLogger.Printf("server stop failed:%v\n", r.Addr)
		return err
	}
	return nil
}

func (r *RealServer) HelloHandler(w http.ResponseWriter, req *http.Request) {
	upath := fmt.Sprintf("http://%s%s\n", r.Addr, req.URL.Path)
	realIP := fmt.Sprintf("RemoteAddr=%s,X-Forwarded-For=%v,X-Real-Ip=%v\n", req.RemoteAddr, req.Header.Get("X-Forwarded-For"), req.Header.Get("X-Real-Ip"))
	header := fmt.Sprintf("headers =%v\n", req.Header)
	io.WriteString(w, "hello! this is real server.\n")
	io.WriteString(w, upath)
	io.WriteString(w, realIP)
	io.WriteString(w, header)
	// 写入一句hello

}

func (r *RealServer) ErrorHandler(w http.ResponseWriter, req *http.Request) {
	upath := "error handler"
	w.WriteHeader(500)
	io.WriteString(w, upath)
}

func (r *RealServer) TimeoutHandler(w http.ResponseWriter, req *http.Request) {
	upath := "timeout handler"
	time.Sleep(6 * time.Second) //超时时间(经过这段时间后返回)
	w.WriteHeader(200)          //返回状态码
	io.WriteString(w, upath)
}
