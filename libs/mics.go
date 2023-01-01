package libs

// Cdn credentials for other client
type Cdn struct {
	Bucket      string
	Region      string
	SecretKey   string
	AccessKeyId string
}

// Git credentials for other client
type Git struct {
	BaseURL       string
	Username      string
	Password      string
	Token         string
	Group         string
	DefaultPrefix string
	DefaultTag    string
	DefaultUser   string
	DefaultUID    int
	DeStorage     string
}

// TmuxOpt credentials for other client
type TmuxOpt struct {
	ApplyAll       bool
	SelectedWindow string
	Exclude        string
	Limit          int
}

// Cron credentials for other client
type Cron struct {
	Command  string
	Schedule int
	Forever  bool
}

// Remote credentials for other client
type Remote struct {
	MasterHost string
	MasterCred string
	PoolHost   string
	PoolCred   string
}

// Sync credentials for other client
type Sync struct {
	BaseURL string
	Prefix  string
	Pool    string
}

// Client credentials for other client
type Client struct {
	Username string
	Password string
	JWT      string
	URL      string
}
