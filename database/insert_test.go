package database

import (
    "fmt"
    "github.com/davecgh/go-spew/spew"
    jsoniter "github.com/json-iterator/go"
    "github.com/spf13/cast"
    "os"
    "testing"
)

func TestInsertChange(t *testing.T) {
    setup()
    //ClearDB()
    t.Log("Create record ...")
    //tx := DB.Begin()

    asset := Asset{
        AssetValue: "academy.live-test.shopee.io",
        Score:      0,
        IsAlive:    false,
        ScanRefer:  1,
    }
    asset.Create()

    //var assetObj Asset
    dnsObj := Dns{
        DnsValue:  "live-test.shopee.com",
        DnsType:   "CNAME",
        Domain:    "academy.live-test.shopee.io",
        ScanRefer: 1,
    }
    dnsObj.Create()

    asset.Dns = append(asset.Dns, dnsObj)

    var nDnsObj Dns
    DB.Where("domain = ?", "academy.live-test.shopee.io").First(&nDnsObj)

    nDnsObj.DnsType = "A"
    nDnsObj.DnsValue = "1.2.3.4"
    DB.Save(&nDnsObj)

    //asset.Dns = append(asset.Dns, nDnsObj)
    //DB.Save(&asset)

    t.Log("\n\n--- Querying target ...")
    //	var d []Dns
    //DB.Limit(1).Find(&d)
    //spew.Dump(d)

    var a []Asset
    DB.Preload("Dns").Limit(1).Find(&a)
    spew.Dump(a)
}

func TestInsertHTTP(t *testing.T) {
    setup()
    //ClearDB()
    t.Log("Create record ...")
    //tx := DB.Begin()

    asset := Asset{
        AssetValue: "academy.live-test.shopee.io",
        Score:      0,
        IsAlive:    false,
        ScanRefer:  1,
    }
    asset.Create()

    //var assetObj Asset
    httpObj := HTTP{
        URL:            "https://academy.live-test.shopee.io",
        Redirect:       "",
        Title:          "Title",
        Checksum:       "xxx",
        StatusCode:     200,
        ContentLength:  789,
        ScreenShotData: "sampledata",
        HTTPContent:    "",
        HasChanged:     false,
        //AssetRefer:     asset,
        ScanRefer: 1,
    }
    httpObj.CreateScreenShot()

    asset.HTTP = append(asset.HTTP, httpObj)
    DB.Save(&asset)

    t.Log("\n\n--- Querying target ...")

    var a []Asset
    DB.Table("Assets").Preload("Http").Limit(1).Find(&a)
    spew.Dump(a)
}

func TestGetAssetID(t *testing.T) {
    setup()
    var assetID = os.Getenv("ASSET_ID")

    var a Asset
    DB.Preload("Vulnerability").Preload("Dns").Preload("Link").Preload("Archive").Preload("Http").Preload("Directory").First(&a, assetID)
    //DB.Table("Assets").Preload("Http").Preload("Dns").First(&a, assetID)

    fmt.Println("assetID -->", assetID)
    //tx.First(&a, assetID)
    spew.Dump(a)

}

func TestCountModels(t *testing.T) {
    setup()

    target := Target{
        Model: Model{
            ID: 1,
        },
    }

    scan := Scan{
        Model: Model{
            ID: 1,
        },
        Target:      target,
        TargetRefer: 1,
    }

    var sum int64
    DB.Table("Assets").Where("scan_refer = ?", scan.Model.ID).Count(&sum)
    target.TotalAssets += cast.ToInt(sum)

    DB.Table("ip_ranges").Where("target_refer = ?", scan.TargetRefer).Count(&sum)
    target.TotalIPRange += cast.ToInt(sum)

    scan.SummaryTarget()
    spew.Dump(target)
}

func TestNotiInsert(t *testing.T) {
    setup()
    var a Asset

    //assetID := "1"
    //DB.Preload("Vulnerability").Preload("Dns").Preload("Link").Preload("Archive").First(&a, assetID)
    //
    //data, _ := jsoniter.Marshal(&a)
    //
    //DB.Create(&Notification{
    //	NotificationType:   "new",
    //	NotificationSource: "asset",
    //	OldData:            datatypes.JSON(data),
    //	NewData:            nil,
    //	AssetRefer:         1,
    //})

    var noti Notification
    DB.Table("Notifications").First(&noti)
    //spew.Dump(noti)
    spew.Dump(noti.OldData)

    jsoniter.Unmarshal(noti.OldData, &a)

    spew.Dump(a)
}
