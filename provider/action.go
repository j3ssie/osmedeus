package provider

import (
	"fmt"
	"strings"

	"github.com/cenkalti/backoff/v4"
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/spf13/cast"
)

const (
	GetSSHKey       = "get-sshkey"
	RunBuild        = "run-build"
	ListImage       = "list-image"
	ListInstance    = "list-instance"
	GetInstanceInfo = "get-instance"
	CreateInstance  = "create-instance"
	BootInstance    = "boot-instance"
)

func (p *Provider) GetSSHKey() (err error) {
	if strings.TrimSpace(p.SSHPublicKey) == "" {
		return fmt.Errorf("error getting SSHKey -- blank ssh public key")
	}

	switch p.ProviderName {
	case "do", "digitalocean":
		err = p.GetSSHKeyDO()
	case "ln", "line", "linode":
		err = p.GetSSHKeyLN()
	case "aw", "aws", "asw":
		err = p.GetSSHKeyAWS()
	default:
		err = p.GetSSHKeyDO()
	}
	return err
}

func (p *Provider) ListInstance() (err error) {
	switch p.ProviderName {
	case "do", "digitalocean":
		err = p.ListInstanceDO()
	case "ln", "line", "linode":
		err = p.ListInstanceLN()
	case "aw", "aws", "asw":
		err = p.ListInstanceAWS()
	default:
		err = p.ListInstanceDO()
	}
	return err
}

func (p *Provider) ListSnapShot() (err error) {
	switch p.ProviderName {
	case "do", "digitalocean":
		err = p.ListSnapshotDO()
	case "ln", "line", "linode":
		err = p.ListSnapshotLN()
	case "aw", "aws", "asw":
		err = p.ListSnapshotAWS()
	default:
		err = p.ListSnapshotDO()
	}

	if !p.IsBackgroundCheck {
		utils.InforF("Found base image snapshot with ID: %v", color.HiCyanString(p.SnapshotID))
	}
	return err
}

func (p *Provider) CreateInstance(name string) (err error) {
	var id interface{}
	operation := func() error {
		switch p.ProviderName {
		case "do", "digitalocean":
			id, err = p.CreateInstanceDO(name)
			if err == nil {
				err = p.Action(GetInstanceInfo, id)
			}
		case "ln", "line", "linode":
			id, err = p.CreateInstanceLN(name)
			if err == nil {
				err = p.Action(BootInstance, id)
				err = p.Action(GetInstanceInfo, id)
			}

		case "aw", "aws", "asw":
			id, err = p.CreateInstanceAWS(name)
			if err == nil {
				err = p.Action(BootInstance, id)
				err = p.Action(GetInstanceInfo, id)
			}
		default:
			id, err = p.CreateInstanceDO(name)
			if err == nil {
				err = p.Action(GetInstanceInfo, id)
			}
		}
		return err
	}
	err = backoff.Retry(operation, p.BackOff)
	if err != nil {
		utils.WarnF("error create instance action %v -- %v", p.ProviderName, name)
		return err
	}
	return nil
}

// func (p *Provider) CreateInstanceF(name string) (err error) {
// 	var id int
// 	switch p.ProviderName {
// 	case "do", "digitalocean":
// 		id, err = p.CreateInstanceDO(name)
// 		if err == nil {
// 			err = p.Action(GetInstanceInfo, id)
// 		}
// 	case "ln", "line", "linode":
// 		id, err = p.CreateInstanceLN(name)
// 		if err == nil {
// 			err = p.Action(BootInstance, id)
// 			err = p.Action(GetInstanceInfo, id)
// 		}
// 	default:
// 		id, err = p.CreateInstanceDO(name)
// 		if err == nil {
// 			err = p.Action(GetInstanceInfo, id)
// 		}
// 	}
// 	return err
// }

func (p *Provider) BootInstance(id interface{}) (err error) {
	switch p.ProviderName {
	case "do", "digitalocean":
	case "ln", "line", "linode":
		err = p.BootInstanceLN(cast.ToInt(id))
	case "aw", "aws", "asw":
		// err = p.AllowRootAccessAWS(cast.ToString(id))
	default:
		err = p.BootInstanceLN(cast.ToInt(id))
	}
	if err != nil {
		utils.WarnF("error booting instance: %v", id)
		return err
	}
	return nil
}

func (p *Provider) GetInstanceInfo(id interface{}) (err error) {
	var instance Instance
	switch p.ProviderName {
	case "do", "digitalocean":
		instance, err = p.InstanceInfoDO(cast.ToInt(id))
	case "ln", "line", "linode":
		instance, err = p.InstanceInfoLN(cast.ToInt(id))
	case "aw", "aws", "asw":
		instance, err = p.InstanceInfoAWS(cast.ToString(id))
	default:
		instance, err = p.InstanceInfoDO(cast.ToInt(id))
	}
	if err != nil {
		utils.WarnF("error getting public IP of instance: %v", color.HiBlueString("%v", id))
		return err
	}
	p.Instances = append(p.Instances, instance)
	return nil
}

func (p *Provider) DeleteInstance(id string) (err error) {
	utils.DebugF("[%v] Delete instance: %v", p.ProviderName, id)
	switch p.ProviderName {
	case "do", "digitalocean":
		err = p.DeleteInstanceDO(id)
	case "ln", "line", "linode":
		err = p.DeleteInstanceLN(id)
	case "aw", "aws", "asw":
		err = p.DeleteInstanceAWS(id)
	default:
		err = p.DeleteInstanceDO(id)
	}
	return err
}

func (p *Provider) DeleteOldSnapshot() (err error) {
	for _, id := range p.OldSnapShotID {
		switch p.ProviderName {
		case "do", "digitalocean":
			err = p.DeleteSnapShotDO(id)
		case "ln", "line", "linode":
			err = p.DeleteSnapShotLN(id)
		case "aw", "aws", "asw":
			err = p.DeleteImageAWS(id)
		default:
			err = p.DeleteSnapShotDO(id)
		}
	}

	return err
}
