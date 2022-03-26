package libs

// Update some config path
type Update struct {
    UpdateURL    string // url to download the update script
    UpdateScript string
    MetaDataURL  string
    UpdateKey    string //
    UpdateType   string // git, http
    UpdateConfig string // ~/.osmedeus/update

    UpdateVersion string
    UpdateFolder  string
    UpdateDate    string
    CleanOldData  bool
    VulnUpdate  bool
    GenerateMeta  string
    ForceUpdate   bool
    IsUpdateBin   bool
    EnableUpdate  bool
    NoUpdate      bool
}

type UpdateMetaData struct {
    WorkflowVersion string `json:"workflow_version"`
    CoreVersion     string `json:"core_version"`
    UpdatedAt       string `json:"updated_at"`
}
