package core

import (
    "fmt"
    "net/url"
    "os"
    "path"
    "path/filepath"
    "strings"

    "github.com/mitchellh/go-homedir"

    "github.com/j3ssie/osmedeus/execution"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"

    "github.com/spf13/viper"
)

// InitConfig Init the config
func InitConfig(options *libs.Options) {
    RootFolder := filepath.Dir(utils.NormalizePath(options.ConfigFile))
    if !utils.FolderExists(RootFolder) {
        os.MkdirAll(RootFolder, 0750)
    }

    // Base folder
    BaseFolder := utils.NormalizePath(options.Env.BaseFolder)
    if !utils.FolderExists(BaseFolder) {
        os.MkdirAll(BaseFolder, 0750)
    }

    // init config
    v := viper.New()
    v.AddConfigPath(RootFolder)
    v.SetConfigName("config")
    v.SetConfigType("yaml")

    if !utils.FileExists(options.ConfigFile) {
        // Some default config if config file doesn't exist
        secret := utils.GenHash(utils.RandomString(8) + utils.GetTS())
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
            "cloud_data":      path.Join(RootFolder, "clouds"),
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
            // the repo format should be like this "git@gitlab.com:j3ssie/example.git",
            "secret_key":        utils.GetOSEnv("SECRET_KEY", "SECRET_KEY"),
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

        // used for scaling
        //v.SetDefault("Master", map[string]string{
        //	"host": utils.GetOSEnv("MASTER_HOST", "MASTER_HOST"),
        //	"cred": utils.GetOSEnv("MASTER_CRED", "MASTER_CRED"),
        //})
        //v.SetDefault("Pool", map[string]string{
        //	"host": utils.GetOSEnv("POOL_HOST", "POOL_HOST"),
        //	"cred": utils.GetOSEnv("POOL_CRED", "POOL_CRED"),
        //})
        // enable sync
        //v.SetDefault("Sync", map[string]string{
        //	"firebase_url":    utils.GetOSEnv("FIREBASE_URL", "FIREBASE_URL"),
        //	"firebase_prefix": utils.GetOSEnv("FIREBASE_PREFIX", "FIREBASE_PREFIX"),
        //	"firebase_pool":   utils.GetOSEnv("FIREBASE_POOL", "FIREBASE_POOL"),
        //})

        v.SetDefault("Cdn", map[string]string{
            "osm_cdn_url":    utils.GetOSEnv("OSM_CDN_URL", "OSM_CDN_URL"),
            "osm_cdn_wsurl":  utils.GetOSEnv("OSM_CDN_WSURL", "OSM_CDN_WSURL"),
            "osm_cdn_auth":   utils.GetOSEnv("OSM_CDN_AUTH", "OSM_CDN_AUTH"),
            "osm_cdn_prefix": utils.GetOSEnv("OSM_CDN_PREFIX", "OSM_CDN_PREFIX"),
            "osm_cdn_index":  utils.GetOSEnv("OSM_CDN_INDEX", "OSM_CDN_INDEX"),
            "osm_cdn_secret": utils.GetOSEnv("OSM_CDN_SECRET", "OSM_CDN_SECRET"),
        })

        v.SetDefault("Cloud", map[string]string{
            "cloud_public_key": utils.GetOSEnv("CLOUD_PUBLIC_KEY", "CLOUD_PUBLIC_KEY"),
            "cloud_secret_key": utils.GetOSEnv("CLOUD_SECRET_KEY", "CLOUD_SECRET_KEY"),
            "build_repo":       utils.GetOSEnv("CLOUD_BUILD_REPO", "CLOUD_BUILD_REPO"),
        })

        v.SetDefault("Update", map[string]string{
            "update_type": "git",
            "update_url":  utils.GetOSEnv("UPDATE_BASE_URL", "UPDATE_BASE_URL"),
            "update_date": utils.GetOSEnv("UPDATE_DATE", "UPDATE_DATE"),
            "update_meta": utils.GetOSEnv("META_URL", "META_URL"),
            "workflow":    utils.GetOSEnv("UPDATE_URL", "UPDATE_URL"),
        })

        v.SetDefault("Mics", map[string]string{
            "docs": utils.GetOSEnv("OSM_DOCS", libs.DOCS),
        })

        v.WriteConfigAs(options.ConfigFile)
    }

    if isWritable, _ := utils.IsWritable(options.ConfigFile); isWritable {
        utils.ErrorF("config file does not writable: %v", options.ConfigFile)
        utils.BlockF("fatal", "Make sure you are login as 'root user' if your installation done via root user")
        os.Exit(-1)
    }

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

}

// LoadConfig load config
func LoadConfig(options libs.Options) (*viper.Viper, error) {
    options.ConfigFile, _ = homedir.Expand(options.ConfigFile)
    RootFolder := filepath.Dir(options.ConfigFile)
    v := viper.New()
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath(RootFolder)
    // InitConfig(&options)
    if err := v.ReadInConfig(); err != nil {
        InitConfig(&options)
        return v, nil
    }
    return v, nil
}

