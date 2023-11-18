package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/provider"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var v *viper.Viper

// InitConfig Init the config
func InitConfig(options *libs.Options) error {
	// ~/.osmedeus
	RootFolder := filepath.Dir(utils.NormalizePath(options.ConfigFile))
	if !utils.FolderExists(RootFolder) {
		if err := os.MkdirAll(RootFolder, 0750); err != nil {
			return err
		}
	}

	// Base folder
	// ~/osmedeus-base
	BaseFolder := utils.NormalizePath(options.Env.BaseFolder)
	if !utils.FolderExists(BaseFolder) {
		fmt.Printf("%v Base folder not found at: %v\n", color.RedString("[Panic]"), color.HiGreenString(BaseFolder))
		fmt.Printf(color.HiYellowString("[!]")+" Consider running the installation script first: %v\n", color.HiGreenString("bash <(curl -fsSL %v)", libs.INSTALL))
		fmt.Printf(color.HiYellowString("[!]")+" Or better visit the installation guide at: %v\n", color.HiMagentaString("https://docs.osmedeus.org/installation/"))
		os.Exit(-1)
	}

	// load all the tokens
	options.TokenConfigFile = path.Join(BaseFolder, "token/osm-var.yaml")
	if !utils.FolderExists(path.Dir(options.TokenConfigFile)) {
		utils.MakeDir(path.Dir(options.TokenConfigFile))
	}

	options.Env.WorkspacesFolder = utils.NormalizePath(options.Env.WorkspacesFolder)
	if !utils.FolderExists(options.Env.WorkspacesFolder) {
		utils.MakeDir(options.Env.WorkspacesFolder)
	}

	// init config
	options.ConfigFile = utils.NormalizePath(options.ConfigFile)
	v = viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(options.ConfigFile)
	v.AddConfigPath(path.Dir(options.ConfigFile))

	err := v.ReadInConfig()
	if err != nil {
		secret := utils.GenHash(utils.RandomString(8) + utils.GetTS())[:32] // only 32 char
		prefix := secret[len(secret)-20 : len(secret)-1]

		// set some default config if config file doesn't exist
		v.SetDefault("Server", map[string]string{
			"bind":        "0.0.0.0:8000",
			"cors":        "*",
			"secret":      secret,
			"prefix":      prefix,
			"ui":          path.Join(RootFolder, "server/ui"),
			"cert_file":   path.Join(RootFolder, "server/ssl/cert.pem"),
			"key_file":    path.Join(RootFolder, "server/ssl/key.pem"),
			"master_pass": "",
		})

		v.SetDefault("Tactic", map[string]any{
			"default":    runtime.NumCPU() * 8, // 8,16,32,64
			"aggressive": runtime.NumCPU() * 16,
			"gently":     runtime.NumCPU() * 2,
		})

		v.SetDefault("Mics", map[string]string{
			"docs": utils.GetOSEnv("OSM_DOCS", libs.DOCS),
		})

		// DB connection config
		dbPath := utils.NormalizePath(path.Join(RootFolder, "sqlite.db"))
		v.SetDefault("Database", map[string]string{
			"db_host": utils.GetOSEnv("DB_HOST", "127.0.0.1"),
			"db_port": utils.GetOSEnv("DB_PORT", "3306"),
			"db_name": utils.GetOSEnv("DB_NAME", "osm-core"),
			"db_user": utils.GetOSEnv("DB_USER", "root"),
			"db_pass": utils.GetOSEnv("DB_PASS", ""),
			// default will be file system
			"db_path": utils.GetOSEnv("DB_PATH", dbPath),
			// sqlite or mysql
			"db_type": utils.GetOSEnv("DB_TYPE", "filesystem"),
		})

		// default user
		password := utils.GenHash(utils.GetTS())[:15]
		v.SetDefault("Client", map[string]string{
			"username": "osmedeus",
			"password": password,
			"jwt":      "",
			"dest":     "http://127.0.0.1:8000",
		})

		v.SetDefault("Environments", map[string]string{
			// RootFolder --> ~/.osmedeus/
			"storages":        path.Join(RootFolder, "storages"),
			"backups":         path.Join(RootFolder, "backups"),
			"provider_config": path.Join(RootFolder, "provider"),
			"instances":       path.Join(RootFolder, "instances"),

			// store all the result
			"workspaces": options.Env.WorkspacesFolder,

			// this update occasionally
			// BaseFolder --> ~/osmedeus-base/
			"workflows":    path.Join(BaseFolder, "workflow"),
			"binaries":     path.Join(BaseFolder, "binaries"),
			"data":         path.Join(BaseFolder, "data"),
			"cloud_config": path.Join(BaseFolder, "cloud"),
		})

		if err := v.WriteConfigAs(options.ConfigFile); err != nil {
			utils.ErrorF("Error writing config file: %s", err)
		}
		utils.InforF("Created a new configuration file at %s", color.HiCyanString(options.ConfigFile))
	}

	if isWritable, _ := utils.IsWritable(options.ConfigFile); isWritable {
		utils.ErrorF("config file does not writable: %v", color.HiCyanString(options.ConfigFile))
		utils.BlockF("fatal", "Make sure you are login as 'root user' if your installation done via root user")
		os.Exit(-1)
	}
	return nil
}

