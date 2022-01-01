package core

import (
    "fmt"
    "github.com/Jeffail/gabs/v2"
    "github.com/j3ssie/osmedeus/database"
    "github.com/j3ssie/osmedeus/utils"
    jsoniter "github.com/json-iterator/go"
    "github.com/spf13/cast"
    "gorm.io/datatypes"
    "net/url"
    "strings"
)

func (r *Runner) StartScanNoti() error {
    data, err := jsoniter.Marshal(&r.ScanObj)
    if err != nil {
        return fmt.Errorf("err marshal object: %v", err)
    }
    noti := database.Notification{
        NotificationType:   "start",
        NotificationSource: "scan",
        NewData:            datatypes.JSON(data),
        ObjRefer:           r.ScanObj.ID,
        ScanRefer:          r.ScanObj.ID,
        TargetRefer:        r.TargetObj.ID,
    }
    return noti.Create()
}

func (r *Runner) ScanDoneNoti() error {
    data, err := jsoniter.Marshal(&r.ScanObj)
    if err != nil {
        return fmt.Errorf("err marshal object: %v", err)
    }
    noti := database.Notification{
        NotificationType:   "done",
        NotificationSource: "scan",
        NewData:            datatypes.JSON(data),
        ObjRefer:           r.ScanObj.ID,
        ScanRefer:          r.ScanObj.ID,
        TargetRefer:        r.TargetObj.ID,
    }
    return noti.Create()
}

// ImportLinks import new link to DB
func (r *Runner) ImportLinks(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }

    content := utils.ReadingLines(src)
    for _, line := range content {
        if strings.TrimSpace(line) == "" {
            continue
        }
        // {"input":"http://academy.live-test.shopee.io","source":"body","type":"url","output":"http://academy.live-test.shopee.io","status":307,"length":7}
        jsonParsed, err := gabs.ParseJSON([]byte(line))
        if err != nil {
            utils.ErrorF("Error parse JSON Data")
            continue
        }

        obj := database.Link{
            LinkValue:   jsonParsed.S("output").Data().(string),
            LinkSource:  jsonParsed.S("source").Data().(string),
            URL:         jsonParsed.S("input").Data().(string),
            LinkType:    jsonParsed.S("type").Data().(string),
            ScanRefer:   r.ScanObj.ID,
            TargetRefer: r.TargetObj.ID,
        }
        obj.Create()
    }
}

func (r *Runner) ImportArchive(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }

    content := utils.ReadingLines(src)
    for _, line := range content {
        if strings.TrimSpace(line) == "" {
            continue
        }

        // ignore root domain
        u, err := url.Parse(line)
        if err != nil || u.Path == "" || u.Path == "/" {
            continue
        }

        obj := database.Archive{
            ArchiveValue:    line,
            ArchiveChecksum: utils.GenHash(line),
            ScanRefer:       r.ScanObj.ID,
            TargetRefer:     r.TargetObj.ID,
        }
        obj.Create()
    }
}

func (r *Runner) ImportIPRange(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }

    content := utils.ReadingLines(src)
    for _, line := range content {
        if strings.TrimSpace(line) == "" {
            continue
        }
        jsonParsed, err := gabs.ParseJSON([]byte(line))
        if err != nil {
            utils.ErrorF("Error parse JSON Data")
            continue
        }

        utils.DebugF("Processing: %v", line)
        obj := database.IPRange{
            ASNumber:    cast.ToString(jsonParsed.S("Number").Data()),
            Country:     cast.ToString(jsonParsed.S("CountryCode").Data()),
            Value:       cast.ToString(jsonParsed.S("CIDR").Data()),
            Info:        cast.ToString(jsonParsed.S("Description").Data()),
            Amount:      cast.ToUint(jsonParsed.S("Count").Data()),
            ScanRefer:   r.ScanObj.ID,
            TargetRefer: r.TargetObj.ID,
        }
        obj.Create()
    }
}

func (r *Runner) ImportCert(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }

    content := utils.ReadingLines(src)
    for _, line := range content {
        if strings.TrimSpace(line) == "" {
            continue
        }
        jsonParsed, err := gabs.ParseJSON([]byte(line))
        if err != nil {
            utils.ErrorF("Error parse JSON Data")
            continue
        }

        domain := cast.ToString(jsonParsed.S("Domain").Data())
        isWildCard := false
        if strings.Contains(domain, "*.") {
            isWildCard = true
            domain = strings.TrimLeft(domain, "*.")
        }

        obj := database.CertInfo{
            Domain:      domain,
            CertInfo:    cast.ToString(jsonParsed.S("CertInfo").Data()),
            OrgInfo:     cast.ToString(jsonParsed.S("OrgInfo").Data()),
            IsWildcard:  isWildCard,
            TargetRefer: r.TargetObj.ID,
        }
        obj.Create()
    }
}

func (r *Runner) ImportCred(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }

    content := utils.ReadingLines(src)
    for _, line := range content {
        if strings.TrimSpace(line) == "" {
            continue
        }

        jsonParsed, err := gabs.ParseJSON([]byte(line))
        if err != nil {
            utils.ErrorF("Error parse JSON Data")
            continue
        }

        email := cast.ToString(jsonParsed.S("email").Data())
        credID := cast.ToString(jsonParsed.S("id").Data())
        username := cast.ToString(jsonParsed.S("username").Data())
        password := cast.ToString(jsonParsed.S("password").Data())
        hashedPassword := cast.ToString(jsonParsed.S("hashed_password").Data())
        name := cast.ToString(jsonParsed.S("name").Data())
        ipAddress := cast.ToString(jsonParsed.S("ip_address").Data())
        phone := cast.ToString(jsonParsed.S("phone").Data())
        source := cast.ToString(jsonParsed.S("database_name").Data())

        obj := database.Credential{
            CredID:         credID,
            Email:          email,
            Username:       username,
            Password:       password,
            HashedPassword: hashedPassword,
            Name:           name,
            Phone:          phone,
            IPAddress:      ipAddress,
            Source:         source,
            TargetRefer:    r.TargetObj.ID,
            ScanRefer:      r.ScanObj.ID,
        }
        obj.Create()
    }
}

func (r *Runner) ImportCloudBrute(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }

    content := utils.ReadingLines(src)
    for _, line := range content {
        if strings.TrimSpace(line) == "" {
            continue
        }

        if !strings.Contains(line, " - ") {
            continue
        }

        cloudDomain := strings.Split(line, " - ")[1]
        status := strings.Split(line, " - ")[0]
        if strings.Contains(status, ": ") {
            status = strings.Split(strings.Split(line, " - ")[0], ": ")[1]
        }

        obj := database.CloudBrute{
            Status:      status,
            CloudDomain: cloudDomain,
            RawData:     line,
            ScanRefer:   r.ScanObj.ID,
            TargetRefer: r.TargetObj.ID,
        }
        obj.Create()
    }
}
