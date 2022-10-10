package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/linode/linodego"
	//"github.com/linode/linodego/pkg/errors"
	"github.com/spf13/cast"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"strings"
	"time"
)

// DefaultLinode set some default data for DO provider
func (p *Provider) DefaultLinode() {
	p.Region = "us-east"
	p.Size = "g6-standard-1"
	if p.ProviderConfig.Region != "" {
		p.Region = p.ProviderConfig.Region
	}
	if p.ProviderConfig.Size != "" {
		p.Size = p.ProviderConfig.Size
	}

	if p.Opt.Cloud.Size != "" {
		p.Size = p.Opt.Cloud.Size
	}
	if p.Opt.Cloud.Region != "" {
		p.Region = p.Opt.Cloud.Region
	}

	p.LinodeDiskMap()
}

func (p *Provider) ClientLinode() error {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: p.Token})
	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	linodeClient := linodego.NewClient(oauth2Client)

	p.Client = linodeClient
	return nil
}

func (p *Provider) AccountLN() error {
	client, ok := p.Client.(linodego.Client)
	if !ok {
		err := fmt.Errorf("error convert client")
		utils.ErrorF("%v", err)
		return err
	}
	ctx := context.TODO()
	account, err := client.GetAccount(ctx)
	if err != nil {
		return fmt.Errorf("error getting account information")
	}
	utils.InforF("Account Billing Information: BalanceUninvoiced: %v -- AccountBalance: %v", color.HiRedString("%v", account.BalanceUninvoiced), color.HiGreenString("%v", account.Balance))

	//accountSett, err := client.GetAccountSettings(ctx)
	//if err != nil {
	//	//err := fmt.Errorf("error convert client")
	//	utils.ErrorF("%v", err)
	//	return err
	//}
	//if err != nil && accountSett.NetworkHelper == false {
	helper := false
	opt := linodego.AccountSettingsUpdateOptions{
		NetworkHelper: &helper,
	}

	// @TODO: no idea why this function false
	// client.UpdateAccountSettings(context.Background(), opt)

	req := client.R(ctx).SetResult(&linodego.AccountSettings{})
	if bodyData, err := json.Marshal(&opt); err == nil {
		req.URL = "https://api.linode.com/v4/account/settings"
		//fmt.Println("err marshal", err)
		body := string(bodyData)
		req.SetBody(body).Put(req.URL)
	}
	//}
	return err
}

// LinodeTest list all instances
func (p *Provider) LinodeTest() error {
	linodeClient, ok := p.Client.(linodego.Client)
	if !ok {
		return fmt.Errorf("error convert client")
	}

	res, err := linodeClient.GetInstance(context.Background(), 4090913)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v", res)

	return nil
}

func (p *Provider) ListInstanceLN() error {
	client, ok := p.Client.(linodego.Client)
	if !ok {
		return fmt.Errorf("error convert client")
	}

	opt := &linodego.ListOptions{
		PageOptions: nil,
		Filter:      "",
	}

	instances, err := client.ListInstances(context.TODO(), opt)
	if err != nil {
		return fmt.Errorf("error getting linode instance")
	}

	for _, instance := range instances {
		//instance.IPv4
		if len(instance.IPv4) == 0 {
			utils.ErrorF("Instance has no public IP: %v -- %v", instance.ID, instance.Label)
			continue
		}
		ipAddress := instance.IPv4[0]

		parsedInstance := Instance{
			InstanceID:   cast.ToString(instance.ID),
			IPAddress:    cast.ToString(ipAddress),
			InstanceName: instance.Label,
			ImageID:      cast.ToString(instance.Image),
			ImageName:    instance.Image,
			Region:       instance.Region,
			Memory:       cast.ToString(instance.Specs.Memory),
			CPU:          cast.ToString(instance.Specs.VCPUs),
			Disk:         cast.ToString(instance.Specs.Disk),
			Status:       cast.ToString(instance.Status),
			CreatedAt:    cast.ToString(instance.Created),

			InputName:    "",
			ProviderName: "do",
		}

		p.Instances = append(p.Instances, parsedInstance)
	}
	return nil
}