func LoadConfig(options *libs.Options) *viper.Viper {
	v = viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(options.ConfigFile)
	v.AddConfigPath(path.Dir(options.ConfigFile))

	if err := v.ReadInConfig(); err != nil {
		utils.ErrorF("Error reading config file, %s", err)
	}
	return v
}

func ParsingConfig(options *libs.Options) {
	v = LoadConfig(options)
	GetEnv(options)
	GetServer(options)
	GetClient(options)
	SetupOpt(options)
	// get the config for cloud provider
	GetCloud(options)
	SetupOSEnv(options)
}

// GetEnv get environment options
func GetEnv(options *libs.Options) {
	envs := v.GetStringMapString("Environments")

	options.Env.BinariesFolder = utils.NormalizePath(envs["binaries"])
	utils.MakeDir(options.Env.BinariesFolder)

	if options.Env.WorkFlowsFolder != "" {
		options.Env.WorkFlowsFolder = utils.NormalizePath(options.Env.WorkFlowsFolder)
	}

	options.Env.DataFolder = utils.NormalizePath(envs["data"])
	utils.MakeDir(options.Env.DataFolder)
	// ose folder
	options.Env.OseFolder = path.Join(options.Env.BaseFolder, "ose")
	utils.MakeDir(options.Env.DataFolder)
	options.Env.BackupFolder = utils.NormalizePath(envs["backups"])
	utils.MakeDir(options.Env.BackupFolder)
	options.Env.UIFolder = path.Join(options.Env.BaseFolder, "ui")

	// local data
	options.Env.StoragesFolder = utils.NormalizePath(envs["storages"])
	utils.MakeDir(options.Env.StoragesFolder)
	options.Env.WorkspacesFolder = utils.NormalizePath(envs["workspaces"])
	utils.MakeDir(options.Env.WorkspacesFolder)

	customWorkflow := utils.GetOSEnv("CUSTOM_OSM_WORKFLOW", "CUSTOM_OSM_WORKFLOW")
	if customWorkflow != "CUSTOM_OSM_WORKFLOW" && options.Env.WorkFlowsFolder == "" {
		options.Env.WorkFlowsFolder = utils.NormalizePath(customWorkflow)
	}

	if options.Env.WorkFlowsFolder == "" {
		options.Env.WorkFlowsFolder = utils.NormalizePath(envs["workflows"])
	}

	// @NOTE: well of course you can rebuild the core engine binary to bypass this check
	// However, the premium package primarily focuses on exclusive workflow and specialized wordlists.
	// see more about it here: https://docs.osmedeus.org/faq/#premium-package-related-questions
	// and https://docs.osmedeus.org/premium/
	if utils.FileExists(path.Join(options.Env.WorkFlowsFolder, "premium.md")) {
		options.PremiumPackage = true
	}

	// cloud stuff

	// ~/.osmedeus/providers/
	options.Env.ProviderFolder = utils.NormalizePath(envs["provider_config"])
	options.Env.InstancesFolder = utils.NormalizePath(envs["instances"])
	if options.Env.InstancesFolder == "" {
		options.Env.InstancesFolder = utils.NormalizePath(path.Join(options.Env.RootFolder, "instances"))
	}
	utils.MakeDir(options.Env.ProviderFolder)
	utils.MakeDir(options.Env.InstancesFolder)

	// ~/osmedeus-base/clouds/
	options.Env.CloudConfigFolder = utils.NormalizePath(envs["cloud_config"])
}

