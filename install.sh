#!/bin/bash

# global stuff
CWD=$(pwd)
PLUGINS_PATH="$CWD/plugins"
DATA_PATH="$CWD/data"
DEFAULT_SHELL="$HOME/.bashrc"
PACKGE_MANAGER="apt-get"
GO_DIR=~/go/bin

# some function
install_banner() {
    name=$1
    echo -e "\033[1;32m[+] Installing $name \033[1;37m"
}

install_banner "git, nmap, masscan, chromium, npm, golang"
[ -x "$(command -v git)" ] || sudo $PACKGE_MANAGER install git -y 2>/dev/null
[ -x "$(command -v nmap)" ] || sudo $PACKGE_MANAGER install nmap -y 2>/dev/null
[ -x "$(command -v masscan)" ] || sudo $PACKGE_MANAGER install masscan -y 2>/dev/null
[ -x "$(command -v chromium)" ] || sudo $PACKGE_MANAGER install chromium -y 2>/dev/null
[ -x "$(command -v make)" ] || sudo $PACKGE_MANAGER install build-essential -y 2>/dev/null
[ -x "$(command -v csvlook)" ] || sudo $PACKGE_MANAGER install csvkit -y 2>/dev/null
[ -x "$(command -v rg)" ] || sudo $PACKGE_MANAGER install ripgrep -y 2>/dev/null
[ -x "$(command -v unzip)" ] || sudo $PACKGE_MANAGER install unzip -y 2>/dev/null
[ -x "$(command -v chromium-browser)" ] || sudo $PACKGE_MANAGER install chromium-browser -y 2>/dev/null
[ -x "$(command -v xsltproc)" ] || sudo $PACKGE_MANAGER install xsltproc -y 2>/dev/null
[ -x "$(command -v make)" ] || sudo $PACKGE_MANAGER install build-essential -y 2>/dev/null
[ -x "$(command -v pip)" ] || sudo $PACKGE_MANAGER install python-pip -y 2>/dev/null
[ -x "$(command -v pip3)" ] || sudo $PACKGE_MANAGER install python3-pip -y 2>/dev/null

pip install setuptools 2>/dev/null
pip3 install setuptools 2>/dev/null
pip install wheel 2>/dev/null
pip3 install wheel 2>/dev/null

#### Download stuff directly
install_banner "wordlists"
mkdir -p $DATA_PATH 2>/dev/null
mkdir -p $DATA_PATH/wordlists/ 2>/dev/null
mkdir -p $DATA_PATH/wordlists/dns/ 2>/dev/null
mkdir -p $DATA_PATH/wordlists/content/ 2>/dev/null
mkdir -p $DATA_PATH/wordlists/params/ 2>/dev/null
mkdir -p $PLUGINS_PATH 2>/dev/null
mkdir -p $DATA_PATH/nmap-stuff/ 2>/dev/null
mkdir -p $PLUGINS_PATH/nmap-stuff/ 2>/dev/null
mkdir -p "$GO_DIR" 2>/dev/null
mkdir -p "$PLUGINS_PATH/go/" 2>/dev/null

# domain discovery

[[ -f $DATA_PATH/wordlists/dns/all.txt ]] || wget -q -O $DATA_PATH/wordlists/dns/all.txt https://gist.githubusercontent.com/jhaddix/86a06c5dc309d08580a018c66354a056/raw/96f4e51d96b2203f19f6381c8c545b278eaa0837/all.txt

[[ -f $DATA_PATH/wordlists/dns/commonspeak2-subdomains.txt ]] || wget -q -O $DATA_PATH/wordlists/dns/commonspeak2-subdomains.txt https://raw.githubusercontent.com/assetnote/commonspeak2-wordlists/master/subdomains/subdomains.txt

[[ -f $DATA_PATH/wordlists/dns/shorts.txt ]] || wget -q -O $DATA_PATH/wordlists/dns/shorts.txt https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/DNS/subdomains-top1million-20000.txt

# permutation domain
[[ -f $DATA_PATH/wordlists/dns/short-permutation.txt ]] || wget -q -O $DATA_PATH/wordlists/dns/short-permutation.txt https://raw.githubusercontent.com/subfinder/goaltdns/master/words.txt

# vhost domain
[[ -f $DATA_PATH/wordlists/dns/virtual-host-scanning.txt ]] || wget -q -O $DATA_PATH/wordlists/dns/virtual-host-scanning.txt https://raw.githubusercontent.com/codingo/VHostScan/master/VHostScan/wordlists/virtual-host-scanning.txt

# content discovery
[[ -f $DATA_PATH/wordlists/content/raft-large-directories.txt ]] || wget -q -O $DATA_PATH/wordlists/content/raft-large-directories.txt https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/raft-large-directories.txt

