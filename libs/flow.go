package libs

// Routine for each scan
type Routine struct {
    RoutineName   string
    FlowFolder    string `yaml:"flow"`
    Timeout       string `yaml:"timeout"`
    ParsedModules []Module
    Modules       []string
}

// Flow struct to define specific field for a mode
type Flow struct {
    NoDB        bool `yaml:"nodb"`
    ForceParams bool `yaml:"force-params"`
    Input       string
    Validator   string // domain, cidr, ip or domain-file, cidr-file and so on

    Name        string
    Type        string
    DefaultType string
    Desc        string
    Usage       string

    Params   []map[string]string
    Routines []Routine

    RemotePreRun []string `yaml:"remote_pre_run"`
    // run script on local machine after scan done
    LocalPreRun  []string `yaml:"local_pre_run"`
    LocalPostRun []string `yaml:"local_post_run"`
}

// Module struct to define specific field for a module
type Module struct {
    NoDB        bool   `yaml:"nodb"`
    Validator   string // domain, cidr, ip
    ForceParams bool   `yaml:"force-params"`

    // just for print some info
    Name  string
    Desc  string
    Usage string

    // enable resume, if all reports file exist then skip the module
    Resume bool
    // run module despite resume enable
    Forced bool

    MTimeout   string `yaml:"mtimeout"`
    Params     []map[string]string
    ModulePath string

    PreRun []string `yaml:"pre_run"`
    Report struct {
        Final []string
        Noti  []string
        Diff  []string
    }
    Steps   []Step
    PostRun []string `yaml:"post_run"`

    RemotePreRun []string `yaml:"remote_pre_run"`
    // run script on local machine after scan done
    LocalSteps   []Step   `yaml:"local_steps"`
    LocalPreRun  []string `yaml:"local_pre_run"`
    LocalPostRun []string `yaml:"local_post_run"`
}
