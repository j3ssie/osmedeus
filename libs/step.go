package libs

// Step struct to define component about a command
type Step struct {
    // timeout for commands and script
    Timeout string
    // use for run loop command
    Parallel int
    Threads  string
    Source   string

    Label string

    Conditions []string
    Required   []string

    Commands []string
    Ose      []string `yaml:"ose"`
    Scripts  []string

    // run when conditions are false
    RCommands []string `yaml:"rcommands"`
    RScripts  []string `yaml:"rscripts"`

    // post condition and script
    PConditions []string
    PScripts    []string

    //Output []string
    Std string
}
