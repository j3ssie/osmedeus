package provider

import (
    "context"
    "fmt"
    "io/ioutil"
    "strings"
    "time"

    "github.com/digitalocean/godo"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/spf13/cast"
)

// DefaultDO set some default data for DO provider
func (p *Provider) DefaultDO() {
    p.Region = "sfo3"
    p.Size = "s-2vcpu-4gb"

    if p.ProviderConfig.Region != "" {
        p.Region = p.ProviderConfig.Region
    }
    if p.ProviderConfig.Size != "" {
        p.Size = p.ProviderConfig.Size
    }
}

func (p *Provider) ClientDO() {
    client := godo.NewFromToken(p.Token)
    p.Client = client
}

func (p *Provider) AccountDO() error {
    client, ok := p.Client.(*godo.Client)
    if !ok {
        return fmt.Errorf("error convert client")
    }
    ctx := context.TODO()

    account, _, err := client.Account.Get(ctx)
    if err != nil {
        return fmt.Errorf("error getting account information")
    }
    p.InstanceLimit = account.DropletLimit

    bill, _, err := client.Balance.Get(ctx)
    if err != nil {
        return fmt.Errorf("error getting account information")
    }
    utils.InforF("Account Billing Information: MonthToDateBalance: %v -- AccountBalance: %v", color.HiRedString(bill.MonthToDateBalance), color.HiGreenString(bill.AccountBalance))

    return nil
}

func (p *Provider) ListInstanceDO() error {
    client, ok := p.Client.(*godo.Client)
    if !ok {
        return fmt.Errorf("error convert client")
    }

    ctx := context.TODO()
    opt := &godo.ListOptions{
        Page:    1,
        PerPage: 1000,
    }

    droplets, _, err := client.Droplets.List(ctx, opt)
    if err != nil {
        return fmt.Errorf("error getting digital ocean instance")
    }
    utils.InforF("found %v instances", len(droplets))

    for _, instance := range droplets {
        ipAddress, ok := instance.PublicIPv4()
        if ok != nil || ipAddress == "" {
            utils.ErrorF("Instance has no public IP: %v -- %v", instance.ID, instance.Name)
            continue
        }
        parsedInstance := Instance{
            InstanceID:   cast.ToString(instance.ID),
            IPAddress:    ipAddress,
            InstanceName: instance.Name,
            ImageID:      cast.ToString(instance.Image.ID),
            ImageName:    instance.Image.Name,
            Region:       instance.Region.Slug,
            Memory:       cast.ToString(instance.Memory),
            CPU:          cast.ToString(instance.Vcpus),
            Disk:         cast.ToString(instance.Disk),
            Status:       instance.Status,
            CreatedAt:    instance.Created,

            InputName:    "",
            ProviderName: "do",
        }

        p.Instances = append(p.Instances, parsedInstance)
    }

    // check if we reach max instance number
    if p.InstanceLimit > 0 {
        if len(p.Instances) >= p.InstanceLimit {
            p.Available = false
        }
    }

    return nil

}

func (p *Provider) GetSSHKeyDO() error {
    client, ok := p.Client.(*godo.Client)
    if !ok {
        return fmt.Errorf("error convert client")
    }

    ctx := context.TODO()
    opt := &godo.ListOptions{
        Page:    1,
        PerPage: 1000,
    }

    keys, _, err := client.Keys.List(ctx, opt)
    if err != nil {
        return fmt.Errorf("error listing ssh key -- %v", err)
    }

    for _, key := range keys {
        if strings.TrimSpace(key.PublicKey) == strings.TrimSpace(p.SSHPublicKey) {
            p.SSHKeyID = cast.ToString(key.ID)
            p.SSHKeyFound = true
            utils.DebugF("Found SSH Key: %v -- %v ", key.Name, p.SSHKeyID)
        }
    }

    // create one if not found
    if !p.SSHKeyFound {
        utils.DebugF("No SSHKey found. create a new one")
        createRequest := &godo.KeyCreateRequest{
            Name:      p.SSHKeyName,
            PublicKey: p.SSHPublicKey,
        }
        transfer, _, err := client.Keys.Create(ctx, createRequest)
        if err == nil {
            p.SSHKeyID = cast.ToString(transfer.ID)
            p.SSHKeyFound = true
            utils.DebugF("Created new SSH Key: %v", p.SSHKeyID)
        }
    }

    return nil
}

