package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-agent/internal/server/grpc"
	"github.com/myczh-1/lazy-ctrl-agent/internal/server/http"
	"github.com/myczh-1/lazy-ctrl-agent/internal/server/mqtt"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/command"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/security"
	"github.com/sirupsen/logrus"
)

var (
	configPath = flag.String("config", "", "Path to config file")
	version    = flag.Bool("version", false, "Show version")
)

const (
	appVersion = "2.0.0"
	appName    = "lazy-ctrl-agent"
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s v%s\n", appName, appVersion)
		os.Exit(0)
	}

	// 初始化日志
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// 加载配置
	if err := config.Load(*configPath); err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	cfg := config.Get()
	
	// 设置日志级别
	if level, err := logrus.ParseLevel(cfg.Log.Level); err == nil {
		logger.SetLevel(level)
	}

	// 设置日志格式
	if cfg.Log.Format == "text" {
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	// 设置日志输出
	if cfg.Log.OutputPath != "" {
		file, err := os.OpenFile(cfg.Log.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.WithError(err).Warn("Failed to open log file, using stdout")
		} else {
			logger.SetOutput(file)
		}
	}

	logger.WithFields(logrus.Fields{
		"version": appVersion,
		"config":  *configPath,
	}).Info("Starting lazy-ctrl-agent")

	// 初始化服务
	commandService := command.NewService(cfg)
	executorService := executor.NewService(logger)
	securityService := security.NewService(cfg, logger)

	// 启动定期清理
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			securityService.CleanupRateLimiter()
		}
	}()

	// 创建服务器
	var servers []Server
	
	if cfg.Server.HTTP.Enabled {
		httpServer := http.NewServer(cfg, logger, commandService, executorService, securityService)
		servers = append(servers, httpServer)
		logger.WithField("port", cfg.Server.HTTP.Port).Info("HTTP server enabled")
	}

	if cfg.Server.GRPC.Enabled {
		grpcServer := grpc.NewServer(cfg, logger, commandService, executorService, securityService)
		servers = append(servers, grpcServer)
		logger.WithField("port", cfg.Server.GRPC.Port).Info("gRPC server enabled")
	}

	if cfg.MQTT.Enabled {
		mqttClient := mqtt.NewClient(cfg, logger, commandService, executorService, securityService)
		servers = append(servers, mqttClient)
		logger.WithField("broker", cfg.MQTT.Broker).Info("MQTT client enabled")
	}

	if len(servers) == 0 {
		logger.Fatal("No servers enabled in configuration")
	}

	// 启动所有服务器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	errChan := make(chan error, len(servers))

	for _, server := range servers {
		wg.Add(1)
		go func(s Server) {
			defer wg.Done()
			if err := s.Start(); err != nil {
				errChan <- fmt.Errorf("server start failed: %w", err)
			}
		}(server)
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		logger.WithError(err).Error("Server error")
		cancel()
	case sig := <-sigChan:
		logger.WithField("signal", sig).Info("Received shutdown signal")
		cancel()
	}

	// 优雅关闭
	logger.Info("Shutting down...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	shutdownChan := make(chan struct{})
	go func() {
		// 停止所有服务器
		for _, server := range servers {
			if stopper, ok := server.(Stopper); ok {
				stopper.Stop()
			}
		}
		wg.Wait()
		close(shutdownChan)
	}()

	select {
	case <-shutdownChan:
		logger.Info("Shutdown complete")
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout, forcing exit")
	}
}

// Server 接口定义
type Server interface {
	Start() error
}

// Stopper 接口定义优雅关闭
type Stopper interface {
	Stop()
}