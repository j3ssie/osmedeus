# Osmedeus Core Engine

<p align="center">
  <a href="https://www.osmedeus.org"><img alt="Osmedeus" src="https://raw.githubusercontent.com/osmedeus/assets/main/logo-transparent.png" height="140" /></a>
  <br />
  <strong>Osmedeus - A Workflow Engine for Offensive Security</strong>

  <p align="center">
  <a href="https://docs.osmedeus.org/"><img src="https://img.shields.io/badge/Documentation-0078D4?style=for-the-badge&logo=GitBook&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://docs.osmedeus.org/donation/"><img src="https://img.shields.io/badge/Sponsors-0078D4?style=for-the-badge&logo=GitHub-Sponsors&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://twitter.com/OsmedeusEngine"><img src="https://img.shields.io/badge/%40OsmedeusEngine-0078D4?style=for-the-badge&logo=Twitter&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://discord.gg/gy4SWhpaPU"><img src="https://img.shields.io/badge/Discord%20Server-0078D4?style=for-the-badge&logo=Discord&logoColor=39ff14&labelColor=black&color=black"></a>
  <a href="https://discord.gg/gy4SWhpaPU"><img src="https://img.shields.io/github/release/j3ssie/osmedeus?style=for-the-badge&labelColor=black&color=2fc414&logo=Github"></a>
  </p>
</p>

***

## 🔥 What is Osmedeus?

Osmedeus is a Workflow Engine for Offensive Security. It was designed to build a foundation with the capability and
flexibility that allows you to build your own reconnaissance system and run it on a large number of targets.

## 📖 Documentation & FAQ

You can check out the documentation at [**docs.osmedeus.org**](https://docs.osmedeus.org) and the Frequently Asked
Questions at [**here**](https://docs.osmedeus.org/faq) for more information.

## 📦 Installation

> NOTE that you need some essential tools like `curl, wget, git, zip` and login as **root** to start

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/osmedeus/osmedeus-base/master/install.sh)
```

### Build the engine from the source

Make sure you installed `golang >= v1.17`

```bash
go install -v github.com/j3ssie/osmedeus@latest
```

Check out [**this page**](https://docs.osmedeus.org/installation/) for more the install on other platforms and [**docker
image**](https://docs.osmedeus.org/installation/using-docker/).

## 🚀 Key Features of Osmedeus

- [x] Significantly speed up your recon process
- [x] Organize your scan results
- [x] Efficiently to customize and optimize your recon process
- [x] Seamlessly integrate with new public and private tools
- [x] Easy to scale across large number of targets
- [x] Easy to synchronize the results across many places

## 💡 Usage

```bash
# Example Scan Commands:
  ## Start a simple scan with default 'general' flow
  osmedeus scan -t sample.com

  ## Start a general scan but exclude some of the module
  osmedeus scan -t sample.com -x screenshot -x spider

  ## Start a scan directly with a module with inputs as a list of http domains like this https://sub.example.com
  osmedeus scan -m content-discovery -t http-file.txt

  ## Initiate the scan using a speed option other than the default setting
  osmedeus scan -f vuln --tactic gently -t sample.com
  osmedeus scan --threads-hold=10 -t sample.com
  osmedeus scan -B 5 -t sample.com

  ## Start a simple scan with other flow
  osmedeus scan -f vuln -t sample.com
  osmedeus scan -f extensive -t sample.com -t another.com
  osmedeus scan -f urls -t list-of-urls.txt

  ## Scan list of targets
  osmedeus scan -T list_of_targets.txt
  osmedeus scan -f vuln -T list-of-targets.txt

  ## Performing static vulnerability scan and secret scan on a git repo
  osmedeus scan -m repo-scan -t https://github.com/j3ssie/sample-repo
  osmedeus scan -m repo-scan -t /tmp/source-code-folder
  osmedeus scan -m repo-scan -T list-of-repo.txt

  ## Scan for CIDR with file contains CIDR with the format '1.2.3.4/24'
  osmedeus scan -f cidr -t list-of-ciders.txt
  osmedeus scan -f cidr -t '1.2.3.4/24' # this will auto convert the single input to the file and run

  ## Directly run on vuln scan and directory scan on list of domains
  osmedeus scan -f domains -t list-of-domains.txt
  osmedeus scan -f vuln-and-dirb -t list-of-domains.txt

  ## Use a custom wordlist
  osmedeus scan -t sample.com -p 'wordlists={{Data}}/wordlists/content/big.txt'

  ## Use a custom wordlist
  cat list_of_targets.txt | osmedeus scan -c 2

  ## Start a normal scan and backup entire workflow folder to the backup folder
  osmedeus scan --backup -f domains -t list-of-subdomains.txt

  ## Start the scan with chunk inputs to review the output way more much faster
  osmedeus scan --chunk --chunk-parts 20 -f cidr -t list-of-100-cidr.txt

  ## Continuously run the scan on a target right after it finished
  osmedeus utils cron --for --cmd 'osmedeus scan -t example.com'

  ## Backing up all workspaces
  ls ~/workspaces-osmedeus | osmedeus report compress


# Scan Usage:
  osmedeus scan -f [flowName] -t [target]
  osmedeus scan -m [modulePath] -T [targetsFile]
  osmedeus scan -f /path/to/flow.yaml -t [target]
  osmedeus scan -m /path/to/module.yaml -t [target] --params 'port=9200'
  osmedeus scan -m /path/to/module.yaml -t [target] -l /tmp/log.log
  osmedeus scan --tactic aggressive -m module -t [target]
  cat targets | osmedeus scan -f sample

# Practical Scan Usage:
  osmedeus scan -T list_of_targets.txt -W custom_workspaces
  osmedeus scan -t target.com -w workspace_name --debug
  osmedeus scan -f general -t sample.com
  osmedeus scan --tactic aggressive -f general -t sample.com
  osmedeus scan -f extensive -t sample.com -t another.com
  cat list_of_urls.txt | osmedeus scan -f urls
  osmedeus scan --threads-hold=15 -f cidr -t 1.2.3.4/24
  osmedeus scan -m ~/.osmedeus/core/workflow/test/dirbscan.yaml -t list_of_urls.txt
  osmedeus scan --wfFolder ~/custom-workflow/ -f your-custom-workflow -t list_of_urls.txt
  osmedeus scan --chunk --chunk-part 40 -c 2 -f cidr -t list-of-cidr.txt

💡 For full help message, please run: osmedeus --hh or osmedeus scan --hh
📖 Documentation can be found here: https://docs.osmedeus.org
```

Check out [**this page**](https://docs.osmedeus.org/installation/usage/) for full usage and the [**Practical Usage**](https://docs.osmedeus.org/installation/practical-usage/) to see how to use Osmedeus in a practical way.

## 💬 Community & Discussion

Join Our Discord server [here](https://discord.gg/mtQG2FQsYA)

## 💎 Donation & Sponsor

<h3 align="center">
 <img alt="Osmedeus" src="https://raw.githubusercontent.com/osmedeus/assets/main/premium-package.gif" />

 <p align="center"> Check out for a couple of <strong><a href="https://docs.osmedeus.org/donation/">donation methods here</a></strong> to get a <strong><a href="https://docs.osmedeus.org/premium/">premium package</a></strong><p>
</h3>

## License

`Osmedeus` is made with ♥ by [@j3ssiejjj](https://twitter.com/j3ssiejjj) and it is released under the MIT license.
