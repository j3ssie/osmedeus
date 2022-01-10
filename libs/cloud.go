package libs

// Cloud struct define folder to push data
type Cloud struct {
    CheckingLimit      bool
    ReBuildBaseImage   bool
    IgnoreConfigFile   bool
    BackgroundRun      bool
    OnlyCreateDroplet  bool
    OnlyCreateInstance bool
    NoDelete           bool
    EnablePrivateIP    bool

    EnableSyncWorkflow   bool
    RemoteWorkflowFolder string
    TokensFile           string

    CopyWorkspaceToGit bool
    ClearTime          string
    InstanceName       string
    TempTarget         string
    // content of secret key to avoid reading it too much
    SecretKeyContent string
    PublicKeyContent string

    // enable terraform
    EnableTerraform bool

    // chunk options
    ChunkInputs      string
    BaseWorkspace    string
    LocalSyncFolder  string
    DisableLocalSync bool
    RemoteRunList    bool
    TargetAsFile     bool
    EnableChunk      bool
    NumberOfParts    int
    Threads          int

    // specific cloud instance resources
    Size        string
    Region      string
    Token       string
    Provider    string
    IgnoreSetup bool

    // for pre and post commands
    RemotePreRun []string
    // run script on local machine after scan done
    LocalSteps   []Step `yaml:"local_steps"`
    LocalPreRun  []string
    LocalPostRun []string

    // use to clone build-osm repo
    SecretKey   string
    PublicKey   string
    BuildRepo   string
    Binary      string
    CloudWait   string
    Retry       int
    UnzipResult bool

    // raw command here
    Extra      string
    Flow       string
    Module     string
    Workspace  string
    RawCommand string
    Params     []string

    WsSource string
    WsDest   string

    Input      string
    Inputs     []string
    InputsFile string

    Target map[string]string
}

// Request all information about request
type Request struct {
    Timeout  int
    Repeat   int
    Scheme   string
    Host     string
    Port     string
    Path     string
    URL      string
    Proxy    string
    Method   string
    Redirect bool
    Headers  []map[string]string
    Body     string
    Beautify string
}

// Response all information about response
type Response struct {
    HasPopUp       bool
    StatusCode     int
    Status         string
    ContentType    string
    Headers        []map[string]string
    Body           string
    ResponseTime   float64
    Length         int
    Beautify       string
    Location       string
    BeautifyHeader string
}
