package provider

import (
    "github.com/davecgh/go-spew/spew"
    "testing"
)

var sampleToken = "d17ab13b0a3fccafefc932b9db95be45464339073100423336045977c0924491"

func TestProviderList(t *testing.T) {
    provider, err := InitProvider("do", sampleToken)
    if err != nil {
        t.Errorf("error")
    }
    provider.SSHPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC0ZmMumN5GTKuPXVVqugz+6BZs6JyNaQGsPgsZk86uON//oXi5fkutsNu5IIPDZXph3P5NbUj2dPalNyzNY5jClPJeT0F/eowIhPeo5GfjmJlXR4TTgHSwPYrOQNop2w+xh2z5h4IXMEPVLpEsN67MuUjTKzRwirwPjigZS/gayhwfOfDsPaIwBhpoBGuR1x+Rzzxuiy7TToNoWhF6pT9qONoCtr0VrPMsmjVpEPKD/uTW/8KeFL0pb/9z18M4IlbtvkO0Y6RhrpFGNSmZTWc1eDsJpFJerrVd48rgx3aRHriijl4zX4GBhc0zjqJwv+nGTGFPJ9Tx/3kPMDUGna/f91VU7sL7YqeiSed8S0YcWfntYy64OknvMpN8VIoQ7WiJAkR3wPw+tL3ZduXXAiKHFTAiXev02mOvo2F2nQKdGS98lOH5m+zuUm8abYbyXYlGNEzz576ksb6nMWCSSXwhA5f4clPKaPmgBQFQMUtq6Wgb8Fjq2r1MpjIWwUvx84s= osmp-cloud"
    provider.Prepare()

    provider.ListInstance()
    spew.Dump(provider.Instances)
}

func TestProviderCreate(t *testing.T) {
    provider, err := InitProvider("do", sampleToken)
    if err != nil {
        t.Errorf("error ")
    }
    provider.SSHPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC0ZmMumN5GTKuPXVVqugz+6BZs6JyNaQGsPgsZk86uON//oXi5fkutsNu5IIPDZXph3P5NbUj2dPalNyzNY5jClPJeT0F/eowIhPeo5GfjmJlXR4TTgHSwPYrOQNop2w+xh2z5h4IXMEPVLpEsN67MuUjTKzRwirwPjigZS/gayhwfOfDsPaIwBhpoBGuR1x+Rzzxuiy7TToNoWhF6pT9qONoCtr0VrPMsmjVpEPKD/uTW/8KeFL0pb/9z18M4IlbtvkO0Y6RhrpFGNSmZTWc1eDsJpFJerrVd48rgx3aRHriijl4zX4GBhc0zjqJwv+nGTGFPJ9Tx/3kPMDUGna/f91VU7sL7YqeiSed8S0YcWfntYy64OknvMpN8VIoQ7WiJAkR3wPw+tL3ZduXXAiKHFTAiXev02mOvo2F2nQKdGS98lOH5m+zuUm8abYbyXYlGNEzz576ksb6nMWCSSXwhA5f4clPKaPmgBQFQMUtq6Wgb8Fjq2r1MpjIWwUvx84s= osmp-cloud"
    provider.Prepare()

    id, err := provider.CreateInstanceDO("example.com")
    if err == nil {
        provider.GetInstanceInfo(id)
    }
    spew.Dump(provider.Instances)
}

func TestProviderDelete(t *testing.T) {
    provider, err := InitProvider("do", sampleToken)
    if err != nil {
        t.Errorf("error ")
    }

    //provider.SSHPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC0ZmMumN5GTKuPXVVqugz+6BZs6JyNaQGsPgsZk86uON//oXi5fkutsNu5IIPDZXph3P5NbUj2dPalNyzNY5jClPJeT0F/eowIhPeo5GfjmJlXR4TTgHSwPYrOQNop2w+xh2z5h4IXMEPVLpEsN67MuUjTKzRwirwPjigZS/gayhwfOfDsPaIwBhpoBGuR1x+Rzzxuiy7TToNoWhF6pT9qONoCtr0VrPMsmjVpEPKD/uTW/8KeFL0pb/9z18M4IlbtvkO0Y6RhrpFGNSmZTWc1eDsJpFJerrVd48rgx3aRHriijl4zX4GBhc0zjqJwv+nGTGFPJ9Tx/3kPMDUGna/f91VU7sL7YqeiSed8S0YcWfntYy64OknvMpN8VIoQ7WiJAkR3wPw+tL3ZduXXAiKHFTAiXev02mOvo2F2nQKdGS98lOH5m+zuUm8abYbyXYlGNEzz576ksb6nMWCSSXwhA5f4clPKaPmgBQFQMUtq6Wgb8Fjq2r1MpjIWwUvx84s= osmp-cloud"
    //
    //provider.Prepare()

    provider.DeleteInstance("258443728")
    spew.Dump(provider.Instances)
}

func TestProviderAccount(t *testing.T) {
    provider, err := InitProvider("do", sampleToken)
    if err != nil {
        t.Errorf("error ")
    }

    provider.AccountDO()
}

func TestProvider_DeleteSnapshot(t *testing.T) {
    provider, err := InitProvider("do", sampleToken)
    if err != nil {
        t.Errorf("error ")
    }

    provider.DeleteSnapShotDO("89053110")
    provider.ListSnapshotDO()
    spew.Dump(provider.OldSnapShotID)
    spew.Dump(provider.SnapshotID)
}

func TestProviderLN(t *testing.T) {
    sampleToken = "6634197757df67b24753d0d241003ae09afd53bc7f9648c191e37acb61bfef37"
    provider, err := InitProvider("linode", sampleToken)
    if err != nil {
        t.Errorf("error ")
    }
    provider.SSHPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC0ZmMumN5GTKuPXVVqugz+6BZs6JyNaQGsPgsZk86uON//oXi5fkutsNu5IIPDZXph3P5NbUj2dPalNyzNY5jClPJeT0F/eowIhPeo5GfjmJlXR4TTgHSwPYrOQNop2w+xh2z5h4IXMEPVLpEsN67MuUjTKzRwirwPjigZS/gayhwfOfDsPaIwBhpoBGuR1x+Rzzxuiy7TToNoWhF6pT9qONoCtr0VrPMsmjVpEPKD/uTW/8KeFL0pb/9z18M4IlbtvkO0Y6RhrpFGNSmZTWc1eDsJpFJerrVd48rgx3aRHriijl4zX4GBhc0zjqJwv+nGTGFPJ9Tx/3kPMDUGna/f91VU7sL7YqeiSed8S0YcWfntYy64OknvMpN8VIoQ7WiJAkR3wPw+tL3ZduXXAiKHFTAiXev02mOvo2F2nQKdGS98lOH5m+zuUm8abYbyXYlGNEzz576ksb6nMWCSSXwhA5f4clPKaPmgBQFQMUtq6Wgb8Fjq2r1MpjIWwUvx84s= osmp-cloud"
    //provider.GetSSHKey()

    provider.Prepare()
    //id, _ := provider.CreateInstanceLN("new2")
    //time.Sleep(90*time.Second)
    //id := 29530061
    //provider.MountDiskLN(id)
    provider.DeleteInstance("29530163")
    //provider.AccountLN()
    spew.Dump(provider.Instances)
}
