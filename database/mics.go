package database

import (
    "fmt"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/thoas/go-funk"
)

func (i *IPRange) Create() error {
    i.IPRangeChecksum = utils.GenHash(fmt.Sprintf("%s-%s-%d", i.ASNumber, i.Value, i.TargetRefer))

    var ipr IPRange
    DB.Model(IPRange{}).Where("iprange_checksum = ?", i.IPRangeChecksum).First(&ipr)
    if ipr.ID == 0 {
        DB.Create(i)
        return nil
    }

    if i.Amount > 0 {
        ipr.Amount += i.Amount
        DB.Save(&ipr)
        return fmt.Errorf("new-amount")
    }
    return nil
}

func (c *CertInfo) Create() error {
    c.CertChecksum = utils.GenHash(fmt.Sprintf("%s-%s-%s-%d", c.CertInfo, c.OrgInfo, c.Domain, c.TargetRefer))

    var ipr IPRange
    DB.Model(IPRange{}).Where("cert_checksum = ?", c.CertChecksum).First(&ipr)
    if ipr.ID == 0 {
        DB.Create(c)
    }
    return nil
}

func (i *Credential) Create() error {
    i.CredChecksum = utils.GenHash(fmt.Sprintf("%s-%s-%d", i.Email, i.CredID, i.TargetRefer))

    var cred Credential
    DB.Model(Credential{}).Where("cred_checksum = ?", i.CredChecksum).First(&cred)

    if cred.ID == 0 {
        DB.Create(i)
        return fmt.Errorf("new-cred")
    }

    return nil
}

func (i *CloudBrute) Create() error {
    tx := DB.Model(CloudBrute{}).Where("cloud_domain = ?", i.CloudDomain)
    if tx.RowsAffected == 0 {
        DB.Create(i)
        return fmt.Errorf("new-cloud")
    }

    return nil
}

func (l *Link) Create() error {
    domain, err := utils.GetDomain(l.URL)
    if err != nil {
        return fmt.Errorf("error import http record")
    }
    var assetObj Asset
    DB.Model(Asset{}).Preload("Https").Where("asset_value = ?", domain).First(&assetObj)
    if assetObj.ID == 0 {
        return fmt.Errorf("error import http record")
    }

    l.LinkChecksum = utils.GenHash(fmt.Sprintf("%s-%s-%d", l.LinkType, l.LinkValue, l.ScanRefer))
    l.AssetRefer = assetObj.ID

    var linkObj Link
    DB.Model(Link{}).Where("link_checksum = ?", l.LinkChecksum).First(&linkObj)

    if linkObj.ID == 0 {
        DB.Create(l)
        linkObj.ID = l.ID
    }

    l.ID = linkObj.ID
    assetObj.Link = append(assetObj.Link, *l)
    assetObj.Link = funk.Uniq(assetObj.Link).([]Link)
    DB.Save(&assetObj)
    return nil
}

func (l *Archive) Create() error {
    domain, err := utils.GetDomain(l.ArchiveValue)
    if err != nil {
        return fmt.Errorf("error import http record")
    }

    var assetObj Asset
    DB.Model(Asset{}).Preload("Https").Where("asset_value = ?", domain).First(&assetObj)
    if assetObj.ID == 0 {
        return fmt.Errorf("error import http record")
    }

    l.AssetRefer = assetObj.ID
    var linkObj Archive
    tx := DB.Model(Archive{}).Where("archive_checksum = ?", l.ArchiveChecksum).First(&linkObj)
    if tx.RowsAffected == 0 {
        DB.Create(l)
        linkObj.ID = l.ID
    }

    l.ID = linkObj.ID

    assetObj.Archive = append(assetObj.Archive, *l)
    assetObj.Archive = funk.Uniq(assetObj.Archive).([]Archive)

    DB.Save(&assetObj)
    return nil
}
