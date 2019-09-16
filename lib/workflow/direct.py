
class SubdomainScanning:
    reports = [
        {
            "path": "$WORKSPACE/subdomain/final-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        }
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "Amass",
                "cmd": "$GO_PATH/amass enum -active -max-dns-queries 50000 -dir $WORKSPACE/subdomain/amass-$OUTPUT -d $TARGET -o $WORKSPACE/subdomain/$OUTPUT-amass.txt",
                "output_path": "$WORKSPACE/subdomain/$OUTPUT-amass.txt",
                "std_path": "$WORKSPACE/subdomain/std-$TARGET-amass.std"
            },
            {
                "banner": "Subfinder",
                "cmd": "$GO_PATH/subfinder -d $TARGET -t 100 -o $WORKSPACE/subdomain/$OUTPUT-subfinder.txt -nW",
                "output_path": "$WORKSPACE/subdomain/$OUTPUT-subfinder.txt",
                "std_path": "$WORKSPACE/subdomain/std-$OUTPUT-subfinder.std"
            },
            {
                "banner": "assetfinder",
                "cmd": "$GO_PATH/assetfinder -subs-only $TARGET | tee $WORKSPACE/subdomain/$OUTPUT-assetfinder.txt",
                "output_path": "$WORKSPACE/subdomain/$OUTPUT-assetfinder.txt",
                "std_path": "$WORKSPACE/subdomain/std-$OUTPUT-assetfinder.std"
            },
            {
                "banner": "findomain",
                "cmd": "$PLUGINS_PATH/findomain -t $TARGET -i | tee $WORKSPACE/subdomain/$OUTPUT-findomain.txt",
                "output_path": "$WORKSPACE/subdomain/$OUTPUT-findomain.txt",
                "std_path": "$WORKSPACE/subdomain/std-$OUTPUT-findomain.std",
                "post_run": "clean_findomain",
                "cleaned_output": "$WORKSPACE/subdomain/$OUTPUT-gobuster.txt",
            },
            {
                "banner": "gobuster",
                "cmd": "$GO_PATH/gobuster dns -q -t 100 -w $DATA_PATH/wordlists/dns/shorts.txt -d $TARGET -o $WORKSPACE/subdomain/raw-$OUTPUT-gobuster.txt",
                "output_path": "$WORKSPACE/subdomain/raw-$OUTPUT-gobuster.txt",
                "std_path": "$WORKSPACE/subdomain/std-raw-$OUTPUT-gobuster.std",
                "post_run": "clean_gobuster",
                "cleaned_output": "$WORKSPACE/subdomain/$OUTPUT-gobuster.txt",
            },
        ],
        'slow': [
            {
                "banner": "massdns",
                "cmd": "$PLUGINS_PATH/massdns/scripts/subbrute.py $DATA_PATH/wordlists/dns/all.txt $TARGET | $PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -q -t A -o S -w $WORKSPACE/subdomain/raw-massdns.txt",
                "output_path": "$WORKSPACE/subdomain/raw-massdns.txt",
                "std_path": "$WORKSPACE/subdomain/std-raw-massdns.txt",
                "post_run": "clean_massdns",
                "cleaned_output": "$WORKSPACE/subdomain/$OUTPUT-massdns.txt",
            },

        ],
    }


class VhostScan:
    note = "Pro-only"
    reports = [
        {
            "path": "$WORKSPACE/vhosts/vhost-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        }
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "Gobuster Vhost",
                "cmd": "$GO_PATH/gobuster vhost -U 'Mozilla/5.0 (Windows NT 5.1; rv:11.0) Gecko Firefox/11.0 (via ggpht.com GoogleImageProxy)' -z -q -t 40 -u '[[0]]' -w $DATA_PATH/wordlists/dns/virtual-host-scanning.txt --insecuressl --username admin --password admin --output $WORKSPACE/vhosts/raw/[[0]]-gobuster.txt",
                "output_path": "",
                "std_path": "",
                "chunk": 5,
                "cmd_type": "list",
                "resources": "l0|$WORKSPACE/subdomain/final-$OUTPUT.txt",
                "post_run": "clean_multi_gobuster",
                "cleaned_output": "$WORKSPACE/vhosts/vhosts-$OUTPUT.txt",
            },
        ],
    }


