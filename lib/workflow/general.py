'''
Storing all pre-defined commands
'''


class Sample:
    ''' Sample module '''
    reports = []  # point out reports file
    logs = []  # point out logs file
    # point out commands will run for each mode
    commands = {
        'general': [],  # run any time
        'quick': [],   # default mode
        'slow': [],  # will run if specific --slow '<module name>'
        'direct': [],  # will run in direct mode which is '-i' or '-I' options
    }


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
                "cmd": "cat $WORKSPACE/probing/raw-all-$OUTPUT.txt | $GO_PATH/httprobe -c 100 -t 20000 | tee $WORKSPACE/probing/http-$OUTPUT.txt",
                "output_path": "$WORKSPACE/probing/http-$OUTPUT.txt",
                "std_path": "$WORKSPACE/probing/std-http-$OUTPUT.std",
                # "waiting": "last",
                "post_run": "get_domain",
                "cleaned_output": "$WORKSPACE/probing/domains-$OUTPUT.txt",
            },
        ],
    }


class Formatting:
    reports = [
        {
            "path": "$WORKSPACE/formatted/ip-$OUTPUT.txt",
            "type": "bash",
            "note": "final"
        },
        {
            "path": "$WORKSPACE/formatted/$OUTPUT-scheme.txt",
            "type": "bash"
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                "requirement": "$WORKSPACE/probing/domains-$OUTPUT.txt",
                "banner": "Formatting Input",
                "cmd": "$ALIAS_PATH/format_input -i $WORKSPACE/probing/domains-$OUTPUT.txt -o '$WORKSPACE/formatted/$OUTPUT'",
                "output_path": "$WORKSPACE/formatted/$OUTPUT-domains.txt",
                "std_path": "",
                "waiting": "first",
            },
        ],
    }


class Fingerprint:
    reports = [
        {
            "path": "$WORKSPACE/fingerprint/$OUTPUT-technology.json",
            "type": "bash",
            "note": "final",
        }
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "webanalyze",
                "cmd": "$GO_PATH/webanalyze -apps $DATA_PATH/apps.json -hosts $WORKSPACE/probing/resolved-$OUTPUT.txt -output json -worker 20 | tee $WORKSPACE/fingerprint/$OUTPUT-technology.json",
                "output_path": "$WORKSPACE/fingerprint/$OUTPUT-technology.json",
                "std_path": "$WORKSPACE/fingerprint/std-$OUTPUT-technology.std",
                "post_run": "update_tech",
                "cleaned_output": "$WORKSPACE/fingerprint/formatted-tech-$OUTPUT.txt",
            },
            {
                "banner": "meg /",
                "cmd": "$GO_PATH/meg / $WORKSPACE/probing/resolved-$OUTPUT.txt $WORKSPACE/fingerprint/responses/ -v -c 100",
                "output_path": "$WORKSPACE/fingerprint/responses/index",
                "std_path": "",
            },
        ],
    }


class ScreenShot:
    reports = [
        {
            "path": "$WORKSPACE/screenshot/$OUTPUT-aquatone/aquatone_report.html",
            "type": "html",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/screenshot/$OUTPUT-gowitness.html",
            "type": "html",
            "note": "final",
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "aquatone",
                "cmd": "cat $WORKSPACE/probing/resolved-$OUTPUT.txt | $GO_PATH/aquatone -scan-timeout 1000 -threads 20 -out $WORKSPACE/screenshot/$OUTPUT-aquatone",
                "output_path": "$WORKSPACE/screenshot/$OUTPUT-aquatone/aquatone_report.html",
                "std_path": "$WORKSPACE/screenshot/std-$OUTPUT-aquatone.std"
            },
            {
                "banner": "gowitness",
                "cmd": "$GO_PATH/gowitness file -s $WORKSPACE/probing/http-$OUTPUT.txt -t 30 --log-level fatal --destination  $WORKSPACE/screenshot/raw-gowitness/ --db $WORKSPACE/screenshot/gowitness.db",
                "output_path": "$WORKSPACE/screenshot/gowitness.db",
                "std_path": "",
            },
            {
                "banner": "gowitness gen report",
                "cmd": "$GO_PATH/gowitness generate -n $WORKSPACE/screenshot/$OUTPUT-raw-gowitness.html --destination $WORKSPACE/screenshot/raw-gowitness/ --db $WORKSPACE/screenshot/gowitness.db",
                "output_path": "$WORKSPACE/screenshot/$OUTPUT-raw-gowitness.html",
                "std_path": "$WORKSPACE/screenshot/std-$OUTPUT-aquatone.std",
                "waiting": "last",
                "post_run": "clean_gowitness",
                "cleaned_output": "$WORKSPACE/screenshot/$OUTPUT-gowitness.html",
            },
        ],
    }