[[ -f $DATA_PATH/wordlists/content/quick.txt ]] || wget -q -O $DATA_PATH/wordlists/content/quick.txt https://raw.githubusercontent.com/maurosoria/dirsearch/master/db/dicc.txt


[[ -f $DATA_PATH/wordlists/content/top10000.txt ]] || wget -q -O $DATA_PATH/wordlists/content/top10000.txt https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/RobotsDisallowed-Top1000.txt

cat $DATA_PATH/wordlists/content/quick.txt $DATA_PATH/wordlists/content/top10000.txt > $DATA_PATH/wordlists/content/quick-content-discovery.txt

[[ -f $DATA_PATH/wordlists/content/dir-all.txt ]] || wget -q -O $DATA_PATH/wordlists/content/dir-all.txt https://gist.githubusercontent.com/jhaddix/b80ea67d85c13206125806f0828f4d10/raw/c81a34fe84731430741e0463eb6076129c20c4c0/content_discovery_all.txt

# params
[[ -f $DATA_PATH/wordlists/params/param-miner.txt ]] || wget -q -O $DATA_PATH/wordlists/params/param-miner.txt https://raw.githubusercontent.com/PortSwigger/param-miner/master/resources/params

[[ -f $DATA_PATH/wordlists/params/parameth.txt ]] || wget -q -O $DATA_PATH/wordlists/params/parameth.txt https://raw.githubusercontent.com/maK-/parameth/master/lists/all.txt

cat $DATA_PATH/wordlists/params/param-miner.txt $DATA_PATH/wordlists/params/parameth.txt | sort -u > $DATA_PATH/wordlists/params/all.txt

# Subdomain takeover signature
install_banner "providers-data for subdomain takeover"
[[ -f $DATA_PATH/providers-data.csv ]] || wget -q -O $DATA_PATH/providers-data.csv https://raw.githubusercontent.com/anshumanbh/tko-subs/master/providers-data.csv

[[ -f $DATA_PATH/fingerprints.json ]] || wget -q -O $DATA_PATH/fingerprints.json https://raw.githubusercontent.com/haccer/subjack/master/fingerprints.json

# secret words to grep
[[ -f $DATA_PATH/keywords.txt ]] || wget -q -O $DATA_PATH/keywords.txt https://raw.githubusercontent.com/random-robbie/keywords/master/keywords.txt

# resolvers
[[ -f $DATA_PATH/resolvers.txt ]] || wget -q -O $DATA_PATH/resolvers.txt https://raw.githubusercontent.com/Abss0x7tbh/bass/master/resolvers/public.txt



#####
# Start of nmap stuff
####

## technology signature
[[ -f $DATA_PATH/apps.json ]] || wget -q -O $DATA_PATH/apps.json https://raw.githubusercontent.com/AliasIO/Wappalyzer/master/src/technologies.json

### Nmap stuff
install_banner "nmap vulners nse"
# Install vulners nse script
[[ -f $PLUGINS_PATH/nmap-stuff/vulners.nse ]] || wget -q -O $PLUGINS_PATH/nmap-stuff/vulners.nse https://raw.githubusercontent.com/vulnersCom/nmap-vulners/master/vulners.nse

install_banner "nmap bootstrap"
# Install nmap bootstrap
[[ -f $PLUGINS_PATH/nmap-stuff/nmap-bootstrap.xsl ]] || wget -q -O $PLUGINS_PATH/nmap-stuff/nmap-bootstrap.xsl https://raw.githubusercontent.com/honze-net/nmap-bootstrap-xsl/master/nmap-bootstrap.xsl

install_banner "nmap & masscan parser"
[[ -f $PLUGINS_PATH/nmap-stuff/masscan_xml_parser.py ]] || wget -q -O $PLUGINS_PATH/nmap-stuff/masscan_xml_parser.py https://raw.githubusercontent.com/laconicwolf/Masscan-to-CSV/master/masscan_xml_parser.py

[[ -f $PLUGINS_PATH/nmap-stuff/nmaptocsv.py ]] || wget -q -O $PLUGINS_PATH/nmap-stuff/nmaptocsv.py https://raw.githubusercontent.com/maaaaz/nmaptocsv/master/nmaptocsv.py

##
chmod +x osmedeus.py
install_banner "Osmedeus dependencies"
pip3 install -r requirements.txt
cd $CWD
mkdir -p ~/.osmedeus 2>/dev/null
install_banner "Initial config for Osmedeus"
python3 server/manage.py makemigrations
python3 server/manage.py migrate
python3 server/manage.py makemigrations api
python3 server/manage.py migrate api
python3 scripts/init.py
python3 scripts/reload.py

### adding gopath if GOPATH not in default shellrc
if ! grep -Fxq "GOPATH" "$DEFAULT_SHELL"; then
    echo 'export GOPATH=$HOME/go' >>$DEFAULT_SHELL
    echo 'PATH=$GOPATH/bin:$PATH' >>$DEFAULT_SHELL
    source $DEFAULT_SHELL
