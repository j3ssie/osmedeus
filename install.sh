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
[ -x "$(command -v git)" ] || sudo $PACKGE_MANAGER install git -y
[ -x "$(command -v nmap)" ] || sudo $PACKGE_MANAGER install nmap -y
[ -x "$(command -v masscan)" ] || sudo $PACKGE_MANAGER install masscan -y
[ -x "$(command -v chromium)" ] || sudo $PACKGE_MANAGER install chromium -y
[ -x "$(command -v npm)" ] || sudo $PACKGE_MANAGER install npm -y
[ -x "$(command -v go)" ] || sudo $PACKGE_MANAGER install golang -y

#### Download stuff directly
install_banner "wordlists"
mkdir -p $PLUGINS_PATH/wordlists/
[[ -f $PLUGINS_PATH/wordlists/all.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/all.txt https://gist.githubusercontent.com/jhaddix/86a06c5dc309d08580a018c66354a056/raw/96f4e51d96b2203f19f6381c8c545b278eaa0837/all.txt
[[ -f $PLUGINS_PATH/wordlists/shorts.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/shorts.txt https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/DNS/subdomains-top1mil-20000.txt

[[ -f $PLUGINS_PATH/wordlists/dir-all.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/dir-all.txt https://gist.githubusercontent.com/jhaddix/b80ea67d85c13206125806f0828f4d10/raw/c81a34fe84731430741e0463eb6076129c20c4c0/content_discovery_all.txt

[[ -f $PLUGINS_PATH/apps.json ]] || wget -q -O $PLUGINS_PATH/apps.json https://raw.githubusercontent.com/AliasIO/Wappalyzer/master/src/apps.json


install_banner "vulners nse"
# Install vulners nse script
[[ -f /usr/share/nmap/scripts/vulners.nse ]] ||  wget -q -O /usr/share/nmap/scripts/vulners.nse https://raw.githubusercontent.com/vulnersCom/nmap-vulners/master/vulners.nse


install_banner "nmap bootstrap"
# Install nmap bootstrap
[[ -f $PLUGINS_PATH/nmap-bootstrap.xsl ]] ||  wget -q -O $PLUGINS_PATH/nmap-bootstrap.xsl https://raw.githubusercontent.com/honze-net/nmap-bootstrap-xsl/master/nmap-bootstrap.xsl

install_banner "providers-data for subdomain takeover"
[[ -f $PLUGINS_PATH/providers-data.csv ]] ||  wget -q -O $PLUGINS_PATH/providers-data.csv https://raw.githubusercontent.com/anshumanbh/tko-subs/master/providers-data.csv
[[ -f $PLUGINS_PATH/fingerprints.json ]] ||  wget -q -O $PLUGINS_PATH/fingerprints.json https://raw.githubusercontent.com/haccer/subjack/master/fingerprints.json

##
chmod +x osmedeus.py
### adding gopath if GOPATH not in default shellrc
if ! grep -Fxq "GOPATH" "$DEFAULT_SHELL"
then
	echo 'export GOPATH=$HOME/go' >> $DEFAULT_SHELL
	echo 'PATH=$GOPATH/bin:$PATH' >> $DEFAULT_SHELL
	source $DEFAULT_SHELL
fi
PS="$ "
source $DEFAULT_SHELL

##
# Install go stuff
##
mkdir -p "$GO_DIR"
mkdir -p "$PLUGINS_PATH/go/"
install_banner "amass"
go get -u github.com/OWASP/Amass/...
install_banner "subfinder"
go get -u github.com/subfinder/subfinder
install_banner "gobuster"
go get -u github.com/OJ/gobuster
install_banner "aquatone"
go get -u github.com/michenriksen/aquatone
install_banner "gitrob"
go get -u github.com/michenriksen/gitrob
install_banner "subjack"
go get -u github.com/haccer/subjack
install_banner "tko-subs"
go get -u github.com/anshumanbh/tko-subs
install_banner "gitleaks"
go get -u github.com/zricethezav/gitleaks
install_banner "webanalyze"
go get -u github.com/rverton/webanalyze/...

cp $GO_DIR/* "$PLUGINS_PATH/go/"


#Install flask stuff
install_banner "Flask API stuff"
pip3 install flask flask_restful flask_jwt flask_cors flask_jwt_extended

install_banner "observatory"
npm install -g observatory-cli

install_banner "IPOsint"
cd $PLUGINS_PATH
git clone https://github.com/j3ssie/IPOsint
pip3 install -r IPOsint/requirements.txt
cd $CWD

#install massdns
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

##
# Install python stuff
##
install_banner "brutespray"
cd $PLUGINS_PATH
git clone https://github.com/x90skysn3k/brutespray
cd brutespray
pip install -r requirements.txt
cd $CWD

install_banner "truffleHog"
pip install truffleHog

install_banner "testssl.sh"
cd $PLUGINS_PATH
git clone https://github.com/drwetter/testssl.sh
cd $CWD

install_banner "LinkFinder"
cd $PLUGINS_PATH
git clone https://github.com/GerbenJavado/LinkFinder.git
cd LinkFinder
python setup.py install

cd $PLUGINS_PATH
install_banner "sqlmap"
git clone https://github.com/sqlmapproject/sqlmap

cd $PLUGINS_PATH
install_banner "sherlock"
git clone https://github.com/sherlock-project/sherlock.git
cd sherlock
pip3 install -r requirements.txt
cd $CWD

cd $PLUGINS_PATH
install_banner "SleuthQL"
git clone https://github.com/RhinoSecurityLabs/SleuthQL
pip install bs4
cd $CWD

cd $PLUGINS_PATH
install_banner "dirsearch"
git clone https://github.com/maurosoria/dirsearch

install_banner "dirhunt"
git clone https://github.com/Nekmo/dirhunt
cd dirhunt
python3 setup.py install
cd $CWD

install_banner "CORStest"
cd $PLUGINS_PATH
git clone https://github.com/RUB-NDS/CORStest

install_banner "JSParser"
cd $PLUGINS_PATH
git clone https://github.com/nahamsec/JSParser
cd JSParser/
python setup.py install










