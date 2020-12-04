<p align="center">
  <img alt="Osmedeus" src="https://image.flaticon.com/icons/svg/108/108669.svg" height="140" />
  <p align="center">
    <a href="https://github.com/j3ssie/Osmedeus"><img alt="python" src="https://img.shields.io/badge/python-3.6%2B-blue.svg"></a>
    <a href=""><img alt="Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square"></a>
    <a href=""><img alt="tested" src="https://img.shields.io/badge/tested-Kali%20Linux-green.svg"></a>
    <a href="https://github.com/j3ssie/Osmedeus"><img alt="Release" src="https://img.shields.io/badge/version-2.2-red.svg"></a>
    <a href="https://www.youtube.com/watch?v=kZ-uMC7c5OY&list=PLqpLl_iGMLnA6vbi1ZM-HmWLuedIP2PJl"><img alt="Demo" src="https://img.shields.io/badge/demo-youtube-blue.svg"></a>
  </p>
</p>

## What is Osmedeus?

![Architecture](imgs/osmedeus-architecture.png)

Osmedeus allows you automated run the collection of awesome tools to reconnaissance and vulnerability scanning against the target.


## Next generation version?

<p align="center">
  <img alt="huntersuite" src="https://user-images.githubusercontent.com/23289085/101143253-35ea6b80-3649-11eb-9130-d1fc306c9a76.png" height="200" />
  <p align="center">
Enjoying this tool? Support it's development and take your game to the next level by using <a href="https://huntersuite.io">HunterSuite.io</a>
  </p>
</p>


## Installation

```
git clone https://github.com/j3ssie/Osmedeus
cd Osmedeus
./install.sh
```

This install only focus on Kali linux, check more install on [Usage page](https://j3ssie.github.io/Osmedeus/installation/).

**NOTE:** You might need to do `sudo su` before install or using this otherwise you might get issues with dependency problems.

## Using Docker

Check out [docker-osmedeus](https://github.com/mablanco/docker-osmedeus) by [mabnavarrete](https://twitter.com/mabnavarrete)

## How to use

If you have no idea what are you doing just type the command below or check out the [Advanced Usage](https://j3ssie.github.io/Osmedeus/advanced/)

```
./osmedeus.py -t example.com
```

## Features

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
- [x] React [Web UI](https://j3ssie.github.io/Osmedeus/web-ui/).
- [x] Support Continuous Scan.
- [x] Slack notifications.
- [x] Easily view [report](https://j3ssie.github.io/Osmedeus/report/) from commnad line.

Check this [Documentation](https://j3ssie.github.io/Osmedeus/modules/) for more detail about each module.

## Love Osmedeus? Back it!
If you love my tool and would like to support my work. [Become a backer](https://opencollective.com/osmedeus) and you will appear here.
[![Backers](https://opencollective.com/osmedeus/backers.svg?width=890)](https://opencollective.com/osmedeus#backers)

## Demo
[![asciicast](https://asciinema.org/a/266599.svg)](https://asciinema.org/a/266599)

![Web UI 1](imgs/ui-1.png)

![Web UI 3](imgs/ui-3.png)

![Web UI 2](imgs/ui-2.png)

## Example Commands

```
# normal routine
./osmedeus.py -t example.com
./osmedeus.py -T list_of_target.txt

# normal routine but slow speed on all moddule
./osmedeus.py -t example.com --slow 'all'

# normal routine but exclude some modules
./osmedeus.py -t example.com -x 'linkfinding,dirb'

# direct mode examples
./osmedeus.py -m subdomain -t example.com
./osmedeus.py -m portscan -i "1.2.3.4/24"

./osmedeus.py -m "portscan,vulnscan" -i "1.2.3.4/24" -w result_folder

# direct list mode examples
./osmedeus.py -m portscan -I list_of_targets.txt
./osmedeus.py -m portscan,vulnscan -I list_of_targets.txt
./osmedeus.py -m screen -I list_of_targets.txt -w result_folder

# report mode
./osmedeus.py -t example.com --report list
./osmedeus.py -t example.com --report export
./osmedeus.py -t example.com --report sum
./osmedeus.py -t example.com --report short
./osmedeus.py -t example.com --report full

```

## What to do next after do recon? Check out my scanner [Jaeles](https://github.com/jaeles-project/jaeles).

## Disclaimer

Most of this tool done by the authors of the tool that list in [CREDITS.md](https://github.com/j3ssie/Osmedeus/blob/master/CREDITS.md).
I'm just put all the pieces together, plus some extra magic.

This tool is for educational purposes only. You are responsible for your own actions. If you mess something up or break any laws while using this software, it's your fault, and your fault only.

## Credits

Please take a look at [CREDITS.md](https://github.com/j3ssie/Osmedeus/blob/master/CREDITS.md)

## Changelog

Please take a look at [CHANGELOG.md](https://github.com/j3ssie/Osmedeus/blob/master/CHANGELOG.md)

## Contributors

If you have some new idea about this project, issue, feedback or found some valuable tool feel free to open an issue for just DM me via [@j3ssiejjj](https://twitter.com/j3ssiejjj).

### Code Contributors

This project exists thanks to all the people who contribute. [[Contribute](CONTRIBUTING.md)].
<a href="https://github.com/j3ssie/Osmedeus/graphs/contributors"><img src="https://opencollective.com/Osmedeus/contributors.svg?width=890&button=false" /></a>

### Financial Contributors

Become a financial contributor and help us sustain our community. [[Contribute](https://opencollective.com/Osmedeus/contribute)]

#### Individuals

<a href="https://opencollective.com/Osmedeus"><img src="https://opencollective.com/Osmedeus/individuals.svg?width=890"></a>

#### Organizations

Support this project with your organization. Your logo will show up here with a link to your website. [[Contribute](https://opencollective.com/Osmedeus/contribute)]

<a href="https://opencollective.com/Osmedeus/organization/0/website"><img src="https://opencollective.com/Osmedeus/organization/0/avatar.svg"></a>
<a href="https://opencollective.com/Osmedeus/organization/1/website"><img src="https://opencollective.com/Osmedeus/organization/1/avatar.svg"></a>
<a href="https://opencollective.com/Osmedeus/organization/2/website"><img src="https://opencollective.com/Osmedeus/organization/2/avatar.svg"></a>
<a href="https://opencollective.com/Osmedeus/organization/3/website"><img src="https://opencollective.com/Osmedeus/organization/3/avatar.svg"></a>
<a href="https://opencollective.com/Osmedeus/organization/4/website"><img src="https://opencollective.com/Osmedeus/organization/4/avatar.svg"></a>
<a href="https://opencollective.com/Osmedeus/organization/5/website"><img src="https://opencollective.com/Osmedeus/organization/5/avatar.svg"></a>
<a href="https://opencollective.com/Osmedeus/organization/6/website"><img src="https://opencollective.com/Osmedeus/organization/6/avatar.svg"></a>
<a href="https://opencollective.com/Osmedeus/organization/7/website"><img src="https://opencollective.com/Osmedeus/organization/7/avatar.svg"></a>
<a href="https://opencollective.com/Osmedeus/organization/8/website"><img src="https://opencollective.com/Osmedeus/organization/8/avatar.svg"></a>
<a href="https://opencollective.com/Osmedeus/organization/9/website"><img src="https://opencollective.com/Osmedeus/organization/9/avatar.svg"></a>

## Stargazers over time

[![Stargazers over time](https://starchart.cc/j3ssie/Osmedeus.svg)](https://starchart.cc/j3ssie/Osmedeus)


## Donation

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://paypal.me/j3ssiejjj)
