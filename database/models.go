package database

import (
	"time"
)

type Model struct {
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Schedule store task to do every single time
type Schedule struct {
	Model
	TaskName       string `gorm:"type:varchar(255)" json:"task_name"`
	RefreshSeconds string `gorm:"type:varchar(255)" json:"refresh_seconds"`
	Command        string `gorm:"type:longtext" json:"command"`
	Status         string `gorm:"type:varchar(255)" json:"status"`
}

///////////

type Scan struct {
	Model

	// input part
	InputName string `gorm:"type:varchar(255);not null" json:"input_name"`
	InputType string `gorm:"type:varchar(255);default:'general'" json:"input_type"`

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

	// progress checking
	IsRunning bool `json:"is_running"`
	IsDone    bool `json:"is_done"`
	IsNew     bool `json:"is_new"`
	IsError   bool `json:"is_error"`
	IsStarted bool `json:"is_started"`

	// if the task is running by cloud provider
	IsPrepared bool   `json:"is_prepared"`
	IsCloud    bool   `json:"is_cloud"`
	CloudInfo  string `json:"cloud_info"`

	Target Target `json:"target"`
}

// runtime object
type Target struct {
	InputName string `gorm:"type:varchar(255);unique;not null" json:"input_name"`
	// @NOTE: below field shouldn't be show in UI
	// Workspace == InputName but strip out '/'
	Workspace string `gorm:"type:varchar(255);unique;not null" json:"workspace"`
	InputType string `gorm:"type:varchar(255);default:'N/A'" json:"input_type"`

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
	IsNew      bool `json:"is_new"`
	IsWildCard bool `json:"is_wildcard"`

	Reports []Report `json:"reports"`
}

// Report store reports file record
type Report struct {
	ReportName string `gorm:"type:varchar(255)" json:"report_name"`
	ReportPath string `gorm:"type:longtext" json:"report_path"`

	Module     string `gorm:"type:varchar(255)" json:"module"`
	ModulePath string `gorm:"type:longtext" json:"module_path"`

	WorkspaceName string `gorm:"type:varchar(255)" json:"workspace_name"`
	ReportType    string `gorm:"type:varchar(255);default:'text'" json:"report_type"`
}
