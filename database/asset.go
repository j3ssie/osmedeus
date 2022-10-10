package database

// Asset store task to do every single time
type Asset struct {
	Model
	AssetValue string `gorm:"type:varchar(255)" json:"asset_value"`

	//DnsRefer uint `json:"dns_refer"`
	Dns []Dns `gorm:"foreignKey:AssetRefer;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"dns"`

	HTTP          []HTTP          `gorm:"foreignKey:AssetRefer;constraint:OnUpdate:CASCADE" json:"http"`
	Directory     []Directory     `gorm:"foreignKey:AssetRefer;constraint:OnUpdate:CASCADE" json:"directory"`
	Vulnerability []Vulnerability `gorm:"foreignKey:AssetRefer;constraint:OnUpdate:CASCADE" json:"vulnerability"`

	Link    []Link    `gorm:"foreignKey:AssetRefer;constraint:OnUpdate:CASCADE" json:"link"`
	Archive []Archive `gorm:"foreignKey:AssetRefer;constraint:OnUpdate:CASCADE" json:"archive"`

	Technology string `gorm:"type:longtext;" json:"technology"`
	Score      uint   `json:"score"`
	IsAlive    bool   `json:"is_alive"`

	ScanRefer   uint `gorm:"index:idx_scan" json:"scan_refer"`
	TargetRefer uint `gorm:"index:idx_id" json:"target_refer"`
}

// Dns store task to do every single time
type Dns struct {
	Model
	// IP Address
	DnsValue string `gorm:"type:varchar(255)" json:"dns_value"`
	// A, CNAME, NX
	DnsType string `gorm:"type:varchar(255);default:'A'" json:"dns_type"`

	Domain      string `gorm:"type:varchar(255)" json:"domain"`
	DnsChecksum string `gorm:"type:varchar(255)" json:"dns_checksum"`
	Ports       string `gorm:"type:longtext;" json:"ports"`

	AssetRefer  uint `gorm:"index:idx_asset" json:"asset_refer"`
	ScanRefer   uint `gorm:"index:idx_scan" json:"scan_refer"`
	TargetRefer uint `gorm:"index:idx_id" json:"target_refer"`
}

// HTTP store task to do every single time
type HTTP struct {
	Model
	URL      string `gorm:"type:varchar(255);" json:"url"`
	Redirect string `gorm:"type:varchar(255)" json:"redirect"`
	Domain   string `gorm:"type:varchar(255)" json:"domain"`

	Title         string `gorm:"type:varchar(255);" json:"title"`
	Checksum      string `gorm:"type:varchar(255);" json:"checksum"`
	StatusCode    int    `json:"status_code"`
	ContentLength int    `json:"content_length"`

	ScreenShotData string `gorm:"type:longtext;" json:"screen_shot_data"`
	HTTPContent    string `gorm:"type:longtext;" json:"http_content"`
	HasChanged     bool   `json:"has_changed"`

	AssetRefer  uint `gorm:"index:idx_asset" json:"asset_refer"`
	ScanRefer   uint `gorm:"index:idx_scan" json:"scan_refer"`
	TargetRefer uint `gorm:"index:idx_id" json:"target_refer"`
}

// CertInfo store task to do every single time
type CertInfo struct {
	Model
	// IP Address
	Domain   string `gorm:"type:varchar(255)" json:"domain"`
	CertInfo string `gorm:"type:varchar(255)" json:"cert_info"`
	OrgInfo  string `gorm:"type:varchar(255)" json:"org_info"`

	IsWildcard   bool   `json:"is_wildcard"`
	CertChecksum string `gorm:"type:varchar(255)" json:"cert_checksum"`
	TargetRefer  uint   `gorm:"index:idx_id" json:"target_refer"`
}

