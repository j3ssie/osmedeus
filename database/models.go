package database

import (
    "time"
)

type Model struct {
    ID        uint       `gorm:"primary_key" json:"id"`
    CreatedAt time.Time  `json:"created_at,omitempty"`
    UpdatedAt time.Time  `json:"updated_at,omitempty"`
    DeletedAt *time.Time `sql:"index" json:"deleted_at,omitempty"`
}

// Schedule store task to do every single time
type Schedule struct {
    Model
    TaskName       string `gorm:"type:varchar(255)" json:"task_name"`
    RefreshSeconds string `gorm:"type:varchar(255)" json:"refresh_seconds"`
    Command        string `gorm:"type:longtext" json:"command"`
    Status         string `gorm:"type:varchar(255)" json:"status"`
}

type Target struct {
    Model
    InputName string `gorm:"type:varchar(255);unique;not null" json:"input_name"`

    // @NOTE: below field shouldn't be show in UI
    // Workspace == InputName but strip out '/'
    Workspace string `gorm:"type:varchar(255);unique;not null" json:"workspace"`
    InputType string `gorm:"type:varchar(255);default:'N/A'" json:"input_type"`
    // TargetType string `gorm:"type:varchar(255);default:'public'"`

    // total number for stat
    TotalAssets        int `json:"total_assets"`
    TotalDns           int `json:"total_dns"`
    TotalTech          int `json:"total_tech"`
    TotalScreenShot    int `json:"total_screenshot"`
    TotalVulnerability int `json:"total_vulnerability"`
    TotalDirb          int `json:"total_dirb"`
    TotalLink          int `json:"total_link"`
    TotalArchive       int `json:"total_archive"`
    TotalIPRange       int `json:"total_ip_range"`
    TotalCloud         int `json:"total_cloud"`
    TotalCred          int `json:"total_cred"`

    // flag information
    // IsFinished bool
    IsNew      bool `json:"is_new"`
    IsWildCard bool `json:"is_wildcard"`
    IsViewOnly bool `json:"is_viewonly"` // true mean target view only, false mean normal target

    //OrgRefer uint `json:"org_refer"` // use for permission check
    // Scans    []Scan   `gorm:"foreignKey:TargetRefer"`
    Reports []Report `gorm:"foreignKey:TargetRefer" json:"reports"`
}

type Scan struct {
    Model

    // input part
    InputName string `gorm:"type:varchar(255);not null" json:"input_name"`
    InputType string `gorm:"type:varchar(255);default:'general'" json:"input_type"`

    // UID-TaskType-TaskName
    //ScanID       string `gorm:"type:varchar(255);unique" json:"scan_id"`
    //ScanCheckSum string `gorm:"type:varchar(255);unique" json:"scan_checksum"`
    //UID          string `gorm:"type:varchar(255)"` // should be unique like uuid of user

    // ModuleName string `gorm:"type:varchar(255)"`
    TaskName string `gorm:"type:varchar(255)" json:"task_name"`
    TaskPath string `gorm:"type:varchar(255)" json:"task_path"`
    TaskType string `gorm:"type:varchar(255);default:'flow'" json:"task_type"`

    RunningTime   int    `json:"running_time"` // as seconds
    CurrentModule string `gorm:"type:varchar(255)" json:"current_module"`
    DoneStep      int    `json:"done_step"`
    TotalSteps    int    `json:"total_steps"`

    // mics part
    LogFile   string `json:"log_file"`
    ProcessID int    `json:"process_id"`

    IsRunning  bool `json:"is_running"`
    IsDone     bool `json:"is_done"`
    IsFinished bool `json:"is_finished"`
    IsNew      bool `json:"is_new"`
    IsError    bool `json:"is_error"`

    // if the task is running by cloud provider
    IsPrepared bool   `json:"is_prepared"`
    IsCloud    bool   `json:"is_cloud"`
    CloudInfo  string `json:"cloud_info"`

    Target      Target `gorm:"foreignKey:TargetRefer;association_autoupdate:true;association_autocreate:false" json:"target"`
    TargetRefer uint   `json:"target_refer"`
}

// CloudInstance store cloud scan information to check
type CloudInstance struct {
    Model

    Token    string `gorm:"type:varchar(255)" json:"token"`
    Provider string `gorm:"type:varchar(255)" json:"provider"`

    InputName  string `gorm:"type:varchar(255)" json:"input_name"`
    IPAddress  string `gorm:"type:varchar(255)" json:"ip_address"`
    InstanceID string `gorm:"type:varchar(255)" json:"instance_id"`
    SnapShotID string `gorm:"type:varchar(255)" json:"snap_shot_id"`

    // running / deleted / preparing
    Status  string `gorm:"type:varchar(255)" json:"status"`
    IsError bool   `json:"is_error"`

    IsChunk bool `json:"is_chunk"`

    ScanRefer   uint   `json:"scan_refer"`
    Target      Target `gorm:"foreignKey:TargetRefer;association_autoupdate:true;association_autocreate:false" json:"target"`
    TargetRefer uint   `json:"target_refer"`
}

type User struct {
    Model
    Username string `gorm:"type:varchar(255);unique;not null" json:"username"`
    Password string `gorm:"type:varchar(255);not null" json:"password"`

    UID      string `gorm:"type:varchar(255);" json:"uid"`
    Role     string `gorm:"type:varchar(255);default:'admin'" json:"role"`
    Secret   string `gorm:"type:varchar(255);" json:"secret"`
    APIToken string `gorm:"type:varchar(255);" json:"api_token"`

    OrgRefer uint `json:"org_refer"`
}

// @TODO: should be delete when going to production

// Report store reports file record
type Report struct {
    Model

    ReportName string `gorm:"type:varchar(255)" json:"report_name"`
    ReportPath string `gorm:"type:longtext" json:"report_path"`

    Module     string `gorm:"type:varchar(255)" json:"module"`
    ModulePath string `gorm:"type:longtext" json:"module_path"`

    WorkspaceName string `gorm:"type:varchar(255)" json:"workspace_name"`
    ReportType    string `gorm:"type:varchar(255);default:'text'" json:"report_type"`

    TargetRefer uint `json:"target_refer"`
}

//
//type Org struct {
//	Model
//	Name    string   `gorm:"type:varchar(255);" json:"name"`
//	Desc    string   `gorm:"type:varchar(255);" json:"desc"`
//	Users   []User   `gorm:"foreignKey:OrgRefer" json:"org_refer"`
//	Targets []Target `gorm:"foreignKey:OrgRefer" json:"targets"`
//}

//func (s *Scan) BeforeCreate(tx *gorm.DB) (err error) {
//	// this should be change to real UID
//	rScanID := fmt.Sprintf("%s-%s-%s-%s", s.UID, s.TaskType, s.TaskName, s.InputName)
//	s.ScanID = rScanID
//	s.ScanCheckSum = utils.GenHash(rScanID)
//	return
//}
