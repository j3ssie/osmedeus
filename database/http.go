package database

import (
	"fmt"
	"github.com/j3ssie/osmedeus/utils"
	"strings"
)

func (h *HTTP) CreateHTTP() error {
	domain, err := utils.GetDomain(h.URL)
	if err != nil {
		return fmt.Errorf("error import http record")
	}

	var assetObj Asset
	DB.Model(Asset{}).Preload("Dns").Preload("Http").Where("asset_value = ?", domain).First(&assetObj)
	if assetObj.ID == 0 {
		assetObj = Asset{
			AssetValue:  domain,
			Score:       0,
			IsAlive:     true,
			ScanRefer:   h.ScanRefer,
			TargetRefer: h.TargetRefer,
		}
		err := DB.Create(&assetObj).Error
		if err != nil {
			utils.ErrorF("error to find asset with http record: %v -- %v", domain, err)
		}
	}

	h.UpdateDnsPort(&assetObj)
	h.Domain = domain
	h.AssetRefer = assetObj.ID
	var httpObj HTTP
	//var oldContent string

	DB.Model(HTTP{}).Where("url = ?", h.URL).First(&httpObj)
	if httpObj.ID == 0 {
		DB.Create(h)
		httpObj.ID = h.ID
	}

	if httpObj.ID != 0 {
		h.ScreenShotData = httpObj.ScreenShotData
	}

	h.ID = httpObj.ID
	// update anything here if changes happened
	if httpObj.Checksum != "" && h.Checksum != httpObj.Checksum {
		h.HasChanged = true
	}
	DB.Save(h)

	assetObj.IsAlive = true
	DB.Save(&assetObj)

	h.CreateNoti(httpObj)
	return nil
}

func (h *HTTP) CreateScreenShot() error {
	domain, err := utils.GetDomain(h.URL)
	if err != nil {
		return fmt.Errorf("error import http record")
	}

	var assetObj Asset
	DB.Model(Model{}).Preload("Http").Where("asset_value = ?", domain).First(&assetObj)
	if assetObj.ID == 0 {
		assetObj = Asset{
			AssetValue:  domain,
			Score:       0,
			IsAlive:     true,
			ScanRefer:   h.ScanRefer,
			TargetRefer: h.TargetRefer,
		}
		err := DB.Create(&assetObj).Error
		if err != nil {
			//utils.ErrorF("error to find asset with http record: %v -- %v", domain, err)
			utils.ErrorF("error to find asset with http screenshot record -- %v", domain)
		}
		//return fmt.Errorf("error to find asset with http screenshot record")
	}

	h.UpdateDnsPort(&assetObj)
	h.AssetRefer = assetObj.ID
	h.Domain = domain
	h.HasChanged = false

	var httpObj HTTP
	DB.Model(HTTP{}).Where("url = ?", h.URL).First(&httpObj)
	if httpObj.ID == 0 {
		DB.Create(h)
		httpObj.ID = h.ID
	}

	// save with old http content
	h.HTTPContent = httpObj.HTTPContent
	h.ID = httpObj.ID
	DB.Save(h)

	assetObj.IsAlive = true
	DB.Save(&assetObj)

	utils.DebugF("New Screenshot record ID:%v -- assetID:%v", h.ID, assetObj.ID)
	return nil
}

func (h *HTTP) UpdateDnsPort(assetObj *Asset) {
	//utils.DebugF("Update asset --> dns port for %v -- %v : %v", dns.ID, dns.DnsValue, dns.Ports)

	for _, dns := range assetObj.Dns {
		if strings.HasPrefix(h.URL, "http://") {
			dns.UpdateHTTPPort("80/http")
		} else {
			dns.UpdateHTTPPort("443/https")
		}
		//utils.DebugF("Update asset --> dns port for %v -- %v : %v", dns.ID, dns.DnsValue, dns.Ports)
		DB.Save(&dns)
		DB.Save(assetObj)
	}
}
