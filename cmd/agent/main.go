package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/agent/client"
	"github.com/open-beagle/beagle-wind-game/internal/agent/server"
)

var (
	runMode    = flag.String("mode", "both", "运行模式: server, client, both")
	serverAddr = flag.String("server", "localhost:50051", "服务器地址")
	nodeID     = flag.String("node", "", "节点ID (如不指定，将自动生成)")
)

func main() {
	flag.Parse()

	fmt.Println("Beagle Wind Game - Agent示例程序")
	fmt.Printf("运行模式: %s\n", *runMode)

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 根据运行模式启动相应组件
	switch *runMode {
	case "server":
		runServer(ctx, sigChan)
	case "client":
		runClient(ctx, sigChan)
	case "both":
		runBoth(ctx, sigChan)
	default:
		fmt.Printf("无效的运行模式: %s\n", *runMode)
		os.Exit(1)
	}
}

// runServer 运行服务端
func runServer(ctx context.Context, sigChan chan os.Signal) {
	fmt.Println("启动Agent服务端...")

	// 创建服务器选项
	opts := server.DefaultServerOptions
	opts.ListenAddr = *serverAddr

	// 创建服务端
	srv := server.NewAgentServer(opts)

	// 启动服务
	go func() {
		err := srv.Start()
		if err != nil {
			log.Fatalf("服务端启动失败: %v", err)
		}
	}()

	// 等待信号
	<-sigChan
	fmt.Println("正在关闭服务端...")
	srv.Stop()
	fmt.Println("服务端已关闭")
}

// runClient 运行客户端
func runClient(ctx context.Context, sigChan chan os.Signal) {
	fmt.Println("启动Agent客户端...")

	// 创建客户端选项
	opts := client.DefaultClientOptions
	opts.ServerAddr = *serverAddr
	if *nodeID != "" {
		opts.NodeID = *nodeID
	}

	// 创建客户端
	cli := client.NewAgentClient(opts)

	// 启动客户端
	err := cli.Start()
	if err != nil {
		log.Fatalf("客户端启动失败: %v", err)
	}

	// 订阅事件
	eventChan, err := cli.SubscribeEvents(ctx, []string{"container", "pipeline"})
	if err != nil {
		log.Printf("订阅事件失败: %v", err)
	} else {
		// 处理事件
		go func() {
			for event := range eventChan {
				log.Printf("收到事件: %s, 消息: %s", event.Type, event.Message)
			}
		}()
	}

	// 等待信号
	<-sigChan
	fmt.Println("正在关闭客户端...")
	cli.Stop()
	fmt.Println("客户端已关闭")
}

// runBoth 同时运行服务端和客户端
func runBoth(ctx context.Context, sigChan chan os.Signal) {
	fmt.Println("同时启动Agent服务端和客户端...")

	// 创建服务器选项
	serverOpts := server.DefaultServerOptions
	serverOpts.ListenAddr = *serverAddr

	// 创建服务端
	srv := server.NewAgentServer(serverOpts)

	// 启动服务
	go func() {
		err := srv.Start()
		if err != nil {
			log.Fatalf("服务端启动失败: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	// 创建客户端选项
	clientOpts := client.DefaultClientOptions
	clientOpts.ServerAddr = *serverAddr
	if *nodeID != "" {
		clientOpts.NodeID = *nodeID
	}

	// 创建客户端
	cli := client.NewAgentClient(clientOpts)

	// 启动客户端
	err := cli.Start()
	if err != nil {
		log.Printf("客户端启动失败: %v", err)
	} else {
		// 订阅事件
		eventCtx, eventCancel := context.WithCancel(ctx)
		defer eventCancel()

		eventChan, err := cli.SubscribeEvents(eventCtx, []string{"container", "pipeline"})
		if err != nil {
			log.Printf("订阅事件失败: %v", err)
		} else {
			// 处理事件
			go func() {
				for event := range eventChan {
					log.Printf("收到事件: %s, 消息: %s", event.Type, event.Message)
				}
			}()
		}
	}

	// 等待信号
	<-sigChan
	fmt.Println("正在关闭服务端和客户端...")
	cli.Stop()
	srv.Stop()
	fmt.Println("服务端和客户端已关闭")
}