func (p *Provider) GetSSHKeyLN() error {
	client, ok := p.Client.(linodego.Client)
	if !ok {
		return fmt.Errorf("error convert client")
	}
	ctx := context.TODO()
	opt := &linodego.ListOptions{
		PageOptions: nil,
		Filter:      "",
	}

	keys, err := client.ListSSHKeys(ctx, opt)
	if err != nil {
		return fmt.Errorf("error listing ssh key -- %v", err)
	}

	for _, key := range keys {
		if strings.TrimSpace(key.SSHKey) == strings.TrimSpace(p.SSHPublicKey) {
			p.SSHKeyID = cast.ToString(key.ID)
			p.SSHKeyFound = true
			utils.DebugF("Found SSH Key: %v -- %v ", key.Label, p.SSHKeyID)
		}
	}

	// create one if not found
	if !p.SSHKeyFound {
		utils.DebugF("No SSHKey found. create a new one")
		createRequest := linodego.SSHKeyCreateOptions{
			Label:  p.SSHKeyName,
			SSHKey: p.SSHPublicKey,
		}
		key, err := client.CreateSSHKey(ctx, createRequest)
		if err == nil {
			p.SSHKeyID = cast.ToString(key.ID)
			p.SSHKeyFound = true
			utils.DebugF("Created new SSH Key: %v", p.SSHKeyID)
		}
	}

	return nil
}

func (p *Provider) ListSnapshotLN() error {
	client, ok := p.Client.(linodego.Client)
	if !ok {
		return fmt.Errorf("error convert client")
	}
	ctx := context.TODO()
	opt := &linodego.ListOptions{
		//PageOptions: linodego.PageOptions{
		//	Page:    0,
		//	Pages:   0,
		//	Results: 0,
		//},
		PageOptions: nil,
		Filter:      "",
	}

	snapshots, err := client.ListImages(ctx, opt)
	if err != nil {
		return fmt.Errorf("error getting linode images")
	}
	for _, image := range snapshots {
		name := image.Label
		id := cast.ToString(image.ID)
		if strings.HasPrefix(name, libs.SNAPSHOT) {
			p.OldSnapShotID = append(p.OldSnapShotID, id)
		}

		if strings.TrimSpace(name) == strings.TrimSpace(p.SnapshotName) {
			utils.InforF("Found base image snapshot with ID: %s", id)
			p.SnapshotID = id
			p.SnapshotName = name
			p.SnapshotFound = true
		}
	}

	return nil
}

func (p *Provider) LinodeDiskMap() {
	p.SwapSizeMap = make(map[string]int)
	p.SwapSizeMap["g6-nanode-1"] = 20000
	p.SwapSizeMap["g6-standard-1"] = 4000
	p.SwapSizeMap["g6-standard-2"] = 8000
	p.SwapSizeMap["g6-standard-4"] = 16000
	p.SwapSizeMap["g6-standard-6"] = 32000
}

func (p *Provider) CreateInstanceLN(name string) (dropletId int, err error) {
	client, ok := p.Client.(linodego.Client)
	if !ok {
		return 0, fmt.Errorf("error convert client")
	}

	ctx := context.TODO()
	booted := false
	swapSize := 4000
	if swap, ok := p.SwapSizeMap[p.Size]; ok {
		swapSize = swap
	}

	createRequest := linodego.InstanceCreateOptions{
		Region:   p.Region,
		Type:     p.Size,
		Label:    name,
		RootPass: utils.RandomString(10),
		AuthorizedKeys: []string{
			p.SSHPublicKey,
			"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC/oKuLLgmqcGK8lf9D2GPZqSXzvE9QqsSPmHdWOybOIYkUzAioG97vcTy6iBiuoRDqRwxeTt7JE/IncfdEc8e7ofa5IoGg5Xx65G1tw/V2rY2yux/IQYV10GVfPlV4Un04z7Vx7oFGcgIcJRxTlyoVPrSCTM1cepp5hAW8eqyF7mJZU/5cWCJMRK7F3CX3p6Li1y0f4dOcMuGvZePQUt4P1ntkrKVijxihiJ0nEfSINee2oKTNM0ZIpoxTmmMIBksMK9Ayl4jgxMf5kPoMQHJM+kTg6IFcarYvKdG2qizqDbjgiORG+2AgWaiAmJb+uik0cNi+gyMNanrxL979yUsj root@j3ssie",
		},
		//AuthorizedUsers: []string{"root"},
		Image: p.SnapshotID,
		Tags:  []string{libs.SNAPSHOT},

		SwapSize: &swapSize,
		Booted:   &booted,
		//StackScriptID:   0,
		//Group:           "",
		//StackScriptData: nil,
		//BackupID:        0,
		//BackupsEnabled:  false,
		//PrivateIP:       false,
	}

	utils.DebugF("Creating instance based on %v/%v image with password: %v", p.Size, p.SnapshotID, createRequest.RootPass)
	instance, err := client.CreateInstance(ctx, createRequest)
	if err != nil {
		utils.ErrorF("error create linode instance %v -- %v", name, err)
		return dropletId, fmt.Errorf("error create linode instance -- %v", err)
	}

	//spew.Dump(instance)
	// get droplet IP info
	dropletId = instance.ID
	time.Sleep(60 * time.Second)
	return dropletId, nil
}