// SetupOpt get storage repos
func SetupOpt(options *libs.Options) {
	// auto append PATH with Plugin folder
	osPATH := utils.GetOSEnv("PATH", "PATH")
	if !strings.Contains(osPATH, options.Env.BinariesFolder) {
		utils.DebugF("Append $PATH with: %s", options.Env.BinariesFolder)
		os.Setenv("PATH", fmt.Sprintf("%s:%s", osPATH, strings.TrimRight(options.Env.BinariesFolder, "/")))
	}

	tactics := v.GetStringMap("Tactic")
	options.Tactics = strings.ToLower(options.Tactics)
	options.ThreadsHold.Default = cast.ToInt(tactics["default"])
	options.ThreadsHold.Aggressive = cast.ToInt(tactics["aggressive"])
	options.ThreadsHold.Gently = cast.ToInt(tactics["gently"])

	// try to autocorrect the tactic name
	switch options.Tactics {
	case "aggressive", "agg", "aggr", "aggrsive":
		options.Tactics = "aggressive"
	case "gently", "gen", "gent", "gentl":
		options.Tactics = "gently"
	}
	defaultThreadsHold := cast.ToInt(tactics[options.Tactics])

	if defaultThreadsHold == 0 {
		utils.ErrorF("tactic %s not found, switching to the %v one", color.HiRedString(options.Tactics), color.HiYellowString("default"))
		defaultThreadsHold = cast.ToInt(tactics["default"])
		options.Tactics = "default"
		// in case you're still using the old config
		if defaultThreadsHold == 0 {
			defaultThreadsHold = 4
		}
	}

	// override if you put --threads-hold flag
	if options.Threads == 0 {
		options.Threads = defaultThreadsHold
	}

	/* some special conditions below */

	// change {{.Storage}} from ~/.osmedeus/storages to ~/.osmedeus/destorages
	if options.EnableDeStorage {
		utils.DebugF("Dedicated Storage Enabled")
		options.Env.StoragesFolder = options.Git.DeStorage
		if !utils.FolderExists(options.Env.StoragesFolder) {
			utils.MakeDir(options.Env.StoragesFolder)
		}
	}
}

