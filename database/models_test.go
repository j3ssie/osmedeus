package database

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"testing"
)

func setup() {
	var err error

	var opt libs.Options
	opt.Server.DBPath = utils.NormalizePath("~/.osmedeus/sqlite.db")
	DB, err = InitDB(opt)
	if err != nil {
		panic(err)
	}
}

func setupRemote() {
	var err error

	var opt libs.Options
	//opt.Server.DBPath = utils.NormalizePath("~/.osmedeus/sqlite.db")
	opt.Server.DBType = "mysql"
	opt.Server.DBConnection = "doadmin:ekwk4flm93oucg52@tcp(hts-ng-do-user-6480911-0.b.db.ondigitalocean.com:25060)/hts-engine?charset=utf8&parseTime=True&loc=Local"
	DB, err = InitDB(opt)
	if err != nil {
		panic(err)
	}
}

func TestDnsModels(t *testing.T) {
	setup()
	t.Log("Create record ...")
	tx := DB.Begin()

	asset := Asset{
		AssetValue: "academy.live-test.redacted.io",
		Score:      0,
		IsAlive:    false,
		ScanRefer:  1,
	}
	asset.Create()

	var assetObj Asset

	dnsObj := Dns{
		DnsValue:  "live-test.shopee.com",
		DnsType:   "CNAME",
		Domain:    "academy.live-test.redacted.io",
		ScanRefer: 1,
	}

	t.Log("Querying asset ...")

	DB.Where("asset_value = ?", "academy.live-test.redacted.io").First(&assetObj)
	if assetObj.ID == 0 {
		assetObj = Asset{
			AssetValue: "academy.live-test.redacted.io",
			//Dns:        []Dns{*d},
			//Dns: dnsObj,
			Score:     0,
			IsAlive:   true,
			ScanRefer: 1,
		}
		tx := DB.Create(&assetObj)
		fmt.Printf("tx --> %v\n", tx.Error)
	} else {
		fmt.Printf("assetObj.ID --> %v\n", assetObj.ID)
	}

	dnsObj.AssetRefer = assetObj.ID
	if DB.Model(&dnsObj).Where("dns_value = ?", dnsObj.DnsValue).Updates(&dnsObj).RowsAffected == 0 {
		DB.Create(&dnsObj)
	}

	assetObj.IsAlive = true
	//assetObj.Dns = dnsObj

	DB.Save(&assetObj)
	//DB.Model(&assetObj).Association("Dns").Append([]Dns{*d})
	fmt.Printf("assetObj: %v -- %v\n", assetObj.ID, assetObj.AssetValue)
	tx = DB.Updates(&assetObj)
	fmt.Printf("tx --> %v\n", tx.Error)

	spew.Dump(assetObj)
	tx.Commit()

	//dnsObj.Create()
	///////
	t.Log("\n\n--- Querying target ...")
	var obj []Asset
	DB.Preload("Dns").Preload("Assets").Limit(1).Find(&obj)
	spew.Dump(obj)
}

func TestQueryModels(t *testing.T) {
	setup()

	//var assetObj Asset
	//tx:= DB.Preload("Dns.HTTP.Directory.Vulnerability.Link.Archive").Where("asset_value = ?", "academy.live-test.redacted.io").First(&assetObj)
	//fmt.Println("tx --> ", tx.RowsAffected, tx.Error)
	//var dnsObj Dns
	//DB.Where("dns_value = ?", "live-test.shopee.com").First(&dnsObj)
	//
	//assetObj.Dns =  append(assetObj.Dns, dnsObj)
	//DB.Save(&assetObj)
	//spew.Dump(assetObj)
	//

	var obj []Asset
	DB.Limit(5).Preload("Dns.Http.Directory.Vulnerability.Link.Archive").Find(&obj)
	spew.Dump(obj)
}

func TestSingleQueryModels(t *testing.T) {
	setup()

	var assetObj Asset

	//tx:= DB.Preload("Dns").Preload("Http").Where("asset_value = ?", "academy.live-test.redacted.io").First(&assetObj)
	DB.Table("Assets").Preload("Link").Preload("Http").Where("asset_value = ?", "dean.redacted.io").First(&assetObj)
	spew.Dump(assetObj)
}

