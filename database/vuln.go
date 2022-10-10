package database

import (
	"fmt"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/thoas/go-funk"
)

type Vulnerability struct {
	Model
	// (STO, CVE, CORS, Leak, ETC)
	VulnerabilityTitle string `gorm:"type:varchar(255);" json:"vulnerability_title"`

	// base64 content
	VulnRequest     string `gorm:"type:longtext;" json:"vuln_request"`
	VulnResponse    string `gorm:"type:longtext;" json:"vuln_response"`
	DetectionString string `gorm:"type:longtext;" json:"detection_string"`

	VulnChecksum string `gorm:"type:varchar(255);" json:"vuln_checksum"`
	SignatureID  string `gorm:"type:varchar(255);" json:"signature_id"`

	// Info, Low, Medium, High, Critical
	Severity string `gorm:"type:varchar(255);" json:"severity"`
	// Tentative, Firm, Certain
	Confidence string `gorm:"type:varchar(255);" json:"confidence"`

	// from Jaeles, Nuclei
	Source string `gorm:"type:varchar(255);" json:"source"`
	URL    string `json:"url"`

	AssetRefer  uint `gorm:"index:idx_asset" json:"asset_refer"`
	ScanRefer   uint `gorm:"index:idx_scan" json:"scan_refer"`
	TargetRefer uint `gorm:"index:idx_id" json:"target_refer"`
}

func (v *Vulnerability) Create() error {
	var assetObj Asset
	domain, err := utils.GetDomain(v.URL)
	if err != nil {
		return fmt.Errorf("error get domain from URL")
	}

	vulnChecksum := fmt.Sprintf("%s-%s-%d", v.SignatureID, domain, v.ScanRefer)
	v.VulnChecksum = utils.GenHash(vulnChecksum)

	DB.Table("Assets").Where("asset_value = ?", domain).First(&assetObj)
	if assetObj.ID == 0 {
		assetObj = Asset{
			Score:         0,
			IsAlive:       true,
			Vulnerability: []Vulnerability{*v},
			ScanRefer:     v.ScanRefer,
			TargetRefer:   v.TargetRefer,
		}
		DB.Create(&assetObj)
	}

	v.AssetRefer = assetObj.ID
	var vulnObj Vulnerability
	DB.Table("Vulnerabilities").Where("vuln_checksum = ?", v.VulnChecksum).First(&vulnObj)

	if vulnObj.ID == 0 {
		DB.Create(v)
		v.CreateNoti()

		assetObj.IsAlive = true
		assetObj.Vulnerability = append(assetObj.Vulnerability, *v)
		DB.Save(&assetObj)
		//fmt.Println("tx.RowsAffected --> ", tx.RowsAffected)
		//fmt.Println("tx.Error --> ", tx.Error)
		//
		//spew.Dump(assetObj)
		utils.DebugF("New Vulnerability record ID:%v -- assetID:%v", v.ID, assetObj.ID)
		return fmt.Errorf("new-vuln")
	}
	return nil
}

func (d *Directory) Create() error {
	var assetObj Asset
	domain, err := utils.GetDomain(d.URL)
	if err != nil {
		return fmt.Errorf("error get domain from URL")
	}

	DB.Table("Assets").Where("asset_value = ?", domain).First(&assetObj)

	tx := DB.Begin()
	if assetObj.ID == 0 {
		assetObj = Asset{
			Score:     0,
			IsAlive:   true,
			Directory: []Directory{*d},
			ScanRefer: d.ScanRefer,
		}
		DB.Create(&assetObj)
	}

	var dirbObj Directory
	tx = DB.Table("Directories").Where("url = ?", d.URL).First(&dirbObj)
	if tx.RowsAffected == 0 {
		DB.Create(d)
		dirbObj.ID = d.ID
	}

	assetObj.IsAlive = true
	assetObj.Directory = append(assetObj.Directory, *d)
	assetObj.Directory = funk.Uniq(assetObj.Directory).([]Directory)
	DB.Save(&assetObj)
	return nil
}
