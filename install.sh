#!/bin/bash

#global stuff
CWD=$(pwd)
WORKSPACE="$CWD/workspaces/"
PLUGINS_PATH="$CWD/plugins/"

echo -e "\033[1;32m[+] Downloading wordlists \033[1;37m"
mkdir -p $PLUGINS_PATH/wordlists/
[[ -f $PLUGINS_PATH/wordlists/all.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/all.txt https://gist.githubusercontent.com/jhaddix/86a06c5dc309d08580a018c66354a056/raw/96f4e51d96b2203f19f6381c8c545b278eaa0837/all.txt

[[ -f $PLUGINS_PATH/wordlists/dir-all.txt ]] || wget -q -O $PLUGINS_PATH/wordlists/dir-all.txt https://gist.githubusercontent.com/jhaddix/b80ea67d85c13206125806f0828f4d10/raw/c81a34fe84731430741e0463eb6076129c20c4c0/content_discovery_all.txt


##
chmod +x osmedeus.py
echo "\033[1;32m[+] Installing Golang"

apt install golang -y
if ! (grep -Fxq "~/.bashrc" "GOPATH")
then
	echo 'export GOPATH=$HOME/go' >> ~/.bashrc
	echo 'PATH=$GOPATH/bin:$PATH' >> ~/.bashrc
	source ~/.bashrc
fi

##
# Install go stuff
##
echo -e "\033[1;32m[+] Installing amass \033[1;37m"
go get -u github.com/OWASP/Amass/...
echo -e "\033[1;32m[+] Installing subfinder \033[1;37m"
go get github.com/subfinder/subfinder
echo -e "\033[1;32m[+] Installing gobuster \033[1;37m"
go get github.com/OJ/gobuster
echo -e "\033[1;32m[+] Installing aquatone \033[1;37m"
go get github.com/michenriksen/aquatone
echo -e "\033[1;32m[+] Installing gitrob \033[1;37m"
go get github.com/michenriksen/gitrob

echo -e "\033[1;32m[+] Installing subjack \033[1;37m"
go get github.com/haccer/subjack

echo -e "\033[1;32m[+] Installing gobuster \033[1;37m"
go get github.com/OJ/gobuster
# echo -e "\033[1;32m[+] Installing SubOver \033[1;37m"
# go get github.com/haccer/subjack


echo -e "\033[1;32m[+] Installing massdns \033[1;37m"
cd $PLUGINS_PATH
git clone https://github.com/blechschmidt/massdns
cd massdns
make


##
# Install python stuff
##
cd $PLUGINS_PATH
git clone https://github.com/x90skysn3k/brutespray
cd brutespray
pip install -r requirements.txt
cd $CWD



echo -e "\033[1;32m[+] Installing truffleHog \033[1;37m"
pip install truffleHog


echo -e "\033[1;32m[+] Installing EyeWitness \033[1;37m"
cd $PLUGINS_PATH
git clone https://github.com/FortyNorthSecurity/EyeWitness
cd EyeWitness
bash setup/setup.sh


echo -e "\033[1;32m[+] Installing LinkFinder \033[1;37m"
cd $PLUGINS_PATH
git clone https://github.com/GerbenJavado/LinkFinder.git
cd LinkFinder
python setup.py install

cd $PLUGINS_PATH
echo -e "\033[1;32m[+] Installing Sqlmap \033[1;37m"
git clone https://github.com/sqlmapproject/sqlmap


echo -e "\033[1;32m[+] Installing SleuthQL \033[1;37m"
git clone https://github.com/RhinoSecurityLabs/SleuthQL
pip install bs4
cd $CWD

cd $PLUGINS_PATH
echo -e "\033[1;32m[+] Installing dirsearch \033[1;37m"
git clone https://github.com/maurosoria/dirsearch

echo -e "\033[1;32m[+] Installing dirhunt \033[1;37m"
git clone https://github.com/Nekmo/dirhunt
cd dirhunt
python3 setup.py install
cd $CWD


echo -e "\033[1;32m[+] Installing JSParser \033[1;37m"
cd $PLUGINS_PATH
git clone https://github.com/nahamsec/JSParser
cd JSParser/
python setup.py install











