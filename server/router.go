package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"

	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// var DB *gorm.DB
var Opt libs.Options

func StartServer(options libs.Options) {
	Opt = options
	var err error

	if options.Server.NoAuthen {
		fmt.Fprintf(os.Stderr, color.RedString("[Critical] The server is currently being executed %v mechanism enabled.\n", color.HiYellowString("WITHOUT ANY AUTHENTICATION")))
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
		utils.GoodF("Web UI available at: %v ", color.HiMagentaString("https://%v/ui/", options.Server.Bind))
		utils.GoodF("Static Content available at: %v", color.HiMagentaString("https://%v/%s/workspaces/", options.Server.Bind, Opt.Server.StaticPrefix))
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
	api := app.Group("/api", logger.New())
	api.Post("/login", Login)

	// disable JWT Middleware when -A is set
	if !Opt.Server.NoAuthen {
		app.Use(jwtware.New(jwtware.Config{
			SigningKey:     []byte(Opt.Server.JWTSecret),
			Filter:         nil,
			SuccessHandler: nil,
			ErrorHandler:   jwtError,
			SigningKeys:    nil,
			SigningMethod:  "",
			ContextKey:     "",
			Claims:         nil,
			TokenLookup:    "",
			AuthScheme:     "Osmedeus",
		}))
	}

	// Middleware
	osmp := api.Group("/osmp")
	osmp.Get("/health", Health)

	// core API e.g: /api/osmp/workspaces
	osmp.Get("/workspaces", ListWorkspaces)
	osmp.Get("/workspace/:wsname/", WorkspaceDetail)
	osmp.Get("/scans", ListAllScan)
	osmp.Delete("/delete/:wsname/", DeleteWorkspace)

	osmp.Get("/ps", Process)
	osmp.Get("/raw", RawWorkspace)
	osmp.Get("/flows", ListFlows)
	osmp.Get("/help", HelperMessage)

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

	// execute endpoints
	osmp.Post("/execute", NewScan)
	osmp.Post("/upload", Upload)
}