func (p *Provider) BootInstanceLN(dropletId int) error {
	client, ok := p.Client.(linodego.Client)
	if !ok {
		return fmt.Errorf("error convert client")
	}

	ctx := context.TODO()
	err := client.BootInstance(ctx, dropletId, 0)
	return err
}

func (p *Provider) MountDiskLN(dropletId int) error {
	client, ok := p.Client.(linodego.Client)
	if !ok {
		return fmt.Errorf("error convert client")
	}
	utils.InforF("Mounting disk: %v", dropletId)
	//err := client.BootInstance(ctx, dropletId, 0)
	ctx := context.TODO()

	disk, err := client.CreateInstanceDisk(ctx, dropletId, linodego.InstanceDiskCreateOptions{
		Label:      "test1",
		Filesystem: "ext4",
		Size:       2000,
	})
	if err != nil {
		utils.ErrorF("Error creating disk for resize: %s", err)
	}

	disk, err = client.WaitForInstanceDiskStatus(ctx, dropletId, disk.ID, linodego.DiskReady, 180)
	if err != nil {
		utils.ErrorF("Error waiting for disk readiness for resize: %s", err)
		return err
	}
	err = client.ResizeInstanceDisk(ctx, dropletId, disk.ID, 4000)
	if err != nil {
		utils.ErrorF("Error resizing instance disk: %s", err)
	}

	return nil
}

func (p *Provider) InstanceInfoLN(id int) (Instance, error) {
	var parsedInstance Instance
	client, ok := p.Client.(linodego.Client)
	if !ok {
		return parsedInstance, fmt.Errorf("error convert client")
	}

	instance, err := client.GetInstance(context.TODO(), id)
	if err != nil {
		return parsedInstance, fmt.Errorf("error getting linode instance")
	}

	if len(instance.IPv4) == 0 {
		utils.ErrorF("Instance has no public IP: %v -- %v", instance.ID, instance.Label)
		return parsedInstance, fmt.Errorf("no public ip address yet")
	}
	ipAddress := instance.IPv4[0]
	parsedInstance = Instance{
		InstanceID:   cast.ToString(instance.ID),
		IPAddress:    cast.ToString(ipAddress),
		InstanceName: instance.Label,
		ImageID:      cast.ToString(instance.Image),
		ImageName:    instance.Image,
		Region:       instance.Region,
		Memory:       cast.ToString(instance.Specs.Memory),
		CPU:          cast.ToString(instance.Specs.VCPUs),
		Disk:         cast.ToString(instance.Specs.Disk),
		Status:       cast.ToString(instance.Status),
		CreatedAt:    cast.ToString(instance.Created),
		InputName:    "",
		ProviderName: "linode",
	}
	p.CreatedInstance = parsedInstance
	utils.DebugF("Created instance ID: %v -- %v -- %v", p.CreatedInstance.InstanceID, p.CreatedInstance.InstanceName, p.CreatedInstance.IPAddress)

	return parsedInstance, nil
}

func (p *Provider) DeleteInstanceLN(id string) error {
	client, ok := p.Client.(linodego.Client)
	if !ok {
		return fmt.Errorf("error convert client")
	}

	ctx := context.TODO()
	err := client.DeleteInstance(ctx, cast.ToInt(id))
	if err != nil {
		utils.ErrorF("error delete instance -- %v", err)
		return fmt.Errorf("error delete instance")
	}
	utils.InforF("Deleted instance ID: %v", color.HiRedString(id))
	return nil
}