func (p *Provider) ListSnapshotDO() error {
    client, ok := p.Client.(*godo.Client)
    if !ok {
        utils.ErrorF("error convert client")
        return fmt.Errorf("error convert client")
    }

    ctx := context.TODO()
    opt := &godo.ListOptions{
        Page:    1,
        PerPage: 1000,
    }

    snapshots, _, err := client.Snapshots.List(ctx, opt)
    if err != nil {
        return fmt.Errorf("error getting digital ocean snapshot")
    }
    for _, instance := range snapshots {
        name := instance.Name
        id := cast.ToString(instance.ID)
        //uuid := instance.

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

func (p *Provider) CreateInstanceDO(name string) (dropletId int, err error) {
    client, ok := p.Client.(*godo.Client)
    if !ok {
        utils.ErrorF("error convert client")
        return dropletId, fmt.Errorf("error convert client")
    }

    ctx := context.TODO()
    createRequest := &godo.DropletCreateRequest{
        Name:   name,
        Region: p.Region,
        Size:   p.Size,
        Image: godo.DropletCreateImage{
            ID: cast.ToInt(p.SnapshotID),
            //Slug: "ubuntu-16-04-x64",
        },
        SSHKeys: []godo.DropletCreateSSHKey{
            godo.DropletCreateSSHKey{ID: cast.ToInt(p.SSHKeyID)},
        },
        Tags: []string{libs.SNAPSHOT},
    }

    instance, res, err := client.Droplets.Create(ctx, createRequest)
    if err != nil {
        utils.ErrorF("error create digital ocean instance -- %v", err)
        content, ok := ioutil.ReadAll(res.Body)
        if ok == nil {
            fmt.Println(string(content))
        }
        return dropletId, fmt.Errorf("error create digital ocean instance -- %v", err)
    }

    // get droplet IP info
    dropletId = instance.ID
    time.Sleep(60 * time.Second)
    return dropletId, nil
}

func (p *Provider) InstanceInfoDO(id int) (Instance, error) {
    var parsedInstance Instance
    client, ok := p.Client.(*godo.Client)
    if !ok {
        utils.ErrorF("error convert client")
        return parsedInstance, fmt.Errorf("error convert client")
    }
    ctx := context.TODO()
    instance, _, err := client.Droplets.Get(ctx, id)
    if err != nil {
        return parsedInstance, fmt.Errorf("error get instance info")
    }

    ipAddress, err := instance.PublicIPv4()
    if err != nil || ipAddress == "" {
        return parsedInstance, fmt.Errorf("no public ip address yet")
    }
    parsedInstance = Instance{
        InstanceID:   cast.ToString(instance.ID),
        IPAddress:    ipAddress,
        InstanceName: instance.Name,
        ImageID:      cast.ToString(instance.Image.ID),
        ImageName:    instance.Image.Name,
        Region:       instance.Region.Slug,
        Memory:       cast.ToString(instance.Memory),
        CPU:          cast.ToString(instance.Vcpus),
        Disk:         cast.ToString(instance.Disk),
        Status:       instance.Status,
        CreatedAt:    instance.Created,

        InputName:    "",
        ProviderName: "do",
    }
    p.CreatedInstance = parsedInstance
    utils.DebugF("Created instance ID: %v -- %v -- %v", p.CreatedInstance.InstanceID, p.CreatedInstance.InstanceName, p.CreatedInstance.IPAddress)

    return parsedInstance, nil
}

func (p *Provider) DeleteInstanceDO(id string) error {
    client, ok := p.Client.(*godo.Client)
    if !ok {
        utils.ErrorF("error convert client")
        return fmt.Errorf("error convert client")
    }
    ctx := context.TODO()
    _, err := client.Droplets.Delete(ctx, cast.ToInt(id))
    if err != nil {
        utils.ErrorF("error delete instance -- %v", err)
        return fmt.Errorf("error delete instance")
    }
    utils.InforF("Deleted instance ID: %v", id)
    return nil
}

func (p *Provider) DeleteSnapShotDO(id string) error {
    client, ok := p.Client.(*godo.Client)
    if !ok {
        utils.ErrorF("error convert client")
        return fmt.Errorf("error convert client")
    }
    ctx := context.TODO()

    _, err := client.Snapshots.Delete(ctx, id)
    if err != nil {
        utils.ErrorF("error delete snapshot -- %v", err)
        return fmt.Errorf("error delete instance")
    }
    utils.InforF("Deleted snapshot ID: %v", id)
    return nil
}