class StoScan:
    reports = [
        {
            "path": "$WORKSPACE/stoscan/takeover-$TARGET-tko-subs.txt",
            "type": "bash",
            "note": "final"
        },
        {
            "path": "$WORKSPACE/stoscan/takeover-$TARGET-subjack.txt",
            "type": "bash"
        },
        {
            "path": "$WORKSPACE/stoscan/all-dig-info.txt",
            "type": "bash"
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "tko-subs",
                "cmd": "$GO_PATH/tko-subs -data $DATA_PATH/providers-data.csv -domains $WORKSPACE/probing/resolved-$OUTPUT.txt -output $WORKSPACE/stoscan/takeover-$TARGET-tko-subs.txt",
                "output_path": "$WORKSPACE/stoscan/takeover-$TARGET-tko-subs.txt",
                "std_path": "$WORKSPACE/stoscan/std-takeover-$TARGET-tko-subs.std",
            },
            {
                "banner": "Subjack",
                "cmd": "$GO_PATH/subjack -v -m -c $DATA_PATH/fingerprints.json -w $WORKSPACE/probing/resolved-$OUTPUT.txt -t 100 -timeout 30 -o $WORKSPACE/stoscan/takeover-$TARGET-subjack.txt -ssl",
                "output_path": "$WORKSPACE/stoscan/takeover-$TARGET-subjack.txt",
                "std_path": "$WORKSPACE/stoscan/std-takeover-$TARGET-subjack.std"
            },
            {
                "banner": "massdns resolve IP",
                "cmd": "cat $WORKSPACE/probing/resolved-$OUTPUT.txt | $PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -q -t A -o F -w $WORKSPACE/stoscan/all-dig-info.txt",
                "output_path": "$WORKSPACE/stoscan/all-dig-info.txt",
                "std_path": "",
                "waiting": "last",
                # "post_run": "grepping_sign",
                # "cleaned_output": "$WORKSPACE/probing/really-final-$OUTPUT.txt",
            },
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
            "path": "$WORKSPACE/links/waybackurls-$OUTPUT.txt",
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
                "cmd": "python3 $PLUGINS_PATH/LinkFinder/linkfinder.py -i [[0]] -d -o cli | tee $WORKSPACE/links/raw/[[1]]-$OUTPUT.txt",
                "output_path": "tee $WORKSPACE/links/raw/[[1]]-$OUTPUT.txt",
                "std_path": "",
                "chunk": 5,
                "cmd_type": "list",
                "resources": "l0|$WORKSPACE/probing/http-$OUTPUT.txt;;l1|$WORKSPACE/probing/domain-$OUTPUT.txt",
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
                "requirement": "$WORKSPACE/probing/ip-$OUTPUT.txt",
                "banner": "Masscan 65535 ports",
                "cmd": "$ALIAS_PATH/portscan -i $WORKSPACE/probing/ip-$OUTPUT.txt -o '$WORKSPACE/portscan/$OUTPUT' -s '$WORKSPACE/portscan/summary.txt' -p '$PLUGINS_PATH'",
                "output_path": "$WORKSPACE/portscan/$OUTPUT.csv",
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
            },
            {
                "requirement": "$WORKSPACE/portscan/$OUTPUT.csv",
                "banner": "Screenshot on ports found",
                "cmd": "$GO_PATH/gowitness file -s $WORKSPACE/portscan/scheme-$OUTPUT.txt -t 30 --log-level fatal --destination $WORKSPACE/portscan/screenshot/raw-gowitness/ --db $WORKSPACE/portscan/screenshot/gowitness.db",
                "output_path": "$WORKSPACE/portscan/screenshot/gowitness.db",
                "std_path": "",
                "waiting": "last",
                "post_run": "clean_gowitness",
                "pre_run": "get_scheme",
                "cleaned_output": "$WORKSPACE/portscan/screenshot-$OUTPUT.html",
            }
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
        'slow': [
            {
                "requirement": "$WORKSPACE/probing/domain-$OUTPUT.txt",
                "banner": "Nmap all port",
                "cmd": "$ALIAS_PATH/vulnscan -i [[0]] -o '$WORKSPACE/vulnscan/details/[[0]]' -s '$WORKSPACE/vulnscan/summary.csv' -p '$PLUGINS_PATH'",
                "output_path": "$WORKSPACE/vulnscan/details/[[0]].gnmap",
                "std_path": "$WORKSPACE/vulnscan/details/std-[[0]].std",
                "chunk": 3,
                "delay": 500,  # decrease this in high traffic network
                "cmd_type": "list",
                "resources": "l0|$WORKSPACE/probing/domain-$OUTPUT.txt",
                "post_run": "gen_summary",
                "waiting": "first",
            },
            {
                "requirement": "$WORKSPACE/vulnscan/summary.csv",
                "banner": "Screenshot on ports found",
                "cmd": "$GO_PATH/gowitness file -s $WORKSPACE/vulnscan/scheme-$OUTPUT.txt -t 30 --log-level fatal --destination  $WORKSPACE/vulnscan/screenshot/raw-gowitness/ --db $WORKSPACE/vulnscan/screenshot/gowitness.db",
                "output_path": "$WORKSPACE/vulnscan/screenshot/gowitness.db",
                "std_path": "",
                # "waiting": "last",
                "post_run": "clean_gowitness",
                "pre_run": "get_scheme",
            },
            {
                "requirement": "$WORKSPACE/vulnscan/summary.csv",
                "banner": "CSV beautify",
                "cmd": "csvcut -c 1-7 $WORKSPACE/vulnscan/$OUTPUT-masscan.csv | csvlook --no-inference | tee $WORKSPACE/vulnscan/beautify-summary-$OUTPUT.txt",
                "output_path": "$WORKSPACE/vulnscan/beautify-summary-$OUTPUT.txt",
                "std_path": "",
                # "waiting": "last",
            }
        ],
    }


