package libs

// Options global options
type Options struct {
	ConfigFile  string
	LogFile     string
	Concurrency int

	// default threads hold for running module
	Tactics string
	Threads int

	Timeout           string
	EnableFormatInput bool
	Verbose           bool
	FullHelp          bool

	// some disable options
	NoPostRun            bool
	NoPreRun             bool
	NoNoti               bool
	NoBanner             bool
	NoGit                bool
	NoClean              bool
	NoDB                 bool
	NoCdn                bool
	DisableValidateInput bool

	PremiumPackage   bool
	Resume           bool
	Quite            bool
	Force            bool
	WildCardCheck    bool
	Debug            bool
	EnableDeStorage  bool
	PID              int
	SyncTimes        int
	PollingTime      int
	MDCodeBlockLimit int
	Exclude          []string
	Params           []string
	CustomGit        bool
	EnableBackup     bool
	JsonOutput       bool

	Client Client
	Queue  Queue
	Git    Git
	Sync   Sync
	Scan   Scan
	Server Server
	Env    Environment
	Noti   Notification
	Flow   Flow
	Module Module
	Tmux   TmuxOpt
	Cron   Cron
	Remote Remote
	Cdn    Cdn
	Update Update

	ThreadsHold     ThreadsHold
	Cloud           Cloud
	Report          Report
	CloudConfigFile string
	TokenConfigFile string
	GitSync         bool

	ScanID   string
	Storages map[string]string
}

// Scan sub options for scan
type Scan struct {
	ROptions   map[string]string
	Params     []string
	Input      string
	InputType  string // domain, url, ip, cidr or domainList, urlList, ipList, cidrList
	ParamsFile string
	Inputs     []string
	InputList  string
	Modules    []string
	Flow       string

	BaseWorkspace   string
	CustomWorkspace string
	SuffixName      string
	Force           bool
	// this is true when calling from cloud scan
	RemoteCall bool
}

type ThreadsHold struct {
	Default    int
	Aggressive int
	Gently     int
}

// Report sub options for report
type Report struct {
	CustomPreFix  string
	PublicIP      string
	ExtractFolder string
	Static        bool
	Raw           bool
}

// Server sub options for api server
type Server struct {
	DisableWorkspaceListing bool
	DisableSSL              bool
	PreFork                 bool
	NoAuthen                bool

	PollingTime    int
	Bind           string
	Port           string
	StaticPrefix   string
	JWTSecret      string
	Cors           string
	UIPath         string
	MasterPassword string

	// database
	DBPath       string
	DBType       string
	DBConnection string
	DBName       string
	DBUser       string
	DBPass       string
	DBHost       string
	DBPort       string

	// for SSL
	CertFile string
	KeyFile  string
}

// Storage struct define folder to push data
type Storage struct {
	SecretKey      string
	SummaryStorage string
	SummaryRepo    string
	HTTPStorage    string
	HTTPRepo       string
	AssetsStorage  string
	AssetsRepo     string
}

// Environment some config path
type Environment struct {
	RootFolder       string // ~/.osmedeus
	StoragesFolder   string // ~/.osmedeus/storages/
	WorkspacesFolder string // ~/workspaces-osmedeus/

	// Base one
	BaseFolder      string // ~/osmedeus-base
	BinariesFolder  string // ~/osmedeus-base/binaries
	DataFolder      string // ~/osmedeus-base/data/
	OseFolder       string // ~/osmedeus-base/ose/
	WorkFlowsFolder string // ~/osmedeus-base/workflow/

	// cloud stuff
	CloudConfigFolder string // ~/osmedeus-base/clouds/
	ProviderFolder    string // ~/.osmedeus/providers/
	InstancesFolder   string // ~/.osmedeus/instances/
	BackupFolder      string

	// Mics
	ScriptsFolder string
	UIFolder      string
}
