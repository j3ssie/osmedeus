# Osmedeus Core Engine

<p align="center">
  <img alt="Osmedeus" src="https://raw.githubusercontent.com/osmedeus/assets/main/logo-transparent.png" height="140" />
  <br />
  <strong>Osmedeus - A Workflow Engine for Offensive Security</strong>
  <p align="center">
  <a href="https://docs.osmedeus.org/"><img src="https://img.shields.io/badge/Documentation-0078D4?style=for-the-badge&logo=Google-Chrome&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://docs.osmedeus.org/donation/"><img src="https://img.shields.io/badge/Sponsors-0078D4?style=for-the-badge&logo=GitHub-Sponsors&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://twitter.com/OsmedeusEngine"><img src="https://img.shields.io/badge/%40OsmedeusEngine-0078D4?style=for-the-badge&logo=Twitter&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://discord.gg/gy4SWhpaPU"><img src="https://img.shields.io/badge/Discord%20Server-0078D4?style=for-the-badge&logo=Discord&logoColor=39ff14&labelColor=black&color=black"></a>
  </p>
</p>

***

## Installation

> NOTE that you need some essential tools like `curl, wget, git, zip` and login as **root** to start

```shell
bash -c "$(curl -fsSL https://raw.githubusercontent.com/osmedeus/osmedeus-base/master/install.sh)"
```

## Build the engine from source

Make sure you installed `golang >= v1.17`

```shell
mkdir -p $GOPATH/src/github.com/j3ssie
git clone --depth=1 https://github.com/j3ssie/osmedeus $GOPATH/src/github.com/j3ssie/osmedeus
cd $GOPATH/src/github.com/j3ssie/osmedeus
make build
```

## Usage

```shell
# Practical Usage:
  osmedeus scan -f [flowName] -t [target]
  osmedeus scan -f [flowName] -T [targetsFile]
  osmedeus scan -f /path/to/flow.yaml -t [target]
  osmedeus scan -m /path/to/module.yaml -t [target] --params 'port=9200'
  osmedeus scan -m /path/to/module.yaml -t [target] -l /tmp/log.log
  cat targets | osmedeus scan -f sample

# Example Commands:
  osmedeus scan -t target.com
  osmedeus scan -T list_of_targets.txt -W custom_workspaces
  osmedeus scan -t target.com -w workspace_name --debug
  osmedeus scan -f single -t www.sample.com
  osmedeus scan -f ovuln-T list_of_target.txt
  osmedeus scan -m ~/osmedeus-base/workflow/test/dirbscan.yaml -t list_of_urls.txt
  osmedeus health
  ls ~/.osmedeus/storages/summary/ | osmedeus scan -m ~/osmedeus-base/workflow/test/dirbscan.yaml
  ls ~/.osmedeus/storages/summary/ | osmedeus scan -m ~/osmedeus-base/workflow/test/busting.yaml -D

# Start Web UI at https://localhost:8000/ui/
osmedeus server
# login with credentials from `~/.osmedeus/config.yaml`

# Delete workspace
osmedeus config delete -w workspace_name
```

## 💬 Community & Discussion

Join Our Discord server [here](https://discord.gg/gy4SWhpaPU)

## 💎 Donation

Please check out [this page](https://docs.osmedeus.org/donation/) for couple donation methods here

## License

`Osmedeus` is made with ♥ by [@j3ssiejjj](https://twitter.com/j3ssiejjj) and it is released under the MIT license.
