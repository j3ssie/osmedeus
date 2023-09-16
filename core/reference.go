package core

/* File to store all the script for better reference */

const (
	Cleaning         = "Cleaning"
	CleanAmass       = "CleanAmass"
	CleanRustScan    = "CleanRustScan"
	CleanGoBuster    = "CleanGoBuster"
	CleanMassdns     = "CleanMassdns"
	CleanSWebanalyze = "CleanSWebanalyze"
	CleanJSONDnsx    = "CleanJSONDnsx"
	CleanWebanalyze  = "CleanWebanalyze"
	CleanArjun       = "CleanArjun"
	GenNucleiReport  = "GenNucleiReport"
	CleanJSONHttpx   = "CleanJSONHttpx"
	CleanFFUFJson    = "CleanFFUFJson"
)

const (
	// noti for slack
	StartNoti   = "StartNoti"
	DoneNoti    = "DoneNoti"
	ReportNoti  = "ReportNoti"
	DiffNoti    = "DiffNoti"
	CustomNoti  = "CustomNoti"
	NotiFile    = "NotiFile"
	WebHookNoti = "WebHookNoti"
	// noti for telegram
	TeleMess       = "TeleMess"
	TeleMessWrap   = "TeleMessWrap"
	TeleMessByFile = "TeleMessByFile"
	TeleSendFile   = "TeleSendFile"
)

const (
	ExecCmd           = "ExecCmd"
	ExecCmdB          = "ExecCmdB"
	ExecCmdWithOutput = "ExecCmdWithOutput"
	ExecContain       = "ExecContain"
	Sleep             = "Sleep"
	Exit              = "Exit"
	CastToInt         = "CastToInt"
	StripSlash        = "StripSlash"
	Printf            = "Printf"
	Cat               = "Cat"
	SortU             = "SortU"
	SplitFile         = "SplitFile"
	Append            = "Append"
	Copy              = "Copy"
	CreateFolder      = "CreateFolder"
	DeleteFile        = "DeleteFile"
	DeleteFolder      = "DeleteFolder"
	SplitFileByPart   = "SplitFileByPart"
	FileLength        = "FileLength"
	FolderLength      = "FolderLength"
	IsFile            = "IsFile"
	EmptyDir          = "EmptyDir"
	EmptyFile         = "EmptyFile"
	ReadLines         = "ReadLines"
	Compress          = "Compress"
	Decompress        = "Decompress"
)

const (
	TotalSubdomain     = "TotalSubdomain"
	TotalDns           = "TotalDns"
	TotalScreenShot    = "TotalScreenShot"
	TotalTech          = "TotalTech"
	TotalVulnerability = "TotalVulnerability"
	TotalArchive       = "TotalArchive"
	TotalLink          = "TotalLink"
	TotalDirb          = "TotalDirb"
	CreateReport       = "CreateReport"
)

const (
	RRSync      = "RRSync"
	Clone       = "Clone"
	FClone      = "FClone"
	PushResult  = "PushResult"
	PushFolder  = "PushFolder"
	PullFolder  = "PullFolder"
	DiffCompare = "DiffCompare"
	GitDiff     = "GitDiff"
	LoopGitDiff = "LoopGitDiff"
	// for gitlab API only
	CreateRepo      = "CreateRepo"
	DeleteRepo      = "DeleteRepo"
	DeleteRepoByPid = "DeleteRepoByPid"
	ListProjects    = "ListProjects"
)

const (
	UploadToS3        = "UploadToS3"
	DownloadFromS3    = "DownloadFromS3"
	DownloadFile      = "DownloadFile"
	GenMarkdownReport = "GenMarkdownReport"
)

const (
	SetVar   = "SetVar"
	SetOSVar = "SetOSVar"
	GetOSEnv = "GetOSEnv"
)
