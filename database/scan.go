package database

import (
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"path"
)

func DBNewScan(inputObj *Scan) {
	var obj Scan
	if inputObj.ID != 0 {
		DB.Model(Scan{}).First(&obj, inputObj.ID)
		if obj.ID != 0 {
			DB.Model(&obj).Updates(inputObj)
			inputObj.ID = obj.ID
			return
		}
	}

	//
	//DB.Where(Scan{InputName: inputObj.InputName}).First(&obj)
	//if obj.ID != 0 {
	//	DB.Model(&obj).Updates(inputObj)
	//	inputObj.ID = obj.ID
	//	return
	//}

	tx := DB.Create(&inputObj)
	if tx.Error != nil {
		utils.ErrorF("error creating target -- %v", tx.Error)
	}
	utils.DebugF("[DB] Creating new scan with ID: %v", inputObj.ID)
}

func DBUpdateScan(inputObj *Scan) {
	var obj Scan
	DB.Where(Scan{InputName: inputObj.InputName}).First(&obj)
	if obj.ID != 0 {
		DB.Model(&obj).Updates(inputObj)
		inputObj.ID = obj.ID
		utils.DebugF("[DB] Updating scan with ID: %v -- %v", obj.InputName, obj.ID)
		return
	}
	DB.Create(&inputObj)
	utils.DebugF("[DB] Creating new scan with ID: %v", inputObj.ID)
}

func DBUpdateTarget(inputObj *Target) {
	var obj Target

	DB.Where(Target{InputName: inputObj.InputName}).First(&obj)
	if obj.ID != 0 {
		utils.DebugF("[DB] Updating target with ID: %v -- %v", inputObj.InputName, obj.ID)
		inputObj.IsNew = false
		DB.Model(&obj).Updates(inputObj)
		inputObj.ID = obj.ID
		return
	}

	err := DB.Create(&inputObj).Error
	if err != nil {
		utils.ErrorF("error creating target -- %v", err)
	}
	utils.DebugF("[DB] Creating new target with ID: %v -- %v", inputObj.InputName, inputObj.ID)
}

////// mics stuff only exist in osm version

//// NewOrg import new scan to DB
//func NewOrg(inputObj *Org) {
//	var obj Org
//
//	DB.Where(Org{Name: inputObj.Name}).First(&obj)
//	if obj.ID != 0 {
//		utils.DebugF("[DB] Creating new org with ID: %v -- %v", inputObj.Name, obj.ID)
//		// @TODO: change target to old
//		DB.Model(&obj).Updates(inputObj)
//		inputObj.ID = obj.ID
//
//		return
//	}
//	DB.Create(&inputObj)
//}

// NewReport import new scan to DB
func NewReport(inputObj *Report) {
	var obj Report
	DB.Where(Report{ReportName: inputObj.ReportName}).First(&obj)
	if obj.ID != 0 {
		utils.DebugF("[DB] Updating report with ID: %v -- %v", inputObj.ReportName, obj.ID)
		// @TODO: change target to old
		DB.Model(&obj).Updates(inputObj)
		inputObj = &obj
		return
	}
	DB.Create(&inputObj)
	utils.DebugF("[DB] Created report with ID: %v -- %v", inputObj.ReportName, inputObj.ID)
}

func DBNewReports(module libs.Module, targetObj *Target) {
	var reports []string
	reports = append(reports, module.Report.Final...)
	reports = append(reports, module.Report.Noti...)
	reports = append(reports, module.Report.Diff...)

	for _, report := range reports {

		// @NOTE: just create a record, UI will take care if file exist or not
		//if !utils.FileExists(report) {
		//	continue
		//}

		reportObj := Report{
			ReportName:  path.Base(report),
			ModulePath:  module.ModulePath,
			Module:      module.Name,
			ReportPath:  report,
			ReportType:  "",
			TargetRefer: targetObj.ID,
		}

		NewReport(&reportObj)
		targetObj.Reports = append(targetObj.Reports, reportObj)
	}
}