class PermutationScan:
    note = "Pro-only"
    reports = [
        {
            "path": "$WORKSPACE/permutation/permutation-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        }
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "goaltdns",
                "cmd": "$GO_PATH/goaltdns -w $DATA_PATH/wordlists/dns/short-permutation.txt -l $WORKSPACE/subdomain/final-$OUTPUT.txt -o $WORKSPACE/permutation/permutation-$OUTPUT.txt",
                "output_path": "$WORKSPACE/permutation/permutation-$OUTPUT.txt",
                "std_path": "",
            },
        ],
    }


class Probing:
    reports = [
        {
            "path": "$WORKSPACE/probing/ip-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/probing/really-final-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/probing/resolved-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/probing/http-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "massdns resolve IP",
                "requirement": "$WORKSPACE/probing/raw-all-$OUTPUT.txt",
                "cmd": "cat $WORKSPACE/probing/raw-all-$OUTPUT.txt | $PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -q -t A -o S -w $WORKSPACE/probing/raw-allmassdns.txt",
                "output_path": "$WORKSPACE/probing/raw-allmassdns.txt",
                "std_path": "",
                "pre_run": "get_subdomains",
                "post_run": "clean_massdns",
                "cleaned_output": "$WORKSPACE/probing/ip-$OUTPUT.txt",
                "waiting": "first",
            },
            {
                "banner": "httprobe",
                "requirement": "$WORKSPACE/probing/raw-all-$OUTPUT.txt",
                "cmd": "cat $WORKSPACE/probing/really-final-$OUTPUT.txt | $GO_PATH/httprobe -c 100 -t 20000 | tee $WORKSPACE/probing/http-$OUTPUT.txt",
                "output_path": "$WORKSPACE/probing/http-$OUTPUT.txt",
                "std_path": "$WORKSPACE/probing/std-http-$OUTPUT.std",
                # "waiting": "last",
                "post_run": "get_domain",
                "cleaned_output": "$WORKSPACE/probing/domain-$OUTPUT.txt",
            },
        ],
    }


class CORSScan:
    reports = [
        {
            "path": "$WORKSPACE/cors/$OUTPUT-corstest.txt",
            "type": "bash",
            "note": "final",
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                "requirement": "$WORKSPACE/probing/http-$OUTPUT.txt",
                "banner": "CORS Scan",
                "cmd": "python2 $PLUGINS_PATH/CORStest/corstest.py -p 50 $WORKSPACE/probing/http-$OUTPUT.txt | tee $WORKSPACE/cors/$OUTPUT-corstest.txt",
                "output_path": "$WORKSPACE/cors/$TARGET-corstest.txt",
                "std_path": "$WORKSPACE/cors/std-$TARGET-corstest.std",
            }
        ],
    }


