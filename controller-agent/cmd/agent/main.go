// Package main provides the lazy-ctrl-agent server
//
//	@title						Lazy-Ctrl Agent API
//	@version					2.0.0
//	@description				Remote computer control agent with HTTP/gRPC/MQTT support
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.url				https://github.com/myczh-1/lazy-ctrl-agent
//	@contact.email				support@lazy-ctrl.com
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//	@host						localhost:7070
//	@BasePath					/api/v1
//	@securityDefinitions.apikey	PinAuth
//	@in							header
//	@name						X-Pin
//	@description				PIN authentication for secure access
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/myczh-1/lazy-ctrl-agent/internal/app"
	"github.com/myczh-1/lazy-ctrl-agent/internal/common"
)

var (
	configPath = flag.String("config", "", "Path to config file")
	version    = flag.Bool("version", false, "Show version")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s v%s\n", common.AppName, common.AppVersion)
		os.Exit(0)
	}

	// Create and run application
	application, err := app.NewApplication(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create application: %v\n", err)
		os.Exit(1)
	}

	// Run the application (this blocks until shutdown)
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}