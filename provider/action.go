package provider

import (
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/j3ssie/osmedeus/utils"
	"strings"
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
	default:
		err = p.ListSnapshotDO()
	}
	return err
}

func (p *Provider) CreateInstance(name string) (err error) {
	var id int
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

func (p *Provider) CreateInstanceF(name string) (err error) {
	var id int
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
	default:
		id, err = p.CreateInstanceDO(name)
		if err == nil {
			err = p.Action(GetInstanceInfo, id)
		}
	}
	return err
}

func (p *Provider) BootInstance(id int) (err error) {
	switch p.ProviderName {
	case "do", "digitalocean":
	case "ln", "line", "linode":
		err = p.BootInstanceLN(id)
	default:
		err = p.BootInstanceLN(id)
	}
	if err != nil {
		utils.WarnF("error booting instance: %v", id)
		return err
	}
	return nil
}

func (p *Provider) GetInstanceInfo(id int) (err error) {
	var instance Instance
	switch p.ProviderName {
	case "do", "digitalocean":
		instance, err = p.InstanceInfoDO(id)
	case "ln", "line", "linode":
		instance, err = p.InstanceInfoLN(id)
	default:
		instance, err = p.InstanceInfoDO(id)
	}
	if err != nil {
		utils.WarnF("error getting public IP of instance: %v", instance.InstanceID)
		return err
	}
	return nil
}

func (p *Provider) DeleteInstance(id string) (err error) {
	utils.DebugF("[%v] Delete instance: %v", p.ProviderName, id)
	switch p.ProviderName {
	case "do", "digitalocean":
		err = p.DeleteInstanceDO(id)
	case "ln", "line", "linode":
		err = p.DeleteInstanceLN(id)
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
		default:
			err = p.DeleteSnapShotDO(id)
		}
	}

	return err
}
