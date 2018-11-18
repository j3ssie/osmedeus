#!/bin/bash

#global stuff
CWD=$(pwd)
WORKSPACE="$CWD/workspaces/"
PLUGINS_PATH="$CWD/plugins/"

#some function
install_banner()
{
  name=$1
  echo -e "\033[1;32m[+] Installing $name \033[1;37m"
}

install_banner "wordlists"
mkdir -p $PLUGINS_PATH/wordlists/
[[ -f $PLUGINS_PATH/wordlists/all.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/all.txt https://gist.githubusercontent.com/jhaddix/86a06c5dc309d08580a018c66354a056/raw/96f4e51d96b2203f19f6381c8c545b278eaa0837/all.txt

[[ -f $PLUGINS_PATH/wordlists/dir-all.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/dir-all.txt https://gist.githubusercontent.com/jhaddix/b80ea67d85c13206125806f0828f4d10/raw/c81a34fe84731430741e0463eb6076129c20c4c0/content_discovery_all.txt

install_banner "vulners nse"
# Install vulners nse script
[[ -f /usr/share/nmap/scripts/vulners.nse ]] ||  wget -q -O /usr/share/nmap/scripts/vulners.nse https://raw.githubusercontent.com/vulnersCom/nmap-vulners/master/vulners.nse


install_banner "nmap bootstrap"
# Install nmap bootstrap
[[ -f $PLUGINS_PATH/nmap-bootstrap.xsl ]] ||  wget -q -O $PLUGINS_PATH/nmap-bootstrap.xsl https://raw.githubusercontent.com/honze-net/nmap-bootstrap-xsl/master/nmap-bootstrap.xsl


##
chmod +x osmedeus.py

install_banner "Golang"
apt install golang -y

###
if ! (grep -Fxq "GOPATH" "$HOME/.bashrc" )
then
	echo 'export GOPATH=$HOME/go' >> ~/.bashrc
	echo 'PATH=$GOPATH/bin:$PATH' >> ~/.bashrc
	source ~/.bashrc
fi

##
# Install go stuff
##
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
install_banner "gobuster"
go get -u github.com/OJ/gobuster


install_banner "massdns"
cd $PLUGINS_PATH
git clone https://github.com/blechschmidt/massdns
cd massdns
make



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


install_banner "EyeWitness"
cd $PLUGINS_PATH
git clone https://github.com/FortyNorthSecurity/EyeWitness
cd EyeWitness
bash setup/setup.sh


install_banner "LinkFinder"
cd $PLUGINS_PATH
git clone https://github.com/GerbenJavado/LinkFinder.git
cd LinkFinder
python setup.py install

cd $PLUGINS_PATH
install_banner "sqlmap"
git clone https://github.com/sqlmapproject/sqlmap


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


install_banner "JSParser"
cd $PLUGINS_PATH
git clone https://github.com/nahamsec/JSParser
cd JSParser/
python setup.py install











