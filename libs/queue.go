package libs

// Queue sub options for quque
type Queue struct {
	QueueFolder string
	QueueFile   string
	RawCommand  string

	InputAsFile bool
	Add         bool
}

type InputFormat struct {
	Input       string   `json:"input"`
	Flow        string   `json:"flow"`
	Modules     []string `json:"module"`
	Params      []string `json:"params"`
	Workspaces  string   `json:"workspace"`
	Extra       string   `json:"extra"`
	Command     string   `json:"command"`
	InputAsFile bool     `json:"input-as-file"`
}