class LinkFinding:
    reports = [
        {
            "path": "$WORKSPACE/links/summary-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/links/raw-wayback-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        }
    ]
    logs = []
    commands = {
        'general': [
            {
                "requirement": "$WORKSPACE/probing/resolved-$OUTPUT.txt",
                "banner": "waybackurls",
                "cmd": "cat $WORKSPACE/probing/resolved-$OUTPUT.txt | $GO_PATH/waybackurls | tee $WORKSPACE/links/raw-wayback-$OUTPUT.txt",
                "output_path": "$WORKSPACE/links/raw-wayback-$OUTPUT.txt",
                "std_path": "$WORKSPACE/links/std-wayback-$OUTPUT.std",
                "post_run": "clean_waybackurls",
                "cleaned_output": "$WORKSPACE/links/waybackurls-$OUTPUT.txt",
            },
            {
                "requirement": "$WORKSPACE/probing/http-$OUTPUT.txt",
                "pre_run": "get_domains",
                "banner": "linkfinder",
                "cmd": "$ALIAS_PATH/linkfinding -i '[[0]]' -o '$WORKSPACE/links/raw/' -s '$WORKSPACE/links/summary-$OUTPUT.txt' -p '$PLUGINS_PATH'",
                "output_path": "$WORKSPACE/links/raw/[[0]]-$OUTPUT.txt",
                "std_path": "",
                "chunk": 5,
                "cmd_type": "list",
                "resources": "l0|$WORKSPACE/probing/http-$OUTPUT.txt",
                "post_run": "clean_linkfinder",
                "cleaned_output": "$WORKSPACE/links/summary-$OUTPUT.txt",
            },
            {
                "requirement": "$WORKSPACE/links/raw-wayback-$OUTPUT.txt",
                "banner": "Formatting Input",
                "cmd": "cat $WORKSPACE/links/raw-wayback-$OUTPUT.txt | unfurl -u format %d%p",
                "output_path": "$WORKSPACE/links/$OUTPUT-paths.txt",
                "std_path": "",
                "waiting": "last",
            },
        ],
    }


class PortScan:
    reports = [
        {
            "path": "$WORKSPACE/portscan/final-$OUTPUT.html",
            "type": "html",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/portscan/screenshot-$OUTPUT.html",
            "type": "html",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/portscan/$OUTPUT-masscan.csv",
            "type": "bash",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/portscan/screenshot/$OUTPUT-raw-gowitness.html",
            "type": "html",
            "note": "",
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                # "requirement": "$WORKSPACE/formatted/ip-$OUTPUT.txt",
                "banner": "Masscan 65535 ports",
                "cmd": "$ALIAS_PATH/portscan -i $TARGET -o '$WORKSPACE/portscan/$OUTPUT' -s '$WORKSPACE/portscan/summary.txt' -p '$PLUGINS_PATH'",
                "output_path": "$WORKSPACE/portscan/$OUTPUT.xml",
                "std_path": "",
                "waiting": "first",
            },
            {
                "requirement": "$WORKSPACE/portscan/$OUTPUT.csv",
                "banner": "CSV beautify",
                "cmd": "cat $WORKSPACE/portscan/$OUTPUT.csv | csvlook --no-inference | tee $WORKSPACE/portscan/beautify-$OUTPUT.txt",
                "output_path": "$WORKSPACE/portscan/beautify-$OUTPUT.txt",
                "std_path": "",
                "pre_run": "update_ports",
                "cleaned_output": "$WORKSPACE/portscan/formatted-$OUTPUT.txt",
                # "waiting": "last",
            },
            {
                "requirement": "$WORKSPACE/portscan/$OUTPUT.csv",
                "banner": "Screenshot on ports found",
                "cmd": "$GO_PATH/gowitness file -s $WORKSPACE/portscan/scheme-$OUTPUT.txt -t 30 --log-level fatal --destination $WORKSPACE/portscan/screenshot/raw-gowitness/ --db $WORKSPACE/portscan/screenshot/gowitness.db",
                "output_path": "$WORKSPACE/portscan/screenshot/gowitness.db",
                "std_path": "",
                "post_run": "clean_gowitness",
                "pre_run": "get_scheme",
                "cleaned_output": "$WORKSPACE/portscan/screenshot-$OUTPUT.html",
                "waiting": "last",
            },
            
        ],
    }


