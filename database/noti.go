package database

import (
	"fmt"
	"github.com/j3ssie/osmedeus/utils"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/datatypes"
)

type Notification struct {
	Model
	// new / change
	NotificationType string `json:"notification_type"`
	// asset / target / dns
	NotificationSource   string `json:"notification_source"`
	NotificationChecksum string `json:"notification_checksum"`

	// raw data
	NewData datatypes.JSON `json:"new_data"`
	OldData datatypes.JSON `json:"old_data"`

	IsPlugin bool `json:"is_plugin"`
	Sent     bool `json:"sent"`

	ObjRefer    uint `gorm:"index:idx_obj" json:"obj_refer"`
	TargetRefer uint `gorm:"index:idx_id" json:"target_refer"`
	AssetRefer  uint `gorm:"index:idx_id" json:"asset_refer"`
	ScanRefer   uint `json:"scan_refer"`
}

func (n *Notification) Create() error {
	checksum := utils.GenHash(fmt.Sprintf(fmt.Sprintf("%v-%v-%v-%v-%v", n.NotificationType, n.NotificationSource, n.ObjRefer, n.AssetRefer, n.TargetRefer)))
	n.NotificationChecksum = checksum

	var notiObj Notification
	DB.Table("Notifications").Where("notification_checksum = ?", n.NotificationChecksum).First(&notiObj)
	if notiObj.ID == 0 {
		tx := DB.Create(&n)
		return tx.Error
	}

	return fmt.Errorf("noti already exist")
}

// CreateNoti use this one for notification
func (v *Vulnerability) CreateNoti() (err error) {
	data, err := jsoniter.Marshal(&v)
	if err != nil {
		return fmt.Errorf("err marshal object: %v", err)
	}
	utils.DebugF("Creating noti for vuln: %v", v.ID)

	noti := Notification{
		NotificationType:   "new",
		NotificationSource: "vulnerability",
		NewData:            datatypes.JSON(data),
		//NewData:            nil,
		ObjRefer:    v.ID,
		AssetRefer:  v.AssetRefer,
		ScanRefer:   v.ScanRefer,
		TargetRefer: v.TargetRefer,
	}

	err = noti.Create()
	if err == nil {
		utils.DebugF("New Noti for Vuln: %v -- %v:%v", v.ID, noti.ID, noti.NotificationChecksum)
	}
	return err
}

// CreateNoti use this one for notification
func (h *HTTP) CreateNoti(oldObj HTTP) (err error) {
	if !h.HasChanged {
		return nil
	}

	if oldObj.ID == 0 {
		utils.ErrorF("no old object found")
		return fmt.Errorf("no old object found")
	}

	data, err := jsoniter.Marshal(&h)
	if err != nil {
		return fmt.Errorf("err marshal object: %v", err)
	}

	odata, err := jsoniter.Marshal(&oldObj)
	if err != nil {
		return fmt.Errorf("err marshal object: %v", err)
	}

	noti := Notification{
		NotificationType:   "change",
		NotificationSource: "HTTP",
		NewData:            datatypes.JSON(data),
		OldData:            datatypes.JSON(odata),
		ObjRefer:           h.ID,
		AssetRefer:         h.AssetRefer,
		ScanRefer:          h.ScanRefer,
		TargetRefer:        h.TargetRefer,
	}

	err = noti.Create()
	if err == nil {
		utils.DebugF("New Noti for HTTP: %v -- %v:%v", noti.ID, h.ID, h.Domain)
	} else {
		utils.ErrorF("%v", err)
	}
	return err
}

//
////AfterCreate use this one for notification
//func (v *Vulnerability) AfterCreate(tx *gorm.DB) (err error) {
//	data, err := jsoniter.Marshal(&v)
//	if err != nil {
//		return fmt.Errorf("err marshal object: %v", err)
//	}
//	utils.DebugF("Creating noti for vuln: %v", v.ID)
//
//
//	noti := Notification{
//		NotificationType:   "new",
//		NotificationSource: "vulnerability",
//		NewData:            datatypes.JSON(data),
//		//NewData:            nil,
//		ObjRefer:    v.ID,
//		AssetRefer:  v.AssetRefer,
//		ScanRefer:   v.ScanRefer,
//		TargetRefer: v.TargetRefer,
//	}
//
//	err = noti.Create()
//	if err == nil {
//		utils.DebugF("New Noti for Vuln: %v -- %v:%v", v.ID, noti.ID, noti.NotificationChecksum)
//	}
//	return err
//}
//
////BeforeUpdate use this one for notification
//func (h *HTTP) BeforeUpdate(tx *gorm.DB) (err error) {
//	var oldObj HTTP
//	if h.HasChanged {
//		return nil
//	}
//
//	tx.Model(&HTTP{}).Where("http_id = ?", h.ID).First(&oldObj)
//	if oldObj.ID == 0 {
//		utils.ErrorF("no old object found")
//		return fmt.Errorf("no old object found")
//	}
//
//	data, err := jsoniter.Marshal(&h)
//	if err != nil {
//		return fmt.Errorf("err marshal object: %v", err)
//	}
//
//	odata, err := jsoniter.Marshal(&oldObj)
//	if err != nil {
//		return fmt.Errorf("err marshal object: %v", err)
//	}
//
//	noti := Notification{
//		NotificationType:   "change",
//		NotificationSource: "HTTP",
//		NewData:            datatypes.JSON(data),
//		OldData:            datatypes.JSON(odata),
//		//NewData:            nil,
//		AssetRefer:  h.AssetRefer,
//		ScanRefer:   h.ScanRefer,
//		TargetRefer: h.TargetRefer,
//	}
//
//	err = noti.Create()
//	if err == nil {
//		utils.DebugF("New Noti for Vuln: %v", h.ID, noti.ID)
//	} else {
//		utils.ErrorF("%v", err)
//	}
//
//	return
//}