func GetCloud(options *libs.Options) {
	if !options.PremiumPackage {
		return
	}

	// ~/osemedeus-base/cloud/provider.yaml
	cloudConfigFile := path.Join(options.Env.CloudConfigFolder, "provider.yaml")

	options.CloudConfigFile = cloudConfigFile
	utils.DebugF("Parsing cloud config from: %s", color.HiCyanString(options.CloudConfigFile))
	providerConfigs, err := provider.ParseProvider(options.CloudConfigFile)

	if err != nil {
		utils.InforF("💡 You can start the wizard with the command: %s", color.HiCyanString("%s provider wizard", libs.BINARY))
	}

	options.Cloud.BuildRepo = providerConfigs.Builder.BuildRepo
	options.Cloud.SecretKey = utils.NormalizePath(providerConfigs.Builder.SecretKey)
	options.Cloud.PublicKey = utils.NormalizePath(providerConfigs.Builder.PublicKey)
	if options.Cloud.SecretKey == "" {
		options.Cloud.SecretKey = path.Join(options.Env.CloudConfigFolder, "ssh/cloud")
		options.Cloud.PublicKey = path.Join(options.Env.CloudConfigFolder, "ssh/cloud.pub")
	}

	// check SSH Keys
	if !utils.FileExists(options.Cloud.SecretKey) {
		keysDir := path.Dir(options.Cloud.SecretKey)
		os.RemoveAll(keysDir)
		utils.MakeDir(keysDir)

		utils.InforF("Generate SSH Key at: %v", options.Cloud.SecretKey)
		var err error
		_, err = utils.RunCommandWithErr(fmt.Sprintf(`ssh-keygen -t ed25519 -f %s -q -N ''`, options.Cloud.SecretKey))
		if err != nil {
			color.Red("[-] error generated SSH Key for cloud config at: %v", options.Cloud.SecretKey)
			return
		}
	}

	if utils.FileExists(options.Cloud.SecretKey) {
		utils.DebugF("Detected secret key: %v", options.Cloud.SecretKey)
		utils.DebugF("Detected public key: %v", options.Cloud.PublicKey)
		options.Cloud.SecretKeyContent = strings.TrimSpace(utils.GetFileContent(options.Cloud.SecretKey))
		options.Cloud.PublicKeyContent = strings.TrimSpace(utils.GetFileContent(options.Cloud.PublicKey))
	}
}

// GetServer get server options
func GetServer(options *libs.Options) {
	server := v.GetStringMapString("Server")

	options.Server.Bind = server["bind"]
	options.Server.Cors = server["cors"]
	options.Server.JWTSecret = server["secret"]
	options.Server.StaticPrefix = server["prefix"]
	options.Server.UIPath = utils.NormalizePath(server["ui"])
	utils.MakeDir(path.Dir(options.Server.UIPath))

	options.Server.MasterPassword = server["master_pass"]
	options.Server.CertFile = utils.NormalizePath(server["cert_file"])
	options.Server.KeyFile = utils.NormalizePath(server["key_file"])
	utils.MakeDir(path.Dir(options.Server.CertFile))

	db := v.GetStringMapString("Database")

	options.Server.DBPath = utils.NormalizePath(db["db_path"])
	options.Server.DBType = db["db_type"]
	// this should be remote one
	if options.Server.DBType == "mysql" {
		options.Server.DBUser = db["db_user"]
		options.Server.DBPass = db["db_pass"]
		options.Server.DBHost = db["db_host"]
		options.Server.DBPort = db["db_port"]
		options.Server.DBName = db["db_name"]

		//  “user:password@/dbname?charset=utf8&parseTime=True&loc=Local”
		cred := fmt.Sprintf("%v:%v", options.Server.DBUser, options.Server.DBPass)
		dest := fmt.Sprintf("%v:%v", options.Server.DBHost, options.Server.DBPort)
		dbURL := fmt.Sprintf("%v@tcp(%v)/%v?charset=utf8&parseTime=True&loc=Local", cred, dest, options.Server.DBName)
		options.Server.DBConnection = dbURL
	}
}

// GetClient get options for client
func GetClient(options *libs.Options) map[string]string {
	client := v.GetStringMapString("Client")
	options.Client.Username = client["username"]
	options.Client.Password = client["password"]
	return client
}

func SetTactic(options *libs.Options) {
	v = LoadConfig(options)
	if err := v.ReadInConfig(); err != nil {
		utils.ErrorF("Error reading config file, %s", err)
	}

	baseThreads := options.Threads
	utils.InforF("Set base threads to %v", color.HiCyanString("%v", baseThreads))
	v.Set("Tactic", map[string]any{
		"default":    baseThreads, // 2,4,8,16
		"aggressive": baseThreads * 4,
		"gently":     int(baseThreads / 2),
	})

	v.WriteConfig()
}
