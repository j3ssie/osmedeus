package database

import (
    "fmt"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/thoas/go-funk"
    "strings"
)

func (a *Asset) Create() {
    if DB.Model(Asset{}).Where("asset_value = ?", a.AssetValue).RowsAffected == 0 {
        DB.Create(a)
    }
}

func (a *Asset) UpdateTech() error {
    var assetObj Asset
    DB.Model(Asset{}).Where("asset_value = ?", a.AssetValue).First(&assetObj)
    if assetObj.ID == 0 {
        a.IsAlive = true
        err := DB.Create(&a).Error
        if err != nil {
            utils.ErrorF("err create asset with tech -- %v:%v -- %v", a.AssetValue, a.Technology, err)
        }
        return nil
    }

    tech := a.Technology
    if assetObj.Technology == tech {
        return fmt.Errorf("duplicate record")
    }

    var finalTech []string
    if strings.Contains(a.Technology, ",") {
        finalTech = strings.Split(a.Technology, ",")
    } else if !strings.Contains(a.Technology, "N/A") {
        finalTech = append(finalTech, a.Technology)
    }

    if strings.Contains(tech, ",") {
        finalTech = append(finalTech, strings.Split(tech, ",")...)
    } else {
        finalTech = append(finalTech, tech)
    }
    finalTech = funk.UniqString(finalTech)

    // update the record
    assetObj.IsAlive = true
    assetObj.Technology = strings.Trim(strings.Join(finalTech, ","), ",")
    DB.Save(&assetObj)
    return fmt.Errorf("new-technology")
}

func (d *Dns) Create() error {
    //var err error
    var assetObj Asset
    tx := DB.Begin()

    DB.Model(Asset{}).Where("asset_value = ?", d.Domain).First(&assetObj)
    if assetObj.ID == 0 {
        assetObj = Asset{
            AssetValue:  d.Domain,
            Dns:         []Dns{*d},
            Score:       0,
            IsAlive:     true,
            ScanRefer:   d.ScanRefer,
            TargetRefer: d.TargetRefer,
        }
        DB.Create(&assetObj)
    }

    d.AssetRefer = assetObj.ID
    // create new one if we didn't find it
    if DB.Table("Dns").Where("dns_checksum = ?", d.DnsChecksum).RowsAffected == 0 {
        DB.Create(d)
    }

    assetObj.IsAlive = true
    assetObj.Dns = append(assetObj.Dns, *d)
    DB.Save(&assetObj)

    return tx.Commit().Error
}

func (d *Dns) UpdatePort() error {
    var dnsObj Dns
    DB.Model(Dns{}).Where("dns_value = ? AND dns_type = ?", d.DnsValue, d.DnsType).First(&dnsObj)
    if dnsObj.ID == 0 {
        d.DnsChecksum = utils.GenHash(fmt.Sprintf("%s-%s-%s", d.Domain, d.DnsType, d.DnsValue))
        err := DB.Create(d).Error
        if err != nil {
            utils.ErrorF("ip with no asset refer --  %v:%v -- %v", d.Domain, d.DnsValue, err)
            return fmt.Errorf("ip with no asset refer")
        }
    }

    dnsObj.Ports = d.Ports
    var assetObj Asset
    DB.Model(Asset{}).Preload("Dns").Where("asset_value = ?", d.Domain).First(&assetObj)

    if len(assetObj.HTTP) > 0 {
        dnsObj.UpdateHTTPPort()
    }

    err := DB.Save(&dnsObj).Error
    if err != nil {
        utils.ErrorF("err  %v:%v -- %v", dnsObj.ID, dnsObj.Domain, err)
    } else {
        utils.DebugF("Update ports for %v:%v - %v", dnsObj.ID, dnsObj.Domain, dnsObj.Ports)

    }

    if dnsObj.Ports == "" {
        return fmt.Errorf("new-port")
    }
    return fmt.Errorf("change-port")
}

func (d *Dns) UpdateHTTPPort(portValue ...string) {
    if len(portValue) == 0 {
        portValue = []string{"80/http", "443/https"}
    }

    if d.Ports == "" {
        d.Ports = strings.Join(portValue, ",")
        return
    }

    if !strings.Contains(d.Ports, ",") {
        portValue = append(portValue, d.Ports)
        portValue = funk.UniqString(portValue)
        d.Ports = strings.Join(portValue, ",")
        return
    }

    // uniq with previous one
    utils.DebugF("Update ports for %v:%v - %v", d.ID, d.Domain, d.Ports)
    rawPorts := strings.Split(d.Ports, ",")
    rawPorts = append(rawPorts, portValue...)
    utils.DebugF("rawPorts: --> %v", rawPorts)
    rawPorts = funk.UniqString(rawPorts)

    d.Ports = strings.Join(rawPorts, ",")
    DB.Save(d)
}