func (p *Provider) DeleteSnapShotLN(id string) error {
	client, ok := p.Client.(linodego.Client)
	if !ok {
		return fmt.Errorf("error convert client")
	}

	ctx := context.TODO()
	err := client.DeleteImage(ctx, id)
	if err != nil {
		utils.ErrorF("error delete snapshot -- %v", err)
		return fmt.Errorf("error delete instance")
	}
	utils.InforF("Deleted snapshot ID: %v", color.HiRedString(id))
	return nil
}

// @NOTE: old note for increase disk in linode
//func (c *CloudRunner) MountDiskLinode() error {
//	diskSize := p.DisksMap[c.Cloud.Size]
//	if diskSize == "" {
//		return fmt.Errorf("can't found disk size")
//	}
//
//	utils.InforF("Mounting Disk size %s for instance %s", diskSize, c.InstanceID)
//	cmd := c.Prefix + fmt.Sprintf(`linodes disks-list %s`, c.InstanceID)
//	out := c.RetryCommandWithExcludeString(cmd, `Request failed:`)
//	jsonParsed, err := gabs.ParseJSON([]byte(out))
//	if err != nil {
//		utils.ErrorF("error when parsing content of droplet list")
//		return err
//	}
//
//	var diskID string
//	for _, instance := range jsonParsed.Children() {
//		filesystem := cast.ToString(instance.S("filesystem").Data())
//		if strings.HasPrefix(filesystem, "ext") {
//			diskID = cast.ToString(instance.S("id").Data())
//		}
//	}
//	if diskID == "" {
//		return fmt.Errorf("error to find disk ID")
//	}
//
//	cmd = c.Prefix + fmt.Sprintf(`linodes disk-resize %s %s --size %s`, c.InstanceID, diskID, diskSize)
//	out = c.RetryCommandWithExcludeString(cmd, `Request failed:`)
//	if strings.Contains(out, "Request failed") {
//		return fmt.Errorf("error to mount disk")
//	}
//
//	time.Sleep(5 * time.Second)
//	// everything done booting the instance up
//	cmd = c.Prefix + fmt.Sprintf(`linodes boot  %s `, c.InstanceID)
//	out = c.RetryCommandWithExcludeString(cmd, `Request failed:`)
//	if strings.Contains(out, "Request failed") {
//		return fmt.Errorf("error to mount disk")
//	}
//
//	return nil
//}
//
//func (c *CloudRunner) LinodeDiskMap() {
//	c.DisksMap = make(map[string]string)
//	c.DisksMap["g6-nanode-1"] = "20000"
//	c.DisksMap["g6-standard-1"] = "45000"
//	c.DisksMap["g6-standard-2"] = "75000"
//	c.DisksMap["g6-standard-4"] = "150000"
//	c.DisksMap["g6-standard-6"] = "320000"
//}
//
//var reservedAddrRanges []*net.IPNet
//
//var ReservedCIDRs = []string{
//	"192.168.0.0/16",
//	"172.16.0.0/12",
//	"10.0.0.0/8",
//	"127.0.0.0/8",
//	"224.0.0.0/4",
//	"240.0.0.0/4",
//	"100.64.0.0/10",
//	"198.18.0.0/15",
//	"169.254.0.0/16",
//	"192.88.99.0/24",
//	"192.0.0.0/24",
//	"192.0.2.0/24",
//	"192.94.77.0/24",
//	"192.94.78.0/24",
//	"192.52.193.0/24",
//	"192.12.109.0/24",
//	"192.31.196.0/24",
//	"192.0.0.0/29",
//}
//
//func init() {
//	for _, cidr := range ReservedCIDRs {
//		if _, ipnet, err := net.ParseCIDR(cidr); err == nil {
//			reservedAddrRanges = append(reservedAddrRanges, ipnet)
//		}
//	}
//}
//
//// IsPrivateIP checks if the addr parameter is within one of the address ranges in the ReservedCIDRs slice.
//func IsPrivateIP(addr string) bool {
//	ip := net.ParseIP(addr)
//	if ip == nil {
//		return false
//	}
//
//	var cidr string
//	for _, block := range reservedAddrRanges {
//		if block.Contains(ip) {
//			cidr = block.String()
//			break
//		}
//	}
//
//	if cidr != "" {
//		return true
//	}
//	return false
//}