func TestInsertPortModels(t *testing.T) {
	setup()
	data := `{"IPAddress":"103.247.206.155","Hostname":"","Ports":[{"Protocol":"tcp","PortID":"4001","State":"open","Service":{"Name":"http","Product":"nginx","Cpe":""},"Script":{"ID":"http-title","Output":"flutter_client"}},{"Protocol":"tcp","PortID":"4002","State":"open","Service":{"Name":"mlchat-proxy","Product":"","Cpe":""},"Script":{"ID":"fingerprint-strings","Output":"\n  FourOhFourRequest: \n    HTTP/1.0 403 Forbidden\n    Content-Type: text/plain; charset=utf-8\n    X-Content-Type-Options: nosniff\n    Date: Fri, 16 Jul 2021 16:52:11 GMT\n    Content-Length: 14\n    Forbidden\n  GenericLines, Help, Kerberos, LDAPSearchReq, LPDString, RTSPRequest, SSLSessionReq, TLSSessionReq, TerminalServerCookie: \n    HTTP/1.1 400 Bad Request\n    Content-Type: text/plain; charset=utf-8\n    Connection: close\n    Request\n  GetRequest, HTTPOptions: \n    HTTP/1.0 403 Forbidden\n    Content-Type: text/plain; charset=utf-8\n    X-Content-Type-Options: nosniff\n    Date: Fri, 16 Jul 2021 16:51:45 GMT\n    Content-Length: 14\n    Forbidden"}},{"Protocol":"tcp","PortID":"4003","State":"open","Service":{"Name":"pxc-splr-ft","Product":"","Cpe":""},"Script":{"ID":"fingerprint-strings","Output":"\n  GenericLines: \n    HTTP/1.1 400 Bad Request\n    Content-Type: text/plain; charset=utf-8\n    Connection: close\n    Request\n  GetRequest, HTTPOptions: \n    HTTP/1.0 200 OK\n    Accept-Ranges: bytes\n    Content-Length: 2558\n    Content-Type: text/html; charset=utf-8\n    Last-Modified: Wed, 31 Mar 2021 08:26:39 GMT\n    Date: Fri, 16 Jul 2021 16:51:45 GMT\n    \u003c!doctype html\u003e\u003chtml lang=\"en\"\u003e\u003chead\u003e\u003cmeta charset=\"utf-8\"/\u003e\u003cmeta name=\"viewport\" content=\"width=device-width,initial-scale=1\"/\u003e\u003cmeta name=\"msapplication-TileColor\" content=\"#da532c\"/\u003e\u003cmeta name=\"theme-color\" content=\"#ffffff\"/\u003e\u003cmeta name=\"description\" content=\"Web site created using create-react-app\"/\u003e\u003clink rel=\"manifest\" href=\"/manifest.json\"/\u003e\u003clink rel=\"apple-touch-icon\" sizes=\"180x180\" href=\"/apple-touch-icon.png\"/\u003e\u003clink rel=\"icon\" type=\"image/png\" sizes=\"32x32\" href=\"/favicon-32x32.png\"/\u003e\u003clink rel=\"icon\" type=\"image/png\" sizes=\"16x16\" href=\"/favicon-16x16.png\"/\u003e\u003clink rel=\"mask-icon\" href=\"/safari-pinned-tab.svg\" color=\"#5bbad5\"/\u003e\u003ctitle\u003eSeameet\u003c/title\u003e\u003clink href=\"/static/css/2.51964e6d.chunk.css\" rel="}},{"Protocol":"tcp","PortID":"4004","State":"open","Service":{"Name":"pxc-roid","Product":"","Cpe":""},"Script":{"ID":"fingerprint-strings","Output":"\n  GenericLines: \n    HTTP/1.1 400 Bad Request\n    Content-Type: text/plain; charset=utf-8\n    Connection: close\n    Request\n  GetRequest, HTTPOptions: \n    HTTP/1.0 200 OK\n    Accept-Ranges: bytes\n    Content-Length: 2558\n    Content-Type: text/html; charset=utf-8\n    Last-Modified: Mon, 12 Apr 2021 04:04:42 GMT\n    Date: Fri, 16 Jul 2021 16:51:45 GMT\n    \u003c!doctype html\u003e\u003chtml lang=\"en\"\u003e\u003chead\u003e\u003cmeta charset=\"utf-8\"/\u003e\u003cmeta name=\"viewport\" content=\"width=device-width,initial-scale=1\"/\u003e\u003cmeta name=\"msapplication-TileColor\" content=\"#da532c\"/\u003e\u003cmeta name=\"theme-color\" content=\"#ffffff\"/\u003e\u003cmeta name=\"description\" content=\"Web site created using create-react-app\"/\u003e\u003clink rel=\"manifest\" href=\"/manifest.json\"/\u003e\u003clink rel=\"apple-touch-icon\" sizes=\"180x180\" href=\"/apple-touch-icon.png\"/\u003e\u003clink rel=\"icon\" type=\"image/png\" sizes=\"32x32\" href=\"/favicon-32x32.png\"/\u003e\u003clink rel=\"icon\" type=\"image/png\" sizes=\"16x16\" href=\"/favicon-16x16.png\"/\u003e\u003clink rel=\"mask-icon\" href=\"/safari-pinned-tab.svg\" color=\"#5bbad5\"/\u003e\u003ctitle\u003eSeameet\u003c/title\u003e\u003clink href=\"/static/css/2.51964e6d.chunk.css\" rel="}},{"Protocol":"tcp","PortID":"4005","State":"open","Service":{"Name":"pxc-pin","Product":"","Cpe":""},"Script":{"ID":"fingerprint-strings","Output":"\n  GenericLines: \n    HTTP/1.1 400 Bad Request\n    Content-Type: text/plain; charset=utf-8\n    Connection: close\n    Request\n  GetRequest, HTTPOptions: \n    HTTP/1.0 200 OK\n    Accept-Ranges: bytes\n    Content-Length: 2558\n    Content-Type: text/html; charset=utf-8\n    Last-Modified: Fri, 18 Dec 2020 04:00:35 GMT\n    Date: Fri, 16 Jul 2021 16:51:45 GMT\n    \u003c!doctype html\u003e\u003chtml lang=\"en\"\u003e\u003chead\u003e\u003cmeta charset=\"utf-8\"/\u003e\u003cmeta name=\"viewport\" content=\"width=device-width,initial-scale=1\"/\u003e\u003cmeta name=\"msapplication-TileColor\" content=\"#da532c\"/\u003e\u003cmeta name=\"theme-color\" content=\"#ffffff\"/\u003e\u003cmeta name=\"description\" content=\"Web site created using create-react-app\"/\u003e\u003clink rel=\"manifest\" href=\"/manifest.json\"/\u003e\u003clink rel=\"apple-touch-icon\" sizes=\"180x180\" href=\"/apple-touch-icon.png\"/\u003e\u003clink rel=\"icon\" type=\"image/png\" sizes=\"32x32\" href=\"/favicon-32x32.png\"/\u003e\u003clink rel=\"icon\" type=\"image/png\" sizes=\"16x16\" href=\"/favicon-16x16.png\"/\u003e\u003clink rel=\"mask-icon\" href=\"/safari-pinned-tab.svg\" color=\"#5bbad5\"/\u003e\u003ctitle\u003eSeameet\u003c/title\u003e\u003clink href=\"/static/css/2.51964e6d.chunk.css\" rel="}},{"Protocol":"tcp","PortID":"4006","State":"open","Service":{"Name":"pxc-spvr","Product":"","Cpe":""},"Script":{"ID":"fingerprint-strings","Output":"\n  GenericLines: \n    HTTP/1.1 400 Bad Request\n    Content-Type: text/plain; charset=utf-8\n    Connection: close\n    Request\n  GetRequest, HTTPOptions: \n    HTTP/1.0 200 OK\n    Accept-Ranges: bytes\n    Content-Length: 2558\n    Content-Type: text/html; charset=utf-8\n    Last-Modified: Wed, 31 Mar 2021 04:30:41 GMT\n    Date: Fri, 16 Jul 2021 16:51:45 GMT\n    \u003c!doctype html\u003e\u003chtml lang=\"en\"\u003e\u003chead\u003e\u003cmeta charset=\"utf-8\"/\u003e\u003cmeta name=\"viewport\" content=\"width=device-width,initial-scale=1\"/\u003e\u003cmeta name=\"msapplication-TileColor\" content=\"#da532c\"/\u003e\u003cmeta name=\"theme-color\" content=\"#ffffff\"/\u003e\u003cmeta name=\"description\" content=\"Web site created using create-react-app\"/\u003e\u003clink rel=\"manifest\" href=\"/manifest.json\"/\u003e\u003clink rel=\"apple-touch-icon\" sizes=\"180x180\" href=\"/apple-touch-icon.png\"/\u003e\u003clink rel=\"icon\" type=\"image/png\" sizes=\"32x32\" href=\"/favicon-32x32.png\"/\u003e\u003clink rel=\"icon\" type=\"image/png\" sizes=\"16x16\" href=\"/favicon-16x16.png\"/\u003e\u003clink rel=\"mask-icon\" href=\"/safari-pinned-tab.svg\" color=\"#5bbad5\"/\u003e\u003ctitle\u003eSeameet\u003c/title\u003e\u003clink href=\"/static/css/2.51964e6d.chunk.css\" rel="}}]}`
	fmt.Println(data)

	dnsObj := Dns{
		DnsValue: "live-test.shopee.com",
		DnsType:  "CNAME",
		Domain:   "academy.live-test.redacted.io",
		//RawPorts: datatypes.JSON([]byte{data}),
		ScanRefer: 1,
	}
	DB.Create(&dnsObj)

	var obj []Asset
	DB.Limit(5).Preload("Asset").Find(&obj)
	spew.Dump(obj)
}

func TestQueryTable(t *testing.T) {
	setup()
	//tx := DB.Table("IPRanges")
	//tx.Where("target_refer = ?", targetId)
	var objs []IPRange
	DB.Find(&objs)

	spew.Dump(objs)
}

func TestQueryDns(t *testing.T) {
	setupRemote()
	var obj Dns

	// {"IPAddress":"103.115.77.222","Hostname":"","Ports":[{"Protocol":"tcp","PortID":"80","State":"open","Service":{"Name":"http","Product":"SGW","Cpe":""},"Script":{"ID":"http-title","Output":"Did not follow redirect to https://shopee.com/"}},{"Protocol":"tcp","PortID":"9101","State":"open","Service":{"Name":"jetdirect","Product":"","Cpe":""},"Script":{"ID":"","Output":""}}]}

	d := Dns{
		DnsValue: "103.115.78.200",
		DnsType:  "A",
		Ports:    "9101/",
	}

	DB.Model(Dns{}).Where("dns_value = ? AND dns_type = ?", d.DnsValue, d.DnsType).First(&obj)

	spew.Dump(obj)
}
