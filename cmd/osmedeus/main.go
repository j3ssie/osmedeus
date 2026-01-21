package main

import (
	"github.com/j3ssie/osmedeus/v5/pkg/cli"
)

// Build info - set via ldflags during build
var (
	BuildTime  = "unknown"
	CommitHash = "unknown"
)

// @title Osmedeus API
// @version 5.0.0
// @description Workflow Engine for Offensive Security - REST API for managing security automation workflows, scans, and distributed task execution.
// @termsOfService https://docs.osmedeus.org/terms/

// @contact.name Osmedeus Support
// @contact.url https://github.com/osmedeus
// @contact.email support@osmedeus.org

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8811
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token authentication. Format: "Bearer {token}"

func main() {
	cli.SetBuildInfo(BuildTime, CommitHash)
	cli.Execute()
}
