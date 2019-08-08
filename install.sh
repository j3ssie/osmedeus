#!/bin/bash

#global stuff
CWD=$(pwd)
WORKSPACE="$CWD/workspaces"
PLUGINS_PATH="$CWD/plugins"
DEFAULT_SHELL="$HOME/.bashrc"
PACKGE_MANAGER="apt-get"
GO_DIR=~/go/bin

#some function
install_banner()
{
  name=$1
  echo -e "\033[1;32m[+] Installing $name \033[1;37m"
}

install_banner "git, nmap, masscan, chromium, npm, golang"
[ -x "$(command -v git)" ] || sudo $PACKGE_MANAGER install git -y 2> /dev/null
[ -x "$(command -v nmap)" ] || sudo $PACKGE_MANAGER install nmap -y 2> /dev/null
[ -x "$(command -v masscan)" ] || sudo $PACKGE_MANAGER install masscan -y 2> /dev/null
[ -x "$(command -v chromium)" ] || sudo $PACKGE_MANAGER install chromium -y 2> /dev/null
[ -x "$(command -v npm)" ] || sudo $PACKGE_MANAGER install npm -y 2> /dev/null
[ -x "$(command -v go)" ] || sudo $PACKGE_MANAGER install golang -y 2> /dev/null
[ -x "$(command -v csvlook)" ] || sudo $PACKGE_MANAGER install csvkit -y 2> /dev/null
[ -x "$(command -v proxychains)" ] || sudo $PACKGE_MANAGER install proxychains -y 2> /dev/null

#### Download stuff directly
install_banner "wordlists"
mkdir -p $PLUGINS_PATH 2> /dev/null
mkdir -p $PLUGINS_PATH/wordlists/ 2> /dev/null
mkdir -p $PLUGINS_PATH/nmap-stuff/ 2> /dev/null
mkdir -p "$GO_DIR" 2> /dev/null
mkdir -p "$PLUGINS_PATH/go/" 2> /dev/null