class VulnScan:
    reports = [
        {
            "path": "$WORKSPACE/vulnscan/final-$OUTPUT.html",
            "type": "html",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/vulnscan/beautify-summary-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                # "requirement": "$TARGET",
                "banner": "Nmap all port",
                "cmd": "$ALIAS_PATH/vulnscan -i $TARGET -o '$WORKSPACE/vulnscan/details/$OUTPUT' -s '$WORKSPACE/vulnscan/summary-$OUTPUT.csv' -p '$PLUGINS_PATH'",
                "output_path": "$WORKSPACE/vulnscan/details/$OUTPUT.gnmap",
                "std_path": "$WORKSPACE/vulnscan/details/std-$OUTPUT.std",
                "chunk": 3,
                "post_run": "gen_summary",
                "waiting": "first",
            },
            {
                "requirement": "$WORKSPACE/vulnscan/summary-$OUTPUT.csv",
                "banner": "CSV beautify",
                "cmd": "cat $WORKSPACE/vulnscan/summary-$OUTPUT.csv | csvlook --no-inference | tee $WORKSPACE/vulnscan/beautify-summary-$OUTPUT.txt",
                "output_path": "$WORKSPACE/vulnscan/beautify-summary-$OUTPUT.txt",
                "std_path": "",
            },
            {
                "requirement": "$WORKSPACE/vulnscan/summary-$OUTPUT.csv",
                "banner": "Screenshot on ports found",
                "cmd": "$GO_PATH/gowitness file -s $WORKSPACE/vulnscan/scheme-$OUTPUT.txt -t 30 --log-level fatal --destination  $WORKSPACE/vulnscan/screenshot/raw-gowitness/ --db $WORKSPACE/vulnscan/screenshot/gowitness.db",
                "output_path": "$WORKSPACE/vulnscan/screenshot/gowitness.db",
                "std_path": "",
                # "waiting": "last",
                "post_run": "clean_gowitness",
                "pre_run": "get_scheme",
            },
        ],
    }


class IPSpace:
    reports = [
        {
            "path": "$WORKSPACE/ipspace/summary-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/ipspace/range-$OUTPUT.txt",
            "type": "bash",
            "note": "final",
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "Metabigor IP Lookup",
                "cmd": "PLUGINS_PATH/Metabigor/metabigor.py -m ip -t $TARGET -o $WORKSPACE/ipspace/range-$OUTPUT.txt",
                "output_path": "$WORKSPACE/ipspace/range-$OUTPUT.txt",
                "std_path": "",
                "post_run": "get_amass",
                "cleaned_output": "$WORKSPACE/ipspace/summary-$OUTPUT.txt",
            },
        ],
    }


class GitScan:
    reports = [
        {
            "path": "$WORKSPACE/gitscan/$OUTPUT-user.txt",
            "type": "bash",
        },
        {
            "path": "$WORKSPACE/gitscan/$TARGET-gitrob.txt",
            "type": "bash",
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "parse input",
                "cmd": "$ALIAS_PATH/git_format -i '$RAW_TARGET' -o '$WORKSPACE/gitscan/$OUTPUT'",
                "output_path": "$WORKSPACE/gitscan/$OUTPUT-user.txt",
                "std_path": "",
                "waiting": "first",
            },
            {
                "requirement": "$WORKSPACE/gitscan/$OUTPUT-repo.txt",
                "banner": "Git recon repo",
                "cmd": "$ALIAS_PATH/gitrecon -r $WORKSPACE/gitscan/$OUTPUT-repo.txt -o '$WORKSPACE/gitscan/$OUTPUT' -k '$GITHUB_API_KEY' -p '$PLUGINS_PATH' ",
                "output_path": "$WORKSPACE/gitscan/$TARGET-gitrob.txt",
                "std_path": "$WORKSPACE/gitscan/std-$TARGET-gitrob.std"
            },
            {
                "requirement": "$WORKSPACE/gitscan/$OUTPUT-user.txt",
                "banner": "Git recon user",
                "cmd": "$ALIAS_PATH/gitrecon -u $WORKSPACE/gitscan/$OUTPUT-user.txt -o '$WORKSPACE/gitscan/$OUTPUT' -k '$GITHUB_API_KEY' -p '$PLUGINS_PATH' ",
                "output_path": "$WORKSPACE/gitscan/$TARGET-gitrob.txt",
                "std_path": "$WORKSPACE/gitscan/std-$TARGET-gitrob.std"
            }
        ],
    }
