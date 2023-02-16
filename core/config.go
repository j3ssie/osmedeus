package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/execution"
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

	// init config
	options.ConfigFile = utils.NormalizePath(options.ConfigFile)
	v = viper.New()
	v.AddConfigPath(options.ConfigFile)
	v.SetConfigType("yaml")

	if !utils.FileExists(options.ConfigFile) {
		// Some default config if config file doesn't exist
		secret := utils.GenHash(utils.RandomString(8) + utils.GetTS())[:32] // only 32 char
		prefix := secret[len(secret)-20 : len(secret)-1]

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

		// DB connection config
		dbPath := utils.NormalizePath(path.Join(RootFolder, "sqlite.db"))
		v.SetDefault("Database", map[string]string{
			"db_host": utils.GetOSEnv("DB_HOST", "127.0.0.1"),
			"db_port": utils.GetOSEnv("DB_PORT", "3306"),
			"db_name": utils.GetOSEnv("DB_NAME", "osm-core"),
			"db_user": utils.GetOSEnv("DB_USER", "root"),
			"db_pass": utils.GetOSEnv("DB_PASS", ""),
			// sqlite or mysql
			"db_path": utils.GetOSEnv("DB_PATH", dbPath),
			"db_type": utils.GetOSEnv("DB_TYPE", "sqlite"),
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
			"workspaces":      path.Join(RootFolder, "workspaces"),
			"backups":         path.Join(RootFolder, "backups"),
			"provider_config": path.Join(RootFolder, "provider"),

			// this update casually
			// BaseFolder --> ~/osmedeus-base/
			"workflows":    path.Join(BaseFolder, "workflow"),
			"binaries":     path.Join(BaseFolder, "binaries"),
			"data":         path.Join(BaseFolder, "data"),
			"cloud_config": path.Join(BaseFolder, "cloud"),
		})

		// things should be reloaded by env
		v.SetDefault("Storages", map[string]string{
			// path of secret key for push result
			// ~/.osmedeus/storages_key
			"secret_key": utils.GetOSEnv("SECRET_KEY", "SECRET_KEY"),
			// the repo format should be like this "git@gitlab.com:j3ssie/example.git",
			"summary_storage":   path.Join(options.Env.RootFolder, "storages/summary"),
			"summary_repo":      utils.GetOSEnv("SUMMARY_REPO", "SUMMARY_REPO"),
			"subdomain_storage": path.Join(options.Env.RootFolder, "storages/subdomain"),
			"subdomain_repo":    utils.GetOSEnv("SUBDOMAIN_REPO", "SUBDOMAIN_REPO"),
			"assets_storage":    path.Join(options.Env.RootFolder, "storages/assets"),
			"assets_repo":       utils.GetOSEnv("ASSETS_REPO", "ASSETS_REPO"),
			"ports_storage":     path.Join(options.Env.RootFolder, "storages/ports"),
			"ports_repo":        utils.GetOSEnv("PORTS_REPO", "PORTS_REPO"),
			"http_storage":      path.Join(options.Env.RootFolder, "storages/http"),
			"http_repo":         utils.GetOSEnv("HTTP_REPO", "HTTP_REPO"),
			"vuln_storage":      path.Join(options.Env.RootFolder, "storages/vuln"),
			"vuln_repo":         utils.GetOSEnv("VULN_REPO", "VULN_REPO"),
			"paths_storage":     path.Join(options.Env.RootFolder, "storages/paths"),
			"paths_repo":        utils.GetOSEnv("PATHS_REPO", "PATHS_REPO"),
			"mics_storage":      path.Join(options.Env.RootFolder, "storages/mics"),
			"mics_repo":         utils.GetOSEnv("MICS_REPO", "MICS_REPO"),
		})

		v.SetDefault("Tokens", map[string]string{
			"slack":    utils.GetOSEnv("SLACK_API_TOKEN", "SLACK_API_TOKEN"),
			"telegram": utils.GetOSEnv("TELEGRAM_API_TOKEN", "TELEGRAM_API_TOKEN"),
			"gitlab":   utils.GetOSEnv("GITLAB_API_TOKEN", "GITLAB_API_TOKEN"),
			"github":   utils.GetOSEnv("GITHUB_API_KEY", "GITHUB_API_KEY"),
		})

		// dedicated storages
		v.SetDefault("Git", map[string]string{
			"base_url":     utils.GetOSEnv("GITLAB_BASE_URL", "https://gitlab.com"),
			"api":          utils.GetOSEnv("GITLAB_API_TOKEN", "GITLAB_API_TOKEN"),
			"username":     utils.GetOSEnv("GITLAB_USER", "GITLAB_USER"),
			"password":     utils.GetOSEnv("GITLAB_PASS", "GITLAB_PASS"),
			"group":        utils.GetOSEnv("GITLAB_GROUP", "GITLAB_GROUP"),
			"prefix_name":  utils.GetOSEnv("GITLAB_PREFIX_NAME", "deosm"),
			"default_tag":  utils.GetOSEnv("GITLAB_DEFAULT_TAG", "osmd"),
			"default_user": utils.GetOSEnv("GITLAB_DEFAULT_USER", "j3ssie"),
			"default_uid":  utils.GetOSEnv("GITLAB_DEFAULT_UID", "3537075"),
			"destorage":    path.Join(options.Env.RootFolder, "destorage"),
		})

		v.SetDefault("Notification", map[string]string{
			"client_name":                utils.GetOSEnv("CLIENT_NAME", "CLIENT_NAME"),
			"slack_status_channel":       utils.GetOSEnv("SLACK_STATUS_CHANNEL", "SLACK_STATUS_CHANNEL"),
			"slack_report_channel":       utils.GetOSEnv("SLACK_REPORT_CHANNEL", "SLACK_REPORT_CHANNEL"),
			"slack_diff_channel":         utils.GetOSEnv("SLACK_DIFF_CHANNEL", "SLACK_DIFF_CHANNEL"),
			"slack_webhook":              utils.GetOSEnv("SLACK_WEBHOOK", "SLACK_WEBHOOK"),
			"telegram_channel":           utils.GetOSEnv("TELEGRAM_CHANNEL", "TELEGRAM_CHANNEL"),
			"telegram_status_channel":    utils.GetOSEnv("TELEGRAM_STATUS_CHANNEL", "TELEGRAM_STATUS_CHANNEL"),
			"telegram_report_channel":    utils.GetOSEnv("TELEGRAM_REPORT_CHANNEL", "TELEGRAM_REPORT_CHANNEL"),
			"telegram_sensitive_channel": utils.GetOSEnv("TELEGRAM_SENSITIVE_CHANNEL", "TELEGRAM_SENSITIVE_CHANNEL"),
			"telegram_dirb_channel":      utils.GetOSEnv("TELEGRAM_DIRB_CHANNEL", "TELEGRAM_DIRB_CHANNEL"),
			"telegram_mics_channel":      utils.GetOSEnv("TELEGRAM_MICS_CHANNEL", "TELEGRAM_MICS_CHANNEL"),
		})

		v.SetDefault("Cdn", map[string]string{
			"cdn_s3_bucket":      utils.GetOSEnv("CDN_S3_BUCKET", "CDN_S3_BUCKET"),
			"cdn_aws_access_key": utils.GetOSEnv("CDN_AWS_ACCESS_KEY", "CDN_AWS_ACCESS_KEY"),
			"cdn_aws_secret_key": utils.GetOSEnv("CDN_AWS_SECRET_KEY", "CDN_AWS_SECRET_KEY"),
			"cdn_aws_region":     utils.GetOSEnv("CDN_AWS_REGION", "ap-southeast-1"),
		})

		v.SetDefault("Tactic", map[string]any{
			"default":    runtime.NumCPU(), // 2,4,8,16
			"aggressive": runtime.NumCPU() * 4,
			"gently":     int(runtime.NumCPU() + 1/2),
		})

		v.SetDefault("Mics", map[string]string{
			"docs": utils.GetOSEnv("OSM_DOCS", libs.DOCS),
		})

		if err := v.WriteConfigAs(options.ConfigFile); err != nil {
			utils.ErrorF("Error writing config file: %s", err)
		}

		utils.InforF("Write config file to %s", options.ConfigFile)
	}

	if isWritable, _ := utils.IsWritable(options.ConfigFile); isWritable {
		utils.ErrorF("config file does not writable: %v", options.ConfigFile)
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
	GetStorages(options)
	GetNotification(options)
	GetServer(options)
	GetClient(options)
	GetRemote(options)
	GetGit(options)
	//GetSync(options)
	GetCdn(options)
	SetupOpt(options)
	GetCloud(options)
}