class DirbScan:
    reports = [
        {
            "path": "$WORKSPACE/directory/quick-$OUTPUT-dirble.txt",
            "type": "bash",
            "note": "final",
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                "requirement": "$WORKSPACE/probing/http-$OUTPUT.txt",
                "banner": "dirble",
                "cmd": "$PLUGINS_PATH/dirble/dirble -U $WORKSPACE/probing/http-$OUTPUT.txt -x '.php,.aspx,.jsp,.swp,.swf,.zip' -t 40 -k -w $DATA_PATH/wordlists/content/really-quick.txt --output-file $WORKSPACE/directory/quick-$OUTPUT-dirble.txt",
                "output_path": "$WORKSPACE/directory/quick-$OUTPUT-dirble.txt",
                "std_path": "$WORKSPACE/directory/std-$OUTPUT-dirble.std"
            },
        ],
    }


class GitScan:
    reports = [
        {
            "path": "$WORKSPACE/links/http-$OUTPUT.txt",
            "type": "bash"
        },
        {
            "path": "$WORKSPACE/assets/wayback-$OUTPUT.txt",
            "type": "bash"
        }
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "gitleaks",
                "cmd": "$GO_PATH/gitleaks -v --repo=$INPUT --report=$WORKSPACE/gitscan/$OUTPUT-gitleaks.json",
                "output_path": "$WORKSPACE/gitscan/$OUTPUT-gitleaks.json",
                "std_path": "$WORKSPACE/gitscan/std-$OUTPUT-gitleaks.std"
            },
            {
                "banner": "trufflehog",
                "cmd": "trufflehog --regex --entropy=True $INPUT | tee $WORKSPACE/gitscan/$OUTPUT-trufflehog.txt",
                "output_path": "$WORKSPACE/gitscan/$OUTPUT-trufflehog.txt",
                "std_path": "$WORKSPACE/gitscan/std-$OUTPUT-trufflehog.std"
            },
            {
                "banner": "gitrob",
                "cmd": "$GO_PATH/gitrob -save $WORKSPACE/gitscan/$TARGET-gitrob -threads 10 -github-access-token $GITHUB_API_KEY $INPUT",
                "output_path": "$WORKSPACE/gitscan/$TARGET-gitrob.txt",
                "std_path": "$WORKSPACE/gitscan/std-$TARGET-gitrob.std"
            }
        ],
    }