// GetEnv get environment options
func GetEnv(options *libs.Options) {
    v, _ := LoadConfig(*options)
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
    options.Env.WorkFlowsFolder = utils.NormalizePath(envs["workflows"])

    if utils.FileExists(path.Join(options.Env.WorkFlowsFolder, "premium.md")) {
        options.PremiumPackage = true
    }

    // backup data
    options.Env.BackupFolder = utils.NormalizePath(envs["backups"])
    utils.MakeDir(options.Env.BackupFolder)

    // cloud stuff
    options.Env.ProviderFolder = utils.NormalizePath(envs["provider_config"])
    utils.MakeDir(options.Env.ProviderFolder)
    options.Env.CloudDataFolder = utils.NormalizePath(envs["cloud_data"])
    utils.MakeDir(options.Env.CloudDataFolder)
    options.Env.CloudConfigFolder = utils.NormalizePath(envs["cloud_config"])

    cloud := v.GetStringMapString("Cloud")
    options.Cloud.BuildRepo = cloud["build_repo"]

    // load the config file here
    // ~/osmedeus-base/cloud/ssh/cloud.pub
    // ~/osmedeus-base/cloud/ssh/cloud.privte
    options.Cloud.SecretKey = utils.NormalizePath(cloud["cloud_secret_key"])
    options.Cloud.PublicKey = utils.NormalizePath(cloud["cloud_public_key"])
    if utils.FileExists(options.Cloud.SecretKey) {
        options.Cloud.SecretKeyContent = strings.TrimSpace(utils.GetFileContent(options.Cloud.SecretKey))
        options.Cloud.PublicKeyContent = strings.TrimSpace(utils.GetFileContent(options.Cloud.PublicKey))
    }

    //
    //options.Env.RootFolder = utils.NormalizePath(options.Env.RootFolder)
    //options.Env.BaseFolder = utils.NormalizePath(options.Env.BaseFolder)

    update := v.GetStringMapString("Update")
    options.Update.UpdateURL = update["update_url"]
    options.Update.UpdateType = update["update_type"]
    options.Update.UpdateDate = update["update_date"]
    options.Update.UpdateKey = update["update_key"]
    options.Update.MetaDataURL = update["update_meta"]
    UpdateMetadata(*options)
}

