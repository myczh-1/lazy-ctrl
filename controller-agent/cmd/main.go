package main

import (
	"log"
	"os"

	"lazy-ctrl/controller-agent/internal/service"
	"lazy-ctrl/controller-agent/pkg/config"
	"lazy-ctrl/controller-agent/pkg/server"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "controller-agent",
	Short: "Lazy Control Agent - Execute commands locally",
	Long:  "A controller agent that executes local scripts and provides communication interfaces.",
	Run:   runServer,
}

var configPath string

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "configs/commands.json", "config file path")
}

func runServer(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	commandService := service.NewCommandService(cfg)
	grpcServer := server.NewGRPCServer(commandService)

	if err := grpcServer.Start(":50051"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}