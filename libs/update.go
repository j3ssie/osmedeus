package libs

// Update some config path
type Update struct {
    UpdateURL    string
    MetaDataURL  string
    UpdateKey    string //
    UpdateType   string // git, http
    UpdateConfig string // ~/.osmedeus/update

    UpdateVersion string
    UpdateFolder  string
    UpdateDate    string
    IsUpdateBin   bool
    EnableUpdate  bool
    NoUpdate      bool
}

type UpdateJSON struct {
    UpdateDate      string `json:"update_date"`
    CoreVersion     string `json:"core_version"`
    WorkflowVersion string `json:"workflow_version"`
}
