package database

import (
	"github.com/j3ssie/osmedeus/utils"
	"github.com/spf13/cast"
)

func ImportAssets(objs []Asset) {
	tx := DB.CreateInBatches(objs, 5000)
	if tx.RowsAffected == 0 {
		utils.DebugF("Insert asset in serial")
		for _, obj := range objs {
			if DB.Model(&obj).Where("asset_value = ?", obj.AssetValue).RowsAffected == 0 {
				DB.Create(&obj)
			}
		}
	}
}

// SummaryTarget update the target summary but with SQL queries
func (s *Scan) SummaryTarget() {
	var target Target
	DB.Model(Target{}).First(&target, s.TargetRefer)
	if target.ID == 0 {
		return
	}

	var sum int64
	DB.Model(Asset{}).Where("target_refer = ?", s.TargetRefer).Count(&sum)
	target.TotalAssets += cast.ToInt(sum)

	DB.Model(Asset{}).Where("target_refer = ? AND technology != ?", s.TargetRefer, "").Count(&sum)
	target.TotalTech += cast.ToInt(sum)

	DB.Model(Dns{}).Where("target_refer = ?", s.TargetRefer).Count(&sum)
	target.TotalDns += cast.ToInt(sum)

	DB.Model(Link{}).Where("target_refer = ?", s.TargetRefer).Count(&sum)
	target.TotalLink += cast.ToInt(sum)

	DB.Model(Archive{}).Where("target_refer = ?", s.TargetRefer).Count(&sum)
	target.TotalArchive += cast.ToInt(sum)

	DB.Model(Directory{}).Where("target_refer = ?", s.TargetRefer).Count(&sum)
	target.TotalDirb += cast.ToInt(sum)

	DB.Model(Vulnerability{}).Where("target_refer = ?", s.TargetRefer).Count(&sum)
	target.TotalVulnerability += cast.ToInt(sum)

	DB.Model(HTTP{}).Where("target_refer = ? AND screen_shot_data != ?", s.TargetRefer, "").Count(&sum)
	target.TotalScreenShot += cast.ToInt(sum)

	DB.Model(IPRange{}).Where("target_refer = ?", s.TargetRefer).Count(&sum)
	target.TotalIPRange += cast.ToInt(sum)

	DB.Model(CloudBrute{}).Where("target_refer = ?", s.TargetRefer).Count(&sum)
	target.TotalCloud += cast.ToInt(sum)

	DB.Model(Credential{}).Where("target_refer = ?", s.TargetRefer).Count(&sum)
	target.TotalCred += cast.ToInt(sum)

	DB.Save(&target)
}