fi
PS="$ "
source $DEFAULT_SHELL

## Install latest go binary
cd $CWD
# update golang version
install_banner "Install Golang latest version"
wget -q -O - https://raw.githubusercontent.com/canha/golang-tools-install-script/master/goinstall.sh | bash

GO_BIN="$HOME/.go/bin/go"
# in case the script fail
[[ -f $GO_BIN ]] || GO_BIN=$(which go)
echo -e "\033[1;32m[+] Detected go binary: $GO_BIN \033[0m"
[[ -d $GO_DIR ]] || GO_DIR=$GOPATH/bin
echo -e "\033[1;32m[+] Detected go tools: $GO_DIR \033[0m"

GOROOT=$HOME/.go
PATH=$GOROOT/bin:$PATH
GOPATH=$HOME/go
PATH=$GOPATH/bin:$PATH

export GO_BIN="$HOME/.go/bin/go"
export GOROOT=$HOME/.go
export PATH=$GOROOT/bin:$PATH
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$PATH
cd $CWD

##
# Install go stuff
##
install_banner "subfinder"
$GO_BIN get -u github.com/projectdiscovery/subfinder/cmd/subfinder
install_banner "gobuster"
$GO_BIN get -u github.com/OJ/gobuster
install_banner "aquatone"
$GO_BIN get -u github.com/michenriksen/aquatone
install_banner "gitrob"
$GO_BIN get -u github.com/michenriksen/gitrob
install_banner "subjack"
$GO_BIN get -u github.com/haccer/subjack
install_banner "tko-subs"
$GO_BIN get -u github.com/anshumanbh/tko-subs
install_banner "subzy"
$GO_BIN get -u github.com/lukasikic/subzy
install_banner "goaltdns"
$GO_BIN get -u github.com/subfinder/goaltdns
install_banner "gitleaks"
$GO_BIN get -u github.com/zricethezav/gitleaks
install_banner "gowitness"
$GO_BIN get -u github.com/sensepost/gowitness
install_banner "webanalyze"
$GO_BIN get -u github.com/rverton/webanalyze/...
install_banner "assetfinder"
$GO_BIN get -u github.com/tomnomnom/assetfinder
install_banner "waybackurls"
$GO_BIN get -u github.com/tomnomnom/waybackurls
install_banner "meg"
$GO_BIN get -u github.com/tomnomnom/meg
install_banner "httprobe"
$GO_BIN get -u github.com/tomnomnom/httprobe
install_banner "unfurl"
$GO_BIN get -u github.com/tomnomnom/unfurl
install_banner "filter-resolved"
$GO_BIN get -u github.com/tomnomnom/hacks/filter-resolved
install_banner "ffuf"
$GO_BIN get -u github.com/ffuf/ffuf
install_banner "metabigor"
$GO_BIN get -u github.com/j3ssie/metabigor
$GO_BIN get -u github.com/jaeles-project/gospider
install_banner "go cli-utils"
$GO_BIN get -u github.com/j3ssie/go-auxs/just-resolved

install_banner "amass"
GO111MODULE=on $GO_BIN get github.com/OWASP/Amass/v3/... 2>/dev/null
install_banner "jaeles"
GO111MODULE=on $GO_BIN get github.com/jaeles-project/jaeles 2>/dev/null

cp $GO_DIR/* "$PLUGINS_PATH/go/" 2>/dev/null
# install_banner "observatory"
# npm install -g observatory-cli 2>/dev/null

# install massdns
install_banner "massdns"
cd $PLUGINS_PATH
git clone https://github.com/blechschmidt/massdns
cd massdns
if [[ "$OSTYPE" == "darwin"* ]]; then
    make nolinux
else
    make
fi
cd $CWD

# findomain
install_banner "findomain"
if [[ "$OSTYPE" == "darwin"* ]]; then
    wget -q -O $PLUGINS_PATH/findomain https://github.com/Edu4rdSHL/findomain/releases/latest/download/findomain-osx
else
    wget -q -O $PLUGINS_PATH/findomain https://github.com/Edu4rdSHL/findomain/releases/latest/download/findomain-linux
fi
chmod +x $PLUGINS_PATH/findomain

##
# Install python stuff
##

install_banner "truffleHog"
pip3 install truffleHog

cd $PLUGINS_PATH

# install_banner "testssl.sh"
# git clone https://github.com/drwetter/testssl.sh 2>/dev/null


install_banner "Arjun"
git clone https://github.com/s0md3v/Arjun 2>/dev/null

install_banner "CORStest"
git clone https://github.com/RUB-NDS/CORStest 2>/dev/null

echo -e "\033[1;32m[+] Installing done... \033[1;37m"
