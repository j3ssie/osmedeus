package database

import (
    "github.com/j3ssie/osmedeus/utils"
    "gorm.io/gorm/clause"
)

// ClearDB clean all scans
func ClearDB() {
    ClearTable()
}

func ClearTable(tableName ...string) {
    tableNames := []string{
        "Scan",
        "Report",
        "Target",
        "Asset",
        "Dns",
        "HTTP",
        "Vulnerability",
        "Notification",
        "Directory",
        "Link",
        "Credential",
        "Archive",
        "CloudBrute",
        "IPRange",
        "Org",
    }

    if len(tableName) > 0 {
        tableNames = tableName
    }

    for _, table := range tableNames {
        switch table {
        case "Scan":
            var obj []Scan
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)
        case "Report":
            var obj []Report
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)
        case "Target":
            var obj []Target
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)

        //case "Org":
        //	var obj []Org
        //	DB.Find(&obj)
        //	DB.Unscoped().Delete(&obj)

        // asset data
        case "Asset":
            var obj []Asset
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)
        case "Dns":
            var obj []Dns
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)
        case "HTTP":
            var obj []HTTP
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)
        case "Vulnerability":
            var obj []Vulnerability
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)
        case "Directory":
            var obj []Directory
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)
        case "Notification":
            var obj []Notification
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)

        case "Link":
            var obj []Link
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)

        case "Archive":
            var obj []Archive
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)

        case "IPRange":
            var obj []IPRange
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)

        case "Credential":
            var obj []Credential
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)

        case "CloudBrute":
            var obj []CloudBrute
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)

        default:
            var obj []Asset
            DB.Find(&obj)
            DB.Unscoped().Delete(&obj)
        }
        utils.InforF("Clear Table %v", table)

    }
}

func CleanWorkspace(wsname string) {
    utils.InforF("Deleted records for: %v", wsname)
    var target Target
    DB.Preload(clause.Associations).Preload("Reports").First(&target, "workspace = ?", wsname)

    DB.Unscoped().Delete(&target)

    var scan Scan
    DB.Preload(clause.Associations).Preload("Targets").First(&scan, "target_refer = ?", target.ID)
    DB.Unscoped().Delete(&scan)

    var reports []Report
    DB.Preload(clause.Associations).Preload("Targets").Find(&reports, "target_refer = ?", target.ID)
    DB.Unscoped().Delete(&reports)

}
