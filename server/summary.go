package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/database"
)

type DefaultQuery struct {
	ID     string `query:"id"`
	Value  string `query:"value"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
}

func GetTarget(c *fiber.Ctx) error {
	query := new(DefaultQuery)
	if err := c.QueryParser(query); err != nil {
		return err
	}

	tx := DB.Table("Targets").Preload("Scan")
	if query.Limit > 0 {
		tx.Limit(query.Limit)
	}

	var objs []database.Target
	tx.Find(&objs)

	return c.JSON(ResponseHTTP{
		Status:  200,
		Data:    objs,
		Type:    "targets",
		Total:   len(objs),
		Message: "Get targets",
	})
}

func GetNoti(c *fiber.Ctx) error {
	query := new(DefaultQuery)
	if err := c.QueryParser(query); err != nil {
		return err
	}

	tx := DB.Table("Notifications")
	if query.Limit > 0 {
		tx.Limit(query.Limit)
	}

	var objs []database.Notification
	tx.Find(&objs)

	return c.JSON(ResponseHTTP{
		Status:  200,
		Data:    objs,
		Type:    "notifications",
		Total:   len(objs),
		Message: "Get Notifications",
	})
}

func GetScan(c *fiber.Ctx) error {
	query := new(DefaultQuery)
	if err := c.QueryParser(query); err != nil {
		return err
	}

	tx := DB.Table("Scans").Preload("Target")
	if query.Limit > 0 {
		tx.Limit(query.Limit)
	}

	var objs []database.Scan
	tx.Find(&objs)

	return c.JSON(ResponseHTTP{
		Status:  200,
		Data:    objs,
		Type:    "scans",
		Total:   len(objs),
		Message: "Get Scan Data",
	})
}

func GetAsset(c *fiber.Ctx) error {
	query := new(DefaultQuery)
	if err := c.QueryParser(query); err != nil {
		return err
	}

	scanID := c.Query("scanId", "0")
	if scanID == "" {
		return c.JSON(ResponseHTTP{
			Status:  500,
			Type:    "assets",
			Message: "Empty scan ID",
		})
	}

	tx := DB.Table("Assets").Preload("Dns").Preload("Http").Preload("Directory").Where("scan_refer = ?", scanID)
	//tx := DB.Table("Assets").Preload("Dns").Preload("Vulnerability").Preload("Link").Preload("Archive").Preload("Http").Preload("Directory").Where("scan_refer = ?", scanID)

	if query.Limit > 0 {
		tx.Limit(query.Limit)
	}
	if query.Offset > 0 {
		tx.Offset(query.Offset)
	}

	var objs []database.Asset
	tx.Find(&objs)

	return c.JSON(ResponseHTTP{
		Status:  200,
		Data:    objs,
		Type:    "assets",
		Total:   len(objs),
		Message: "List Assets",
	})
}

func GetAssetDetail(c *fiber.Ctx) error {
	query := new(DefaultQuery)
	if err := c.QueryParser(query); err != nil {
		return err
	}

	//scanID := c.Query("scanId", "0")
	if query.Value == "" {
		return c.JSON(ResponseHTTP{
			Status:  500,
			Type:    "assets",
			Message: "Empty scan ID or asset value",
		})
	}

	var obj database.Asset
	//DB.Table("Assets").Preload("Dns").Preload("Vulnerability").Preload("Link").Preload("Archive").Preload("Http").Preload("Directory").Where("asset_value = ?", query.Value).First(&obj)
	DB.Table("Assets").Preload("Vulnerability").Preload("Dns").Where("asset_value = ?", query.Value).First(&obj)
	//tx := DB.Table("Assets").Preload("Dns").Preload("Vulnerability").Preload("Link").Preload("Archive").Preload("Http").Preload("Directory").Where("scan_refer = ?", scanID)

	//if query.Value != "" {
	//	tx.Where("asset_value = ?", query.Value)
	//}
	//
	//tx.First(&objs)
	//if query.ID != "" {
	//	tx.First(&objs, query.ID)
	//}

	return c.JSON(ResponseHTTP{
		Status:  200,
		Data:    obj,
		Type:    "asset",
		Message: "Asset Detail",
	})
}

//// Target belong record

func GetIPRange(c *fiber.Ctx) error {
	var objs []database.IPRange
	DB.Find(&objs)

	return c.JSON(ResponseHTTP{
		Status:  200,
		Data:    objs,
		Total:   len(objs),
		Type:    "ip_ranges",
		Message: "List IPRanges",
	})
}

func GetHTTP(c *fiber.Ctx) error {
	query := new(DefaultQuery)
	if err := c.QueryParser(query); err != nil {
		return err
	}

	scanID := c.Query("scanId", "0")
	if scanID == "" {
		return c.JSON(ResponseHTTP{
			Status:  500,
			Type:    "assets",
			Message: "Empty scan ID",
		})
	}

	var objs []database.HTTP
	tx := DB.Where("scan_refer = ?", scanID)

	if query.Limit > 0 {
		tx.Limit(query.Limit)
	}
	if query.Offset > 0 {
		tx.Offset(query.Offset)
	}
	tx.Find(&objs)

	return c.JSON(ResponseHTTP{
		Status:  200,
		Data:    objs,
		Total:   len(objs),
		Type:    "cloud_brutes",
		Message: "List CloudBrute",
	})
}

func GetCloudBrute(c *fiber.Ctx) error {
	query := new(DefaultQuery)
	if err := c.QueryParser(query); err != nil {
		return err
	}

	var objs []database.CloudBrute
	DB.Find(&objs)

	return c.JSON(ResponseHTTP{
		Status:  200,
		Data:    objs,
		Total:   len(objs),
		Type:    "cloud_brutes",
		Message: "List CloudBrute",
	})
}

func GetCredential(c *fiber.Ctx) error {
	query := new(DefaultQuery)
	if err := c.QueryParser(query); err != nil {
		return err
	}

	var objs []database.Credential
	DB.Find(&objs)

	return c.JSON(ResponseHTTP{
		Status:  200,
		Data:    objs,
		Type:    "credentials",
		Total:   len(objs),
		Message: "List Credentials",
	})
}