// SetupOpt get storage repos
func SetupOpt(options *libs.Options) {
    // auto append PATH with Plugin folder
    osPATH := utils.GetOSEnv("PATH", "PATH")
    if !strings.Contains(osPATH, options.Env.BinariesFolder) {
        utils.DebugF("Append $PATH with: %s", options.Env.BinariesFolder)
        os.Setenv("PATH", fmt.Sprintf("%s:%s", osPATH, strings.TrimRight(options.Env.BinariesFolder, "/")))
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
    v, _ := LoadConfig(*options)
    storages := v.GetStringMapString("Storages")

    for k, v := range storages {
        storages[k], _ = homedir.Expand(v)
    }
    storagesOptions := make(map[string]string)
    // cloning stuff
    // path to private to push result
    storagesOptions["secret_key"] = storages["secret_key"]
    options.Storages = storagesOptions

    // load default existing key if it exists
    if options.Storages["secret_key"] == "" || options.Storages["secret_key"] == "SECRET_KEY" {
        // ~/.osmedeus/secret_key.private
        secretKey := path.Join(options.Env.RootFolder, "secret_key.private")
        if utils.FileExists(secretKey) {
            options.Storages["secret_key"] = secretKey
        } else {
            secretKey = path.Join(options.Env.BaseFolder, "secret/secret_key.private")
            if utils.FileExists(secretKey) {
                options.Storages["secret_key"] = secretKey
            }
        }
    }

    // disable git feature or no secret key found
    if options.NoGit {
        options.Storages["secret_key"] = ""
    }
    if options.Storages["secret_key"] == "" {
        options.NoGit = true
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
    v, _ := LoadConfig(*options)
    noti := v.GetStringMapString("Notification")
    tokens := v.GetStringMapString("Tokens")

    options.Noti.SlackToken = tokens["slack"]
    options.Noti.TelegramToken = tokens["telegram"]
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

// GetServer get server options
func GetServer(options *libs.Options) {
    v, _ := LoadConfig(*options)
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
    v, _ := LoadConfig(*options)
    master := v.GetStringMapString("Master")
    pool := v.GetStringMapString("Pool")

    options.Remote.MasterHost = master["host"]
    options.Remote.MasterCred = master["cred"]
    options.Remote.PoolHost = pool["host"]
    options.Remote.PoolCred = pool["cred"]
}

// GetClient get options for client
func GetClient(options *libs.Options) map[string]string {
    v, _ := LoadConfig(*options)
    client := v.GetStringMapString("Client")
    options.Client.Username = client["username"]
    options.Client.Password = client["password"]
    return client
}

// GetGit get options for client
func GetGit(options *libs.Options) {
    v, _ := LoadConfig(*options)
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
//
//// GetSync get options for client
//func GetSync(options *libs.Options) {
//    v, _ := LoadConfig(*options)
//    fb := v.GetStringMapString("Sync")
//    options.Sync.BaseURL = fb["firebase_url"]
//    options.Sync.Prefix = fb["firebase_prefix"]
//    options.Sync.Pool = fb["firebase_pool"]
//}

// GetCdn get options for client
func GetCdn(options *libs.Options) {
    v, _ := LoadConfig(*options)
    cdn := v.GetStringMapString("Cdn")
    options.Cdn.URL = cdn["osm_cdn_url"]
    options.Cdn.WSURL = cdn["osm_cdn_wsurl"]
    options.Cdn.Prefix = cdn["osm_cdn_prefix"]
    options.Cdn.Index = cdn["osm_cdn_index"]

    // in case we have prefix
    if options.Cdn.Prefix != "OSM_CDN_PREFIX" {
        u, err := url.Parse(options.Cdn.URL)
        if err == nil {
            u.Path = path.Join(u.Path, options.Cdn.Prefix)
            options.Cdn.URL = u.String()
        }

        u, err = url.Parse(options.Cdn.WSURL)
        if err == nil {
            u.Path = path.Join(u.Path, options.Cdn.Prefix)
            options.Cdn.WSURL = u.String()
        }
    }

    options.Cdn.Auth = cdn["osm_cdn_auth"]
}

// ReloadConfig get credentials
func ReloadConfig(options libs.Options) {
    utils.InforF("Reload Env for config file: %v", options.ConfigFile)
    v, _ := LoadConfig(options)

    // options.ConfigFile, _ = homedir.Expand(options.ConfigFile)
    RootFolder := filepath.Dir(utils.NormalizePath(options.ConfigFile))
    if !utils.FolderExists(RootFolder) {
        os.MkdirAll(RootFolder, 0750)
    }
    // Base folder
    BaseFolder := utils.NormalizePath(options.Env.BaseFolder)
    if !utils.FolderExists(BaseFolder) {
        os.MkdirAll(BaseFolder, 0750)
    }

    v.Set("Environments", map[string]string{
        // RootFolder --> ~/.osmedeus/
        "storages":        path.Join(RootFolder, "storages"),
        "workspaces":      path.Join(RootFolder, "workspaces"),
        "backups":         path.Join(RootFolder, "backups"),
        "cloud_data":      path.Join(RootFolder, "clouds"),
        "provider_config": path.Join(RootFolder, "provider"),

        // this update casually
        // BaseFolder --> ~/osmedeus-base/
        "workflows":    path.Join(BaseFolder, "workflow"),
        "binaries":     path.Join(BaseFolder, "binaries"),
        "data":         path.Join(BaseFolder, "data"),
        "cloud_config": path.Join(BaseFolder, "cloud"),
    })

    v.Set("Cloud", map[string]string{
        "cloud_secret_key": utils.GetOSEnv("CLOUD_SECRET_KEY", "CLOUD_SECRET_KEY"),
        "cloud_public_key": utils.GetOSEnv("CLOUD_PUBLIC_KEY", "CLOUD_PUBLIC_KEY"),
        "build_repo":       utils.GetOSEnv("CLOUD_BUILD_REPO", "CLOUD_BUILD_REPO"),
    })

    // things should be reload by env
    v.Set("Storages", map[string]string{
        // path of secret key for push result
        // the repo format should be like this "git@gitlab.com:j3ssie/example.git",
        "secret_key":        utils.GetOSEnv("SECRET_KEY", "SECRET_KEY"),
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

    v.Set("Cdn", map[string]string{
        "osm_cdn_url":    utils.GetOSEnv("OSM_CDN_URL", "OSM_CDN_URL"),
        "osm_cdn_wsurl":  utils.GetOSEnv("OSM_CDN_WSURL", "OSM_CDN_WSURL"),
        "osm_cdn_auth":   utils.GetOSEnv("OSM_CDN_AUTH", "OSM_CDN_AUTH"),
        "osm_cdn_prefix": utils.GetOSEnv("OSM_CDN_PREFIX", "OSM_CDN_PREFIX"),
        "osm_cdn_index":  utils.GetOSEnv("OSM_CDN_INDEX", "OSM_CDN_INDEX"),
        "osm_cdn_secret": utils.GetOSEnv("OSM_CDN_SECRET", "OSM_CDN_SECRET"),
    })

    v.Set("Update", map[string]string{
        "update_type": "git",
        "update_url":  utils.GetOSEnv("UPDATE_BASE_URL", "UPDATE_BASE_URL"),
        "update_date": utils.GetOSEnv("UPDATE_DATE", "UPDATE_DATE"),
        "update_meta": utils.GetOSEnv("META_URL", "META_URL"),
        "workflow":    utils.GetOSEnv("UPDATE_URL", "UPDATE_URL"),
    })

    v.WriteConfig()
}