[[ -f $PLUGINS_PATH/wordlists/all.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/all.txt https://gist.githubusercontent.com/jhaddix/86a06c5dc309d08580a018c66354a056/raw/96f4e51d96b2203f19f6381c8c545b278eaa0837/all.txt

# domain discovery
[[ -f $PLUGINS_PATH/wordlists/shorts.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/shorts.txt https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/DNS/subdomains-top1mil-20000.txt

[[ -f $PLUGINS_PATH/wordlists/dir-all.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/dir-all.txt https://gist.githubusercontent.com/jhaddix/b80ea67d85c13206125806f0828f4d10/raw/c81a34fe84731430741e0463eb6076129c20c4c0/content_discovery_all.txt

# permutation domain
[[ -f $PLUGINS_PATH/wordlists/short-permutation.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/short-permutation.txt https://raw.githubusercontent.com/subfinder/goaltdns/master/words.txt

[[ -f $PLUGINS_PATH/wordlists/all-permutation.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/all-permutation.txt https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/DNS/bitquark-subdomains-top100K.txt

# content discovery
[[ -f $PLUGINS_PATH/wordlists/raft-large-directories.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/raft-large-directories.txt	https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/raft-large-directories.txt


[[ -f $PLUGINS_PATH/wordlists/really-quick.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/really-quick.txt 	https://raw.githubusercontent.com/maurosoria/dirsearch/master/db/dicc.txt 

[[ -f $PLUGINS_PATH/wordlists/top10000.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/top10000.txt 	https://raw.githubusercontent.com/danielmiessler/RobotsDisallowed/master/top10000.txt


cat $PLUGINS_PATH/wordlists/really-quick.txt $PLUGINS_PATH/wordlists/top10000.txt > $PLUGINS_PATH/wordlists/quick-content-discovery.txt


## technology signature
[[ -f $PLUGINS_PATH/apps.json ]] || wget -q -O $PLUGINS_PATH/apps.json https://raw.githubusercontent.com/AliasIO/Wappalyzer/master/src/apps.json

### Nmap stuff
install_banner "nmap vulners nse"
# Install vulners nse script
[[ -f $PLUGINS_PATH/nmap-stuff/vulners.nse ]] ||  wget -q -O $PLUGINS_PATH/nmap-stuff/vulners.nse https://raw.githubusercontent.com/vulnersCom/nmap-vulners/master/vulners.nse

install_banner "nmap bootstrap"
# Install nmap bootstrap
[[ -f $PLUGINS_PATH/nmap-stuff/nmap-bootstrap.xsl ]] ||  wget -q -O $PLUGINS_PATH/nmap-stuff/nmap-bootstrap.xsl https://raw.githubusercontent.com/honze-net/nmap-bootstrap-xsl/master/nmap-bootstrap.xsl

install_banner "nmap parser"
[[ -f $PLUGINS_PATH/nmap-stuff/nmap_xml_parser.py ]] ||  wget -q -O $PLUGINS_PATH/nmap-stuff/nmap_xml_parser.py https://raw.githubusercontent.com/laconicwolf/Nmap-Scan-to-CSV/master/nmap_xml_parser.py

[[ -f $PLUGINS_PATH/nmap-stuff/masscan_xml_parser.py ]] ||  wget -q -O $PLUGINS_PATH/nmap-stuff/masscan_xml_parser.py https://raw.githubusercontent.com/laconicwolf/Masscan-to-CSV/master/masscan_xml_parser.py

install_banner "providers-data for subdomain takeover"
[[ -f $PLUGINS_PATH/providers-data.csv ]] ||  wget -q -O $PLUGINS_PATH/providers-data.csv https://raw.githubusercontent.com/anshumanbh/tko-subs/master/providers-data.csv
[[ -f $PLUGINS_PATH/fingerprints.json ]] ||  wget -q -O $PLUGINS_PATH/fingerprints.json https://raw.githubusercontent.com/haccer/subjack/master/fingerprints.json

##
chmod +x osmedeus.py
install_banner "Osmedeus dependencies"
pip3 install -r requirements.txt

### adding gopath if GOPATH not in default shellrc
if ! grep -Fxq "GOPATH" "$DEFAULT_SHELL"
then
	echo 'export GOPATH=$HOME/go' >> $DEFAULT_SHELL
	echo 'PATH=$GOPATH/bin:$PATH' >> $DEFAULT_SHELL
	source $DEFAULT_SHELL
fi
PS="$ "
source $DEFAULT_SHELL

# update golang version
install_banner "Update Golang version"
wget -qO- https://raw.githubusercontent.com/udhos/update-golang/master/update-golang.sh | bash

GO_BIN=$(which go)
[[ -f /usr/local/go/bin/go ]] && GO_BIN=/usr/local/go/bin/go

##
# Install go stuff
##
install_banner "amass"
$GO_BIN get -u github.com/OWASP/Amass/...
install_banner "subfinder"
$GO_BIN get -u github.com/subfinder/subfinder
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
install_banner "goaltdns"
$GO_BIN get -u github.com/subfinder/goaltdns
install_banner "gitleaks"
$GO_BIN get -u github.com/zricethezav/gitleaks
install_banner "gowitness"
$GO_BIN get -u github.com/sensepost/gowitness
install_banner "webanalyze"
$GO_BIN get -u github.com/rverton/webanalyze/...
install_banner "waybackurls"
$GO_BIN get -u github.com/tomnomnom/waybackurls
install_banner "meg"
$GO_BIN get -u github.com/tomnomnom/meg
install_banner "httprobe"
$GO_BIN get -u github.com/tomnomnom/httprobe

install_banner "gf"
$GO_BIN get -u github.com/tomnomnom/gf

cp $GO_DIR/* "$PLUGINS_PATH/go/" 2> /dev/null
install_banner "observatory"
npm install -g observatory-cli

# install massdns
install_banner "massdns"
cd $PLUGINS_PATH
git clone https://github.com/blechschmidt/massdns
cd massdns
if [ "$(uname)" == "Darwin" ]; then
    make nolinux
else
	make
fi
cd $CWD

# install dirble
mkdir -p $PLUGINS_PATH/dirble/ 2> /dev/null
if [ "$(uname)" == "Darwin" ]; then
    curl -s https://api.github.com/repos/nccgroup/dirble/releases/latest | grep "x86_64-apple-darwin.zip" | cut -d '"' -f 4 | wget -O $PLUGINS_PATH/dirble/dirble-release.zip -qi -
else
	curl -s https://api.github.com/repos/nccgroup/dirble/releases/latest | grep "x86_64-linux.zip" | cut -d '"' -f 4 | wget -O $PLUGINS_PATH/dirble/dirble-release.zip -qi -
fi
unzip $PLUGINS_PATH/dirble/dirble-release.zip 2> /dev/null

##
# Install python stuff
##

install_banner "truffleHog, wfuzz"
pip install truffleHog
pip3 install wfuzz

cd $PLUGINS_PATH
install_banner "gf singature"
git clone https://github.com/tomnomnom/gf 2> /dev/null

install_banner "testssl.sh"
git clone https://github.com/drwetter/testssl.sh 2> /dev/null

install_banner "IPOsint"
git clone https://github.com/j3ssie/IPOsint 2> /dev/null
pip3 install -r IPOsint/requirements.txt

install_banner "dirsearch"
git clone https://github.com/maurosoria/dirsearch 2> /dev/null

install_banner "LinkFinder"
git clone https://github.com/GerbenJavado/LinkFinder.git 2> /dev/null
python2.7 LinkFinder/setup.py install

install_banner "sherlock"
git clone https://github.com/sherlock-project/sherlock.git 2> /dev/null
pip3 install -r sherlock/requirements.txt

install_banner "CORStest"
git clone https://github.com/RUB-NDS/CORStest 2> /dev/null

install_banner "brutespray"
git clone https://github.com/x90skysn3k/brutespray 2> /dev/null
pip install -r brutespray/requirements.txt

cd $CWD
echo -e "\033[1;32m[+] Installing done... \033[1;37m"
python3 osmedeus.py --reset