// GetEnv get environment options
func GetEnv(options *libs.Options) {
	envs := v.GetStringMapString("Environments")

	// config
	options.Env.BinariesFolder = utils.NormalizePath(envs["binaries"])
	utils.MakeDir(options.Env.BinariesFolder)

	options.Env.DataFolder = utils.NormalizePath(envs["data"])
	utils.MakeDir(options.Env.DataFolder)
	// ose folder
	options.Env.OseFolder = path.Join(options.Env.BaseFolder, "ose")
	utils.MakeDir(options.Env.DataFolder)

	options.Env.ScriptsFolder = path.Join(options.Env.BaseFolder, "scripts")
	options.Env.UIFolder = path.Join(options.Env.BaseFolder, "ui")

	// local data
	options.Env.StoragesFolder = utils.NormalizePath(envs["storages"])
	utils.MakeDir(options.Env.StoragesFolder)
	options.Env.WorkspacesFolder = utils.NormalizePath(envs["workspaces"])
	utils.MakeDir(options.Env.WorkspacesFolder)

	// get workflow folder
	if options.Env.WorkFlowsFolder != "" {
		options.Env.WorkFlowsFolder = utils.NormalizePath(options.Env.WorkFlowsFolder)
	}

	customWorkflow := utils.GetOSEnv("CUSTOM_OSM_WORKFLOW", "CUSTOM_OSM_WORKFLOW")
	if customWorkflow != "CUSTOM_OSM_WORKFLOW" && options.Env.WorkFlowsFolder == "" {
		options.Env.WorkFlowsFolder = utils.NormalizePath(customWorkflow)
	}

	if options.Env.WorkFlowsFolder == "" {
		options.Env.WorkFlowsFolder = utils.NormalizePath(envs["workflows"])
	}

	// @NOTE: well of course you can rebuild the core engine binary to bypass this check
	// but the premium package is more about custom workflow + wordlists + tools
	// see more about it here: https://docs.osmedeus.org/faq/#premium-package-related-questions and  https://docs.osmedeus.org/premium/
	if utils.FileExists(path.Join(options.Env.WorkFlowsFolder, "premium.md")) {
		options.PremiumPackage = true
	}

	// backup data
	options.Env.BackupFolder = utils.NormalizePath(envs["backups"])
	utils.MakeDir(options.Env.BackupFolder)

	// cloud stuff

	// ~/.osmedeus/providers/
	options.Env.ProviderFolder = utils.NormalizePath(envs["provider_config"])
	utils.MakeDir(options.Env.ProviderFolder)
	// ~/osmedeus-base/clouds/
	options.Env.CloudConfigFolder = utils.NormalizePath(envs["cloud_config"])

	//update := v.GetStringMapString("Update")
	//options.Update.UpdateURL = update["update_url"]
	//options.Update.UpdateType = update["update_type"]
	//options.Update.UpdateDate = update["update_date"]
	//options.Update.UpdateKey = update["update_key"]
	//options.Update.MetaDataURL = update["update_meta"]
	//UpdateMetadata(*options)
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

func GetStorages(options *libs.Options) {
	if !options.PremiumPackage || utils.GetOSEnv("ENABLE_GIT_STORAGES", "") != "TRUE" {
		return
	}

	storages := v.GetStringMapString("Storages")

	// get variables from config.yaml
	storagesOptions := make(map[string]string)
	for k, dest := range storages {
		storages[k] = utils.NormalizePath(dest)
		storagesOptions[k] = utils.NormalizePath(dest)
	}
	secretKey := storages["secret_key"]
	if secretKey == "" || secretKey == "SECRET_KEY" {
		return
	}

	storagesOptions["secret_key"] = secretKey
	// load default existing key if it exists
	defaultKey := path.Join(options.Env.BaseFolder, "secret/storages_key")
	if !utils.FileExists(secretKey) && utils.FileExists(defaultKey) {
		utils.InforF("Loaded default secret for storages from: %v", color.HiCyanString(defaultKey))
		if _, err := utils.RunCommandWithErr(fmt.Sprintf("cp %s %s && chmod 600 %s", defaultKey, secretKey, secretKey)); err != nil {
			utils.ErrorF("error copying default secret key: %v", defaultKey)
		}
	}

	if !utils.FileExists(secretKey) {
		utils.InforF("No SSH key for storages found. Generate a new one at: %v", color.HiCyanString(secretKey))
		if _, err := utils.RunCommandWithErr(fmt.Sprintf(`ssh-keygen -t ed25519 -f %s -q -N ''`, secretKey)); err != nil {
			color.Red("[-] error generated SSH Key for storages at: %v", secretKey)
			return
		}
		utils.InforF("Please add the public key at %v to your gitlab profile", color.HiCyanString(secretKey+".pub"))
	}

	if !utils.FileExists("~/.gitconfig") {
		utils.WarnF("Looks like you didn't set up the git user at %v yet", color.HiCyanString("~/.gitconfig"))
		utils.WarnF("💡 Init git info with this command: %s", color.HiCyanString(`git config --global user.name "your-username" && git config --global user.email "your-username@users.noreply.gitlab.com"`))
	}

	if options.CustomGit {
		// in case custom repo is set
		for _, env := range os.Environ() {
			// the ENV should be OSM_SUMMARY_STORAGE, OSM_SUMMARY_REPO
			if strings.HasSuffix(env, "OSM_") {
				data := strings.Split(env, "=")
				key := strings.ToLower(data[0])
				value := strings.Replace(env, data[0]+"=", "", -1)

				if strings.HasSuffix(key, "summary_storage") {
					storagesOptions["summary_storage"] = value
				}
				if strings.HasSuffix(key, "summary_repo") {
					storagesOptions["summary_repo"] = value
				}

				if strings.HasSuffix(key, "assets_storage") {
					storagesOptions["assets_storage"] = value
				}
				if strings.HasSuffix(key, "assets_repo") {
					storagesOptions["assets_repo"] = value
				}

				if strings.HasSuffix(key, "ports_storage") {
					storagesOptions["ports_storage"] = value
				}
				if strings.HasSuffix(key, "ports_repo") {
					storagesOptions["ports_repo"] = value
				}
			}
		}
	}

	storagesOptions[storages["summary_storage"]] = storages["summary_repo"]
	storagesOptions[storages["subdomain_storage"]] = storages["subdomain_repo"]
	storagesOptions[storages["http_storage"]] = storages["http_repo"]
	storagesOptions[storages["assets_storage"]] = storages["assets_repo"]
	storagesOptions[storages["mics_storage"]] = storages["mics_repo"]
	storagesOptions[storages["ports_storage"]] = storages["ports_repo"]
	storagesOptions[storages["paths_storage"]] = storages["paths_repo"]
	storagesOptions[storages["vuln_storage"]] = storages["vuln_repo"]
	options.Storages = storagesOptions

	// disable git feature or no secret key found
	if options.NoGit {
		options.Storages["secret_key"] = ""
	}
	if options.Storages["secret_key"] == "" {
		options.NoGit = true
	}

	if options.NoGit {
		return
	}
	execution.CloneRepo(storages["summary_repo"], storages["summary_storage"], *options)
	execution.CloneRepo(storages["http_repo"], storages["http_storage"], *options)
	execution.CloneRepo(storages["assets_repo"], storages["assets_storage"], *options)
	execution.CloneRepo(storages["subdomain_repo"], storages["subdomain_storage"], *options)
	execution.CloneRepo(storages["ports_repo"], storages["ports_storage"], *options)
	execution.CloneRepo(storages["mics_repo"], storages["mics_storage"], *options)
	execution.CloneRepo(storages["paths_repo"], storages["paths_storage"], *options)
	execution.CloneRepo(storages["vuln_repo"], storages["vuln_storage"], *options)
}

// GetNotification get storge repos
func GetNotification(options *libs.Options) {
	noti := v.GetStringMapString("Notification")
	tokens := v.GetStringMapString("Tokens")

	// tokens
	options.Noti.SlackToken = tokens["slack"]
	options.Noti.TelegramToken = tokens["telegram"]

	// this mean you're not setup the notification yet
	if len(options.Noti.TelegramToken) < 20 {
		options.NoNoti = true
	}

	options.Noti.ClientName = noti["client_name"]

	options.Noti.SlackStatusChannel = noti["slack_status_channel"]
	options.Noti.SlackReportChannel = noti["slack_report_channel"]
	options.Noti.SlackDiffChannel = noti["slack_diff_channel"]
	options.Noti.SlackWebHook = noti["slack_webhook"]
	options.Noti.TelegramChannel = noti["telegram_channel"]
	options.Noti.TelegramSensitiveChannel = noti["telegram_sensitive_channel"]
	options.Noti.TelegramReportChannel = noti["telegram_report_channel"]
	options.Noti.TelegramStatusChannel = noti["telegram_status_channel"]
	options.Noti.TelegramDirbChannel = noti["telegram_dirb_channel"]
	options.Noti.TelegramMicsChannel = noti["telegram_mics_channel"]
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

// GetRemote get server remote options
func GetRemote(options *libs.Options) {
	master := v.GetStringMapString("Master")
	pool := v.GetStringMapString("Pool")

	options.Remote.MasterHost = master["host"]
	options.Remote.MasterCred = master["cred"]
	options.Remote.PoolHost = pool["host"]
	options.Remote.PoolCred = pool["cred"]
}

// GetCdn get options for client
func GetCdn(options *libs.Options) {
	cdn := v.GetStringMapString("Cdn")
	options.Cdn.Bucket = cdn["cdn_s3_bucket"]
	options.Cdn.AccessKeyId = cdn["cdn_aws_access_key"]
	options.Cdn.SecretKey = cdn["cdn_aws_secret_key"]
	options.Cdn.Region = cdn["cdn_aws_region"]
}

// GetClient get options for client
func GetClient(options *libs.Options) map[string]string {
	client := v.GetStringMapString("Client")
	options.Client.Username = client["username"]
	options.Client.Password = client["password"]
	return client
}

// GetGit get options for client
func GetGit(options *libs.Options) {
	git := v.GetStringMapString("Git")
	options.Git.BaseURL = git["base_url"]
	options.Git.DeStorage = git["destorage"]
	options.Git.Token = git["api"]
	options.Git.Username = git["username"]
	options.Git.Password = git["password"]
	options.Git.Group = git["group"]
	options.Git.DefaultPrefix = git["prefix_name"]
	options.Git.DefaultTag = git["default_tag"]
	options.Git.DefaultUser = git["default_user"]
	options.Git.DefaultUID = utils.StrToInt(git["default_uid"])
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

// ReloadConfig get credentials
func ReloadConfig(options *libs.Options) {
	v = LoadConfig(options)
	utils.InforF("Reload Env to the config file: %v", options.ConfigFile)

	v.Set("Cdn", map[string]string{
		"cdn_s3_bucket":      utils.GetOSEnv("CDN_S3_BUCKET", "CDN_S3_BUCKET"),
		"cdn_aws_access_key": utils.GetOSEnv("CDN_AWS_ACCESS_KEY", "CDN_AWS_ACCESS_KEY"),
		"cdn_aws_secret_key": utils.GetOSEnv("CDN_AWS_SECRET_KEY", "CDN_AWS_SECRET_KEY"),
		"cdn_aws_region":     utils.GetOSEnv("CDN_AWS_REGION", "ap-southeast-1"),
	})

	// things should be reloaded by env
	v.Set("Storages", map[string]string{
		// path of secret key for push result
		"secret_key": utils.GetOSEnv("SECRET_KEY", "SECRET_KEY"),
		// the repo format should be like this "git@gitlab.com:j3ssie/example.git",
		"summary_storage":   path.Join(options.Env.RootFolder, "storages/summary"),
		"summary_repo":      utils.GetOSEnv("SUMMARY_REPO", "SUMMARY_REPO"),
		"subdomain_storage": path.Join(options.Env.RootFolder, "storages/subdomain"),
		"subdomain_repo":    utils.GetOSEnv("SUBDOMAIN_REPO", "SUBDOMAIN_REPO"),
		"assets_storage":    path.Join(options.Env.RootFolder, "storages/assets"),
		"assets_repo":       utils.GetOSEnv("ASSETS_REPO", "ASSETS_REPO"),
		"ports_storage":     path.Join(options.Env.RootFolder, "storages/ports"),
		"ports_repo":        utils.GetOSEnv("PORTS_REPO", "PORTS_REPO"),
		"http_storage":      path.Join(options.Env.RootFolder, "storages/http"),
		"http_repo":         utils.GetOSEnv("HTTP_REPO", "HTTP_REPO"),
		"vuln_storage":      path.Join(options.Env.RootFolder, "storages/vuln"),
		"vuln_repo":         utils.GetOSEnv("VULN_REPO", "VULN_REPO"),
		"paths_storage":     path.Join(options.Env.RootFolder, "storages/paths"),
		"paths_repo":        utils.GetOSEnv("PATHS_REPO", "PATHS_REPO"),
		"mics_storage":      path.Join(options.Env.RootFolder, "storages/mics"),
		"mics_repo":         utils.GetOSEnv("MICS_REPO", "MICS_REPO"),
	})

	v.Set("Tokens", map[string]string{
		"slack":    utils.GetOSEnv("SLACK_API_TOKEN", "SLACK_API_TOKEN"),
		"gitlab":   utils.GetOSEnv("GITLAB_API_TOKEN", "GITLAB_API_TOKEN"),
		"github":   utils.GetOSEnv("GITHUB_API_KEY", "GITHUB_API_KEY"),
		"telegram": utils.GetOSEnv("TELEGRAM_API_TOKEN", "TELEGRAM_API_TOKEN"),
	})

	v.Set("Git", map[string]string{
		"base_url":     utils.GetOSEnv("GITLAB_BASE_URL", "https://gitlab.com"),
		"api":          utils.GetOSEnv("GITLAB_API_TOKEN", "GITLAB_API_TOKEN"),
		"username":     utils.GetOSEnv("GITLAB_USER", "GITLAB_USER"),
		"password":     utils.GetOSEnv("GITLAB_PASS", "GITLAB_PASS"),
		"group":        utils.GetOSEnv("GITLAB_GROUP", "GITLAB_GROUP"),
		"prefix_name":  utils.GetOSEnv("GITLAB_PREFIX_NAME", "deosm"),
		"default_tag":  utils.GetOSEnv("GITLAB_DEFAULT_TAG", "osmd"),
		"default_user": utils.GetOSEnv("GITLAB_DEFAULT_USER", "j3ssie"),
		"default_uid":  utils.GetOSEnv("GITLAB_DEFAULT_UID", "3537075"),
		"destorage":    path.Join(options.Env.RootFolder, "destorage"),
	})

	v.Set("Notification", map[string]string{
		"client_name":                utils.GetOSEnv("CLIENT_NAME", "CLIENT_NAME"),
		"slack_status_channel":       utils.GetOSEnv("SLACK_STATUS_CHANNEL", "SLACK_STATUS_CHANNEL"),
		"slack_report_channel":       utils.GetOSEnv("SLACK_REPORT_CHANNEL", "SLACK_REPORT_CHANNEL"),
		"slack_diff_channel":         utils.GetOSEnv("SLACK_DIFF_CHANNEL", "SLACK_DIFF_CHANNEL"),
		"slack_webhook":              utils.GetOSEnv("SLACK_WEBHOOK", "SLACK_WEBHOOK"),
		"telegram_channel":           utils.GetOSEnv("TELEGRAM_CHANNEL", "TELEGRAM_CHANNEL"),
		"telegram_status_channel":    utils.GetOSEnv("TELEGRAM_STATUS_CHANNEL", "TELEGRAM_STATUS_CHANNEL"),
		"telegram_report_channel":    utils.GetOSEnv("TELEGRAM_REPORT_CHANNEL", "TELEGRAM_REPORT_CHANNEL"),
		"telegram_sensitive_channel": utils.GetOSEnv("TELEGRAM_SENSITIVE_CHANNEL", "TELEGRAM_SENSITIVE_CHANNEL"),
		"telegram_dirb_channel":      utils.GetOSEnv("TELEGRAM_DIRB_CHANNEL", "TELEGRAM_DIRB_CHANNEL"),
		"telegram_mics_channel":      utils.GetOSEnv("TELEGRAM_MICS_CHANNEL", "TELEGRAM_MICS_CHANNEL"),
	})
	v.WriteConfig()
}
