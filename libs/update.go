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

type UpdateMetaData struct {
    WorkflowVersion string `json:"workflow_version"`
    CoreVersion     string `json:"core_version"`
    UpdatedAt       string `json:"updated_at"`
}
