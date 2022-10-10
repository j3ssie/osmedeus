package server

import (
	"crypto/tls"
	"fmt"
	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/j3ssie/osmedeus/database"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"gorm.io/gorm"
	"log"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var DB *gorm.DB
var Opt libs.Options

func StartServer(options libs.Options) {
	Opt = options
	var err error
	DB, err = database.InitDB(options)
	if err != nil {
		return
	}

	if options.Server.NoAuthen {
		fmt.Fprintf(os.Stderr, "[Critical] You're running the server with %v\n", color.RedString("NO AUTHENTICATION"))
	}

	app := fiber.New(fiber.Config{
		Prefork: options.Server.PreFork,
	})
	app.Use(cors.New())
	SetupRoutes(app)

	// mean enable SSL
	var enableSSL bool
	var ln net.Listener
	if !options.Server.DisableSSL {
		err = EnableSSL(options)
		if err == nil {
			enableSSL = true
		}
		cer, err := tls.LoadX509KeyPair(options.Server.CertFile, options.Server.KeyFile)
		if err != nil {
			enableSSL = false
			utils.ErrorF("error create ssl listener: %v", err)
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}

		// Create custom listener
		ln, err = tls.Listen("tcp", options.Server.Bind, config)
		if err != nil {
			utils.ErrorF("error create ssl listener: %v", err)
			enableSSL = false
		}
	}

	if enableSSL {
		utils.GoodF("Web UI available at: https://%v/ui/", options.Server.Bind)
		log.Fatal(app.Listener(ln))
	} else {
		utils.GoodF("Web UI available at: http://%v/ui/", options.Server.Bind)
		log.Fatal(app.Listen(options.Server.Bind))
	}
}

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App) {
	// for UI
	app.Static("/ui", Opt.Server.UIPath)
	app.Get("/ui/*", func(ctx *fiber.Ctx) error {
		return ctx.SendFile(path.Join(Opt.Server.UIPath, "index.html"))
	})

	// for static report file
	app.Use(fmt.Sprintf("%s/workspaces/", Opt.Server.StaticPrefix), filesystem.New(filesystem.Config{
		Root:         http.Dir(Opt.Env.WorkspacesFolder),
		Browse:       !Opt.Server.DisableWorkspaceListing,
		MaxAge:       3600,
		NotFoundFile: "",
	}))
	app.Use(fmt.Sprintf("%s/storages/", Opt.Server.StaticPrefix), filesystem.New(filesystem.Config{
		Root:         http.Dir(Opt.Env.StoragesFolder),
		Browse:       !Opt.Server.DisableWorkspaceListing,
		MaxAge:       3600,
		NotFoundFile: "",
	}))

	// for swagger document
	//app.Get("/docs/*", swagger.Handler) // default
	//app.Get("/docs/*", swagger.New(swagger.Config{ // custom
	//    URL:         "https://osmp.io/doc.json",
	//    DeepLinking: false,
	//}))

	app.Get("/ping", Ping)

	// Auth
	//api := api.Group("/auth/osmp")
	app.Post("/auth/login", Login)

	// Middleware
	api := app.Group("/api", logger.New())
	osmp := api.Group("/osmp")
	osmp.Get("/health", Protected(), Health)
	osmp.Get("/workspaces", Protected(), Workspace)
	osmp.Get("/workspace/:wsname/", Protected(), WorkspaceDetail)
	osmp.Delete("/delete/:wsname/", Protected(), DeleteWorkspace)
	osmp.Get("/scans", Protected(), Scan)
	osmp.Get("/ps", Protected(), Process)
	osmp.Get("/raw", Protected(), RawWorkspace)
	osmp.Get("/flows", Protected(), ListFlows)
	osmp.Get("/help", Protected(), HelperMessage)

	//api.Use(basicauth.New(basicauth.Config{
	//	Users: map[string]string{
	//		Opt.Client.Username: Opt.Client.Password,
	//	},
	//	Realm: "Forbidden",
	//	Unauthorized: func(c *fiber.Ctx) error {
	//		return c.SendString("404 not found")
	//	},
	//}))
	//

	//// beautify data for API
	//api.Get("/v1/target", GetTarget)
	////api.Get("/v1/scan", GetScan)
	////api.Get("/v1/asset", GetAsset)
	////api.Get("/v1/asset/detail", GetAssetDetail)
	//////api.Get("/v1/links", GetAsset)
	////
	//api.Get("/v1/:wsname/", WorkspaceDetail)
	//api.Get("/v1/noti", GetNoti)
	//api.Get("/v1/http", GetHTTP)
	//api.Get("/v1/ipspace", GetIPRange)
	//api.Get("/v1/cloudbrute", GetCloudBrute)
	//api.Get("/v1/credentials", GetCredential)
	//osmp.Get("/v1/scan", GetAssets)

	// execute endpoints
	osmp.Post("/execute", Protected(), NewScan)
	osmp.Post("/upload", Protected(), Upload)

}
