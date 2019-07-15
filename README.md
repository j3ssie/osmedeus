<p align="center">
  <img alt="Osmedeus" src="https://image.flaticon.com/icons/svg/108/108669.svg" height="140" />
  <p align="center">
    <a href="https://github.com/j3ssie/Osmedeus"><img alt="python" src="https://img.shields.io/badge/python-3.6%2B-blue.svg"></a>
    <a href=""><img alt="Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square"></a>
    <a href=""><img alt="tested" src="https://img.shields.io/badge/tested-Kali%20Linux-green.svg"></a>
    <a href="https://github.com/j3ssie/Osmedeus"><img alt="Release" src="https://img.shields.io/badge/version-1.3-red.svg"></a>
    <a href="https://youtu.be/8L5BAyyrNc0"><img alt="Demo" src="https://img.shields.io/badge/demo-youtube-blue.svg"></a>
  </p>
</p>


# What is Osmedeus?

![Architecture](https://raw.githubusercontent.com/j3ssie/Osmedeus/master/imgs/Osmedeus-architecture.png)

Osmedeus allows you automated run the collection of awesome tools to reconnaissance and vulnerability scanning against the target.

# Installation
```
git clone https://github.com/j3ssie/Osmedeus
cd Osmedeus
./install.sh
```
This install only focus on Kali linux, check more install on [Wiki page](https://github.com/j3ssie/Osmedeus/wiki)

# How to use
If you have no idea what are you doing just type the command below or check out the [Advanced Usage](https://github.com/j3ssie/Osmedeus/wiki/Advanced-Usage)
```
./osmedeus.py -t example.com
```

# Using Docker 
Check out [docker-osmedeus](https://github.com/mablanco/docker-osmedeus) by [mabnavarrete](https://twitter.com/mabnavarrete) for docker installation and this [wiki](https://github.com/j3ssie/Osmedeus/wiki/Using-Docker) for more detail.

# Features
- [x] Subdomain Scan.
- [x] Subdomain TakeOver Scan.
- [x] Screenshot the target.
- [x] Basic recon like Whois, Dig info.
- [x] Web Technology detection.
- [x] IP Discovery.
- [x] CORS Scan.
- [x] SSL Scan.
- [x] Wayback Machine Discovery.
- [x] URL Discovery.
- [x] Headers Scan.
- [x] Port Scan.
- [x] Vulnerable Scan.
- [x] Seperate workspaces to store all scan output and details logging.
- [x] REST API.
- [x] React [Web UI](https://github.com/j3ssie/Osmedeus/wiki/Web-UI).
- [x] Support Continuous Scan.
- [x] Slack notifications.

Check this [Wiki page](https://github.com/j3ssie/Osmedeus/wiki/Module-In-Depth) for more detail about each module.

# Demo
[![asciicast](https://asciinema.org/a/230164.svg)](https://asciinema.org/a/230164)

![Slack Noti](https://raw.githubusercontent.com/j3ssie/Osmedeus/master/imgs/slack_noti.png)

![Slack Report](https://raw.githubusercontent.com/j3ssie/Osmedeus/master/imgs/slack_report.png)

![Web UI 1](https://raw.githubusercontent.com/j3ssie/Osmedeus/master/imgs/osmedeus-1.png)

![Web UI 2](https://raw.githubusercontent.com/j3ssie/Osmedeus/master/imgs/osmedeus-2.png)

![Web UI 3](https://raw.githubusercontent.com/j3ssie/Osmedeus/master/imgs/osmedeus-3.png)


# Example Commands
```
# normal routine
./osmedeus.py -t example.com

# normal routine but slow speed on subdomain module
./osmedeus.py -t example.com --slow 'subdomain'

# direct mode examples
./osmedeus.py -m portscan -i "1.2.3.4/24"

./osmedeus.py -m portscan -I list_of_targets.txt -t result_folder

./osmedeus.py -m "portscan,vulnscan" -i "1.2.3.4/24" -t result_folder

./osmedeus.py -m "assets" -i "example.com"
./osmedeus.py -m "assets,dirb" -i "example.com"

```

# More options

```
Basic Usage
===========
python3 osmedeus.py -t <your_target>
python3 osmedeus.py -T <list_of_targets>
python3 osmedeus.py -m <module> [-i <input>|-I <input_file>] [-t workspace_name]

Advanced Usage
==============
[*] List all module
python3 osmedeus.py -M

[*] Running with specific module
python3 osmedeus.py -t <result_folder> -m <module_name> -i <your_target>

[*] Example command
python3 osmedeus.py -m subdomain -t example.com
python3 osmedeus.py -t example.com --slow "subdomain"
python3 osmedeus.py -t sample2 -m vuln -i hosts.txt
python3 osmedeus.py -t sample2 -m dirb -i /tmp/list_of_hosts.txt

Remote Options
==============
--remote REMOTE       Remote address for API, (default: https://127.0.0.1:5000)
--auth AUTH           Specify authentication e.g: --auth="username:password"
                      See your config file for more detail (default: core/config.conf)

--client              just run client stuff in case you ran the flask server before

More options
==============
--update              Update lastest from git

-c CONFIG, --config CONFIG
                      Specify config file (default: core/config.conf)

-w WORKSPACE, --workspace WORKSPACE
                      Custom workspace folder

-f, --force           force to run the module again if output exists
-s, --slow  "all"
                      All module running as slow mode
-s, --slow  "subdomain"
                      Only running slow mode in subdomain module

--debug               Just for debug purpose

```


# Disclaimer
Most of this tool done by the authors of the tool that list in [CREDITS.md](https://github.com/j3ssie/Osmedeus/blob/master/CREDITS.md).
I'm just put all the pieces together, plus some extra magic.

This tool is for educational purposes only. You are responsible for your own actions. If you mess something up or break any laws while using this software, it's your fault, and your fault only.

# Contribute
Please take a look at [CONTRIBUTING.md](https://github.com/j3ssie/Osmedeus/blob/master/CONTRIBUTING.md)

# Changelog
Please take a look at [CHANGELOG.md](https://github.com/j3ssie/Osmedeus/blob/master/CHANGELOG.md)

# CREDITS
Please take a look at [CREDITS.md](https://github.com/j3ssie/Osmedeus/blob/master/CREDITS.md)


# Contact
[@j3ssiejjj](https://twitter.com/j3ssiejjj)