// Directory store task to do every single time
type Directory struct {
	Model
	URL string `json:"url"`

	Status        int `json:"status_code"`
	ContentLength int `json:"content_length"`
	Words         int `json:"words"`

	RedirectURL string `json:"redirect_url"`
	Checksum    string `gorm:"type:varchar(255);" json:"checksum"`

	AssetRefer  uint `gorm:"index:idx_asset" json:"asset_refer"`
	ScanRefer   uint `gorm:"index:idx_scan" json:"scan_refer"`
	TargetRefer uint `gorm:"index:idx_id" json:"target_refer"`
}

// Link store task to do every single time
type Link struct {
	Model
	// Javascript, Archive, etc
	LinkType string `json:"link_type"`
	// body
	LinkSource string `json:"link_source"`
	URL        string `json:"url"`

	// https://google.com/sample
	LinkValue    string `gorm:"type:TEXT;" json:"link_value"`
	LinkChecksum string `gorm:"type:varchar(255)" json:"link_checksum"`

	AssetRefer  uint `gorm:"index:idx_asset" json:"asset_refer"`
	ScanRefer   uint `gorm:"index:idx_scan" json:"scan_refer"`
	TargetRefer uint `gorm:"index:idx_id" json:"target_refer"`
}

type Archive struct {
	Model
	ArchiveValue string `gorm:"type:TEXT;" json:"archive_value"`

	ArchiveChecksum string `gorm:"type:varchar(255)" json:"archive_checksum"`

	AssetRefer  uint `gorm:"index:idx_asset" json:"asset_refer"`
	ScanRefer   uint `gorm:"index:idx_scan" json:"scan_refer"`
	TargetRefer uint `gorm:"index:idx_id" json:"target_refer"`
}

/* Data belongs to targets */

type IPRange struct {
	Model
	Value string `gorm:"type:varchar(255);" json:"value"`
	Info  string `gorm:"type:longtext;" json:"info"`

	Country  string `gorm:"type:varchar(255);" json:"country"`
	ASNumber string `gorm:"type:varchar(255);" json:"as_number"`
	Amount   uint   `gorm:"type:int;" json:"amount"`

	IPRangeChecksum string `gorm:"type:varchar(255)" json:"archive_checksum"`
	TargetRefer     uint   `gorm:"index:idx_id" json:"target_refer"`
	ScanRefer       uint   `json:"scan_refer"`
}

type CloudBrute struct {
	Model
	CloudDomain string `gorm:"type:varchar(255);" json:"cloud_domain"`
	Status      string `gorm:"type:varchar(255);" json:"status"`
	RawData     string `gorm:"type:longtext;" json:"raw_data"`

	TargetRefer uint `gorm:"index:idx_id" json:"target_refer"`
	ScanRefer   uint `json:"scan_refer"`
}

type Credential struct {
	Model
	CredID         string `gorm:"type:varchar(255)" json:"cred_id"`
	CredChecksum   string `gorm:"type:varchar(255)" json:"cred_checksum"`
	Email          string `gorm:"type:varchar(255);" json:"email"`
	Username       string `gorm:"type:varchar(255);" json:"username"`
	Password       string `gorm:"type:varchar(255);" json:"password"`
	HashedPassword string `gorm:"type:varchar(255);" json:"hashed_password"`
	Name           string `gorm:"type:varchar(255);" json:"name"`
	Phone          string `gorm:"type:varchar(255);" json:"phone"`
	IPAddress      string `gorm:"type:varchar(255);" json:"ip_address"`
	Source         string `gorm:"type:varchar(255);" json:"source"`

	TargetRefer uint `gorm:"index:idx_id" json:"target_refer"`
	ScanRefer   uint `json:"scan_refer"`
}

type AssetChanges struct {
	Model
	OldValue string `gorm:"type:longtext;" json:"old_value"`
	NewValue string `gorm:"type:longtext;" json:"new_value"`

	ChangeType        string       // dns, http
	Notification      Notification `gorm:"foreignkey:NotificationRefer;association_autoupdate:false;association_autocreate:false"; json:"notification"`
	NotificationRefer uint         `gorm:"index:idx_id" json:"notification_refer"`
}
