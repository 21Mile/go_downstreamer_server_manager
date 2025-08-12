package tcp_server

import (
	"context"
	"net"
	"strconv"
)

type tcpHandler struct {
}

func (t *tcpHandler) ServeTCP(ctx context.Context, src net.Conn) {
	src.Write([]byte("tcpHandler\n"))
}

func Run_tcp_server(port int, tcpLoggerPath *string) (*TcpServer, error) {
	addr := ":" + strconv.Itoa(port)
	// 初始化日志
	err := initTcpLogger(*tcpLoggerPath)
	if err != nil {
		tcpLogger.Fatalf("初始化日志失败: %v", err)
	}
	defer closeTcpLogger()

	// 记录服务器启动日志
	tcpLogger.Printf("开始启动TCP服务器，端口: %d, 配置路径: %s, 日志路径: %s\n", port, *tcpLoggerPath, tcpLoggerPath)

	tcpServer := TcpServer{
		Addr:    addr,
		Handler: &tcpHandler{},
	}
	// fmt.Println("Starting tcp_server at " + addr)
	go func() {
		if err := tcpServer.ListenAndServe(); err != nil {
			tcpLogger.Printf("TCP server failed:%v\n", port)
		}
	}()
	return &tcpServer, nil
	//代理测试
	//rb := load_balance.LoadBanlanceFactory(load_balance.LbWeightRoundRobin)
	//rb.Add("127.0.0.1:6001", "40")
	//proxy := proxy.NewTcpLoadBalanceReverseProxy(&tcp_middleware.TcpSliceRouterContext{}, rb)
	//tcpServ := tcp_proxy.TcpServer{Addr: addr, Handler: proxy,}
	//fmt.Println("Starting tcp_proxy at " + addr)
	//tcpServ.ListenAndServe()

	//redis服务器测试
	//rb := load_balance.LoadBanlanceFactory(load_balance.LbWeightRoundRobin)
	//rb.Add("127.0.0.1:6379", "40")
	//proxy := proxy.NewTcpLoadBalanceReverseProxy(&tcp_middleware.TcpSliceRouterContext{}, rb)
	//tcpServ := tcp_proxy.TcpServer{Addr: addr, Handler: proxy,}
	//fmt.Println("Starting tcp_proxy at " + addr)
	//tcpServ.ListenAndServe()

	//http服务器测试:
	//缺点对请求的管控不足,比如我们用来做baidu代理,因为无法更改请求host,所以很轻易把我们拒绝
	//rb := load_balance.LoadBanlanceFactory(load_balance.LbWeightRoundRobin)
	//rb.Add("127.0.0.1:2003", "40")
	////rb.Add("www.baidu.com:80", "40")
	//proxy := proxy.NewTcpLoadBalanceReverseProxy(&tcp_tcp_middleware.TcpSliceRouterContext{}, rb)
	//tcpServ := tcp_proxy.TcpServer{Addr: addr, Handler: proxy,}
	//fmt.Println("tcp_proxy start at:" + addr)
	//tcpServ.ListenAndServe()

	//websocket服务器测试:缺点对请求的管控不足
	//rb := load_balance.LoadBanlanceFactory(load_balance.LbWeightRoundRobin)
	//rb.Add("127.0.0.1:2003", "40")
	//proxy := proxy.NewTcpLoadBalanceReverseProxy(&tcp_middleware.TcpSliceRouterContext{}, rb)
	//tcpServ := tcp_proxy.TcpServer{Addr: addr, Handler: proxy,}
	//fmt.Println("Starting tcp_proxy at " + addr)
	//tcpServ.ListenAndServe()

	//http2服务器测试:缺点对请求的管控不足
	//rb := load_balance.LoadBanlanceFactory(load_balance.LbWeightRoundRobin)
	//rb.Add("127.0.0.1:3003", "40")
	//proxy := proxy.NewTcpLoadBalanceReverseProxy(&tcp_middleware.TcpSliceRouterContext{}, rb)
	//tcpServ := tcp_proxy.TcpServer{Addr: addr, Handler: proxy,}
	//fmt.Println("Starting tcp_proxy at " + addr)
	//tcpServ.ListenAndServe()
}
