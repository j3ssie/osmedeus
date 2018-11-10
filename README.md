Osmedeus
============

# What is Osmedeus?
Osmedeus allow you to doing boring stuff in Pentesting automatically like reconnaissance and scanning the target by run the collection of awesome tools.

Most of this tool done by the authors of the tool that you can see in the module folder.
I just put all the pieces together, plus some extra boring stuff that we don't wanna do everyday.


# Installation
```
git clone https://github.com/j3ssie/Osmedeus
cd Osmedeus
chmod +x install.sh
./install.sh
```

# How to use

* Doing normal routine include: Subdomain Scanning, Subdomain TakeOver Scanning, Port Scanning and ScreenShot the target.
```
./osmedeus.py -t example.com
```

* Git repo scanning
```
./osmedeus.py --git https://github.com/whatever/repo
```

* Doing some stuff with Burp State file
```
./osmedeus.py -t example.com --burp yourburpstate.xml
```


# Available modules with list tool being used
* Subdomain Scanning
  * [amass](https://github.com/caffix/amass)
  * [subfinder](https://github.com/subfinder/subfinder)
  * [massdns](https://github.com/blechschmidt/massdns)

* Subdomain TakeOver Scanning
  * [subjack](https://github.com/haccer/subjack)
  * [SubOver](https://github.com/Ice3man543/SubOver)

* Port Scanning and ScreenShot the target
  * [aquaton](https://github.com/michenriksen/aquatone)
  * [EyeWitness](https://github.com/FortyNorthSecurity/EyeWitness)
  * [masscan](https://github.com/robertdavidgraham/masscan)

* Git repo scanning
  * [truffleHog](https://github.com/dxa4481/truffleHog)
  * [gitrob](https://github.com/michenriksen/gitrob)

* Doing some stuff with Burp State file
  * [sqlmap](https://github.com/sqlmapproject/sqlmap)
  * [SleuthQL](https://github.com/RhinoSecurityLabs/SleuthQL)
  * [LinkFinder](https://github.com/GerbenJavado/LinkFinder)

# Demo
* [Asciinema]()
* [Youtube]()


# Contact
[@j3ssiejjj](https://twitter.com/j3ssiejjj)