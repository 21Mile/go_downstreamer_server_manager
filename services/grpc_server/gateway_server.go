// Binary server is an example server.
package grpc_server

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	pb "github.com/21Mile/go_downstreamer_server/services/grpc_server/proto" //定义了服务接口和消息结构

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// server需要实现EchoServer的接口
type server struct {
	pb.UnimplementedEchoServer
}

var (
	streamingCount = 10
)

func (s *server) ServiceStreamingEcho(in *pb.EchoRequest, stream pb.Echo_ServiceStreamingEchoServer) error {
	grpcLogger.Printf("--- ServerStreamingEcho ---\n")
	grpcLogger.Printf("request received: %v\n", in)

	// Read requests and send responses.
	for i := 0; i < streamingCount; i++ {
		grpcLogger.Printf("echo message %v\n", in.Message)
		err := stream.Send(&pb.EchoResponse{Message: in.Message})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *server) ClientStreamingEcho(stream pb.Echo_ClientStreamingEchoServer) error {
	grpcLogger.Printf("--- ClientStreamingEcho ---\n")
	// Read requests and send responses.
	var message string
	for {
		in, err := stream.Recv() //流式接受
		if err == io.EOF {
			//持续接收客户端的消息
			grpcLogger.Printf("echo last received message\n")
			return stream.SendAndClose(&pb.EchoResponse{Message: message})
		}
		message = in.Message
		// 保存当前接收到的消息
		grpcLogger.Printf("request received: %v, building echo\n", in)
		if err != nil {
			return err
		}
	}
}

// 双向流式
func (s *server) BidirectionalStreamingEcho(stream pb.Echo_BidirectionalStreamingEchoServer) error {
	grpcLogger.Printf("--- BidirectionalStreamingEcho ---\n")
	// Read requests and send responses.
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		grpcLogger.Printf("request received %v, sending echo\n", in)
		if err := stream.Send(&pb.EchoResponse{Message: in.Message}); err != nil {
			return err
		}
	}
}

func (s *server) UnaryEcho(ctx context.Context, in *pb.EchoRequest) (*pb.EchoResponse, error) {
	grpcLogger.Printf("--- UnaryEcho ---\n")
	// --- 测试header头 ---
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		grpcLogger.Println("miss metadata from context")
	}
	//这一步需要先在dashboard中添加对应的header值
	testHeaderValue := md.Get("my_test_header_key")
	if len(testHeaderValue) > 0 {
		grpcLogger.Println(testHeaderValue)
	}
	// --- 测试结束 ---
	fmt.Println("md", md)
	grpcLogger.Printf("request received: %v, sending echo\n", in)
	return &pb.EchoResponse{Message: in.Message}, nil
}

func Run_grpc_server(port, configStreamingCount *int, logPath *string) (*grpc.Server, error) {
	// 初始化日志
	err := initGrpcLogger(*logPath)
	if err != nil {
		grpcLogger.Fatalf("初始化日志失败: %v", err)
	}
	defer closeGrpcLogger()

	// 记录服务器启动日志
	grpcLogger.Printf("开始启动gRPC服务器，端口: %d, 配置路径: %s, 日志路径: %s\n", *port, *logPath, logPath)

	streamingCount = *configStreamingCount
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port)) //创建 TCP 监听器 lis。
	if err != nil {
		grpcLogger.Fatalf("failed to listen: %v", err)
	}
	grpcLogger.Printf("grpc server listening at %v\n", lis.Addr())
	s := grpc.NewServer(
		grpc.NumStreamWorkers(32),         // 工作线程数 (默认1)
		grpc.MaxConcurrentStreams(100000), // 最大并发流 (默认100)
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Timeout:           10 * time.Second,
		}),
	) //创建 gRPC 服务器实例。
	// 一个 gRPC 服务器可以注册多个服务
	pb.RegisterEchoServer(s, &server{}) //注册 Echo 服务到 gRPC 服务器。
	// 协程启动监听，返回server句柄
	go func() {
		if err := s.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			grpcLogger.Printf("gRPC server failed: %v", err)
		}
	}()
	return s, nil
}
