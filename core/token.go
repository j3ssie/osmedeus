package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/execution"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/spf13/viper"
)

func SetupOSEnv(options *libs.Options) {
	utils.DebugF("Loading all environment variables from: %v", color.HiGreenString(options.TokenConfigFile))
	v = viper.New()
	v.SetConfigName("osm-var")
	v.SetConfigType("yaml")
	v.AddConfigPath(path.Dir(options.TokenConfigFile))

	// Read the configuration file
	err := v.ReadInConfig()
	if err != nil {
		utils.ErrorF("Error reading config file: %s", err)
		v.SetDefault("Storages", map[string]string{
			// path of secret key for push result
			// ~/.osmedeus/secret/storages_key
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
			"client_name":                GetPublicIP(),
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

		// default tokens but you're feel free to add more and it will be automatically loaded to ENV
		v.SetDefault("Tokens", map[string]string{
			"slack":    utils.GetOSEnv("SLACK_API_TOKEN", "SLACK_API_TOKEN"),
			"telegram": utils.GetOSEnv("TELEGRAM_API_TOKEN", "TELEGRAM_API_TOKEN"),
			"gitlab":   utils.GetOSEnv("GITLAB_API_TOKEN", "GITLAB_API_TOKEN"),
			"github":   utils.GetOSEnv("GITHUB_API_KEY", "GITHUB_API_KEY"),
		})

		if ok := v.WriteConfigAs(options.TokenConfigFile); ok != nil {
			utils.ErrorF("Error writing config file: %s", ok)
		}
		utils.InforF("Created a new token configuration file at %s", color.HiCyanString(options.TokenConfigFile))
	}

	if err := v.ReadInConfig(); err != nil {
		utils.ErrorF("Error reading config file, %s", err)
	}
	// Read tokens as a list of maps
	tokens := v.GetStringMapString("tokens")
	if len(tokens) > 0 {
		utils.DebugF("Adding %v tokens to the environment variables", color.HiMagentaString("%v", len(tokens)))
		// Iterate through each token
		for name, value := range tokens {
			// automatic convert to upper case
			name = strings.ToUpper(name)

			// skip if the value is equal to the name
			if name == value {
				continue
			}

			redactedValue := "*****"
			if len(value) > 5 {
				redactedValue = value[:2] + "***" + value[len(value)-2:]
			}
			utils.DebugF("Setting environment variable: %v -- %v", name, redactedValue)

			err := os.Setenv(name, value)
			if err != nil {
				utils.ErrorF("Error setting environment variable: %v -- %v", name, err)
			}
		}
	}

	// get all the config that need to be set manually
	utils.DebugF("Getting all the config that need to be set manually")
	GetStorages(options)
	GetNotification(options)
	GetGit(options)
	GetCdn(options)
}

func LoadTokenFile(options *libs.Options) *viper.Viper {
	v = viper.New()
	v.SetConfigName("osm-var")
	v.SetConfigType("yaml")
	v.AddConfigPath(path.Dir(options.TokenConfigFile))
	if err := v.ReadInConfig(); err != nil {
		utils.ErrorF("Error reading config file, %s", err)
	}
	return v
}

func GetStorages(options *libs.Options) {
	if !options.PremiumPackage {
		return
	}

	utils.DebugF("Loading git storages from: %v", color.HiGreenString(options.TokenConfigFile))
	v = LoadTokenFile(options)
	storages := v.GetStringMapString("storages")

	// get variables from config.yaml
	storagesOptions := make(map[string]string)
	for k, dest := range storages {
		storages[k] = utils.NormalizePath(dest)
		storagesOptions[k] = utils.NormalizePath(dest)
	}
	secretKey := storages["secret_key"]
	if secretKey == "" || secretKey == "SECRET_KEY" {
		utils.DebugF("No secret key has been set. Quitting the storages setup")
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
	v = LoadTokenFile(options)

	noti := v.GetStringMapString("notification")
	tokens := v.GetStringMapString("tokens")
	// tokens
	options.Noti.SlackToken = tokens["slack_api_token"]
	options.Noti.TelegramToken = tokens["telegram_api_token"]

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

// GetCdn get options for client
func GetCdn(options *libs.Options) {
	v = LoadTokenFile(options)

	cdn := v.GetStringMapString("cdn")
	options.Cdn.Bucket = cdn["cdn_s3_bucket"]
	options.Cdn.AccessKeyId = cdn["cdn_aws_access_key"]
	options.Cdn.SecretKey = cdn["cdn_aws_secret_key"]
	options.Cdn.Region = cdn["cdn_aws_region"]
}

// GetGit get options for client
func GetGit(options *libs.Options) {
	v = LoadTokenFile(options)
	git := v.GetStringMapString("git")
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

func GetPublicIP() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	clientName := hostname + "-" + utils.RandomString(4)

	url := "https://ipinfo.io/ip"
	resp, err := http.Get(url)
	if err == nil {
		if ip, ok := io.ReadAll(resp.Body); ok == nil {
			clientName = string(ip)
		}
	}
	defer resp.Body.Close()
	return clientName
}

func SetClientName(options *libs.Options) {
	options.Noti.ClientName = GetPublicIP()

	// load token file
	v = LoadTokenFile(options)
	if err := v.ReadInConfig(); err != nil {
		utils.ErrorF("Error reading config file, %s", err)
	}
	utils.InforF("Setting up client name: %v", color.HiCyanString(options.Noti.ClientName))
	v.Set("Notification", map[string]any{
		"client_name": options.Noti.ClientName,
	})
	v.WriteConfig()

}
