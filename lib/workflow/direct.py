
import multiprocessing

cpu_cores = multiprocessing.cpu_count()
threads = str(cpu_cores * 2)


class SubdomainScanning:
    reports = [
        {
            "path": "$WORKSPACE/subdomain/final-$OUTPUT.txt",
            "type": "bash",
            "note": "final, slack, diff",
        }
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "Amass",
                "cmd": "$GO_PATH/amass enum -timeout 10 -active -max-dns-queries 10000 -include-unresolvable -active -dir $WORKSPACE/subdomain/amass-$OUTPUT -d $TARGET -o $WORKSPACE/subdomain/$OUTPUT-amass.txt",
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
                "cmd": "$PLUGINS_PATH/findomain -u $WORKSPACE/subdomain/$OUTPUT-findomain.txt -t $TARGET ",
                "output_path": "$WORKSPACE/subdomain/$TARGET-findomain.txt",
                "std_path": "$WORKSPACE/subdomain/std-$OUTPUT-findomain.std",
            },
            {
                "banner": "gobuster",
                "cmd": "$GO_PATH/gobuster dns --wildcard -q -t 100 -w $DATA_PATH/wordlists/dns/shorts.txt -d $TARGET -o $WORKSPACE/subdomain/raw-$OUTPUT-gobuster.txt",
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
                "cmd": "$ALIAS_PATH/vhosts -i '[[0]]' -o '$WORKSPACE/vhosts/raw/' -s '$WORKSPACE/vhosts/raw-summary-$OUTPUT.txt' -p '$PLUGINS_PATH' -w $DATA_PATH/wordlists/dns/virtual-host-scanning.txt",
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
            "note": "final, slack, diff",
        },
        {
            "path": "$WORKSPACE/probing/raw-allmassdns.txt",
            "type": "bash",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/probing/resolved-$OUTPUT.txt",
            "type": "bash",
            "note": "final, diff",
        },
        {
            "path": "$WORKSPACE/probing/http-$OUTPUT.txt",
            "type": "bash",
            "note": "final, slack, diff",
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                # this only run if something wrong with custom resolvers
                "banner": "massdns resolve IP",
                "requirement": "$WORKSPACE/probing/raw-all-$OUTPUT.txt",
                "cmd": "cat $WORKSPACE/probing/raw-all-$OUTPUT.txt | $PLUGINS_PATH/massdns/bin/massdns -r $DATA_PATH/resolvers.txt -q -t A -o S -w $WORKSPACE/probing/raw-allmassdns.txt",
                "output_path": "$WORKSPACE/probing/raw-allmassdns.txt",
                "std_path": "",
                "waiting": "last",
                "pre_run": "get_subdomains",
                "post_run": "clean_massdns",
                "cleaned_output": "$WORKSPACE/probing/ip-$OUTPUT.txt",
            },
            {
                "banner": "httprobe",
                "requirement": "$WORKSPACE/probing/raw-all-$OUTPUT.txt",
                "cmd": "cat $WORKSPACE/probing/raw-all-$OUTPUT.txt | $GO_PATH/httprobe -c 100 -t 20000 | tee $WORKSPACE/probing/http-$OUTPUT.txt",
                "output_path": "$WORKSPACE/probing/http-$OUTPUT.txt",
                "std_path": "$WORKSPACE/probing/std-http-$OUTPUT.std",
                "waiting": "last",
                "post_run": "get_domain",
                "cleaned_output": "$WORKSPACE/probing/domains-$OUTPUT.txt",
            },
        ],
    }


class Formatting:
    reports = [
        {
            "path": "$WORKSPACE/formatted/$OUTPUT-domains.txt",
            "type": "bash",
            "note": "final, slack"
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


class CORScan:
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
                "cmd": f"$GO_PATH/webanalyze -apps $DATA_PATH/apps.json -hosts $WORKSPACE/probing/http-$OUTPUT.txt -output json -worker {threads} | tee $WORKSPACE/fingerprint/$OUTPUT-technology.json",
                "output_path": "$WORKSPACE/fingerprint/$OUTPUT-technology.json",
                "std_path": "$WORKSPACE/fingerprint/std-$OUTPUT-technology.std",
                "post_run": "update_tech",
                "cleaned_output": "$WORKSPACE/fingerprint/formatted-tech-$OUTPUT.txt",
            },
            {
                "banner": "meg /",
                "cmd": "$GO_PATH/meg / $WORKSPACE/probing/http-$OUTPUT.txt $WORKSPACE/fingerprint/responses/ -v -c 100",
                "output_path": "$WORKSPACE/fingerprint/responses/index",
                "std_path": "",
            },
            {
                "requirement": "$WORKSPACE/fingerprint/responses/index",
                "banner": "rgf extract all",
                "cmd": "$GO_PATH/rgf -dir $WORKSPACE/fingerprint/responses/ | tee $WORKSPACE/fingerprint/rgf-all-$OUTPUT.txt",
                "output_path": "$WORKSPACE/fingerprint/rgf-all-$OUTPUT.txt",
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
                "cmd": f"cat $WORKSPACE/probing/resolved-$OUTPUT.txt | $GO_PATH/aquatone -scan-timeout 1000 -threads {threads} -out $WORKSPACE/screenshot/$OUTPUT-aquatone",
                "output_path": "$WORKSPACE/screenshot/$OUTPUT-aquatone/aquatone_report.html",
                "std_path": "$WORKSPACE/screenshot/std-$OUTPUT-aquatone.std"
            },
            {
                "banner": "gowitness",
                "cmd": f"$GO_PATH/gowitness file -s $WORKSPACE/probing/http-$OUTPUT.txt -t {threads}  --log-level fatal --destination  $WORKSPACE/screenshot/raw-gowitness/ --db $WORKSPACE/screenshot/gowitness.db",
                "output_path": "$WORKSPACE/screenshot/gowitness.db",
                "std_path": "",
            },
            {
                "banner": "gowitness gen report",
                "cmd": "$GO_PATH/gowitness report generate -c 99999 -n $WORKSPACE/screenshot/$OUTPUT-raw-gowitness.html --destination $WORKSPACE/screenshot/raw-gowitness/ --db $WORKSPACE/screenshot/gowitness.db",
                "output_path": "$WORKSPACE/screenshot/$OUTPUT-raw-gowitness-0.html",
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
                "banner": "subzy",
                "cmd": "$GO_PATH/subzy -hide_fails -https -concurrency 20 -targets $WORKSPACE/probing/resolved-$OUTPUT.txt | tee $WORKSPACE/stoscan/takeover-$TARGET-subzy.txt",
                "output_path": "$WORKSPACE/stoscan/takeover-$TARGET-subzy.txt",
                "std_path": "$WORKSPACE/stoscan/std-takeover-$TARGET-subzy.std"
            },
            {
                "banner": "massdns resolve IP",
                "cmd": "cat $WORKSPACE/probing/resolved-$OUTPUT.txt | $PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -q -t A -o F -w $WORKSPACE/stoscan/all-dig-info.txt",
                "output_path": "$WORKSPACE/stoscan/all-dig-info.txt",
                "std_path": "",
                "waiting": "first",
            },
            {
                "requirement": "$WORKSPACE/stoscan/all-dig-info.txt",
                "banner": "rgf extract CNAME",
                "cmd": "$GO_PATH/rgf -file $WORKSPACE/stoscan/all-dig-info.txt cname | tee $WORKSPACE/stoscan/have-cname.txt",
                "output_path": "$WORKSPACE/stoscan/have-cname.txt",
                "std_path": "",
            },
            {
                "requirement": "$WORKSPACE/stoscan/all-dig-info.txt",
                "banner": "rgf extract Azure",
                "cmd": "$GO_PATH/rgf -file $WORKSPACE/stoscan/all-dig-info.txt azure | tee $WORKSPACE/stoscan/azure-sto.txt",
                "output_path": "$WORKSPACE/stoscan/azure-sto.txt",
                "std_path": "",
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
            "path": "$WORKSPACE/portscan/$OUTPUT.html",
            "type": "html",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/portscan/beautify-$OUTPUT.txt",
            "type": "bash",
            "note": "final, slack, diff",
        },
        {
            "path": "$WORKSPACE/portscan/$OUTPUT-aquatone/aquatone_report.html",
            "type": "html",
            "note": "final",
        },
        {
            "path": "$WORKSPACE/portscan/screenshot/$OUTPUT-raw-gowitness.html",
            "type": "html",
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
            },
            {
                "requirement": "$WORKSPACE/portscan/$OUTPUT.csv",
                "banner": "Httprode new port",
                "cmd": '''cat $WORKSPACE/portscan/$OUTPUT.csv | awk -F',' '{print $1":"$4}' | httprobe -c 30 | tee $WORKSPACE/portscan/http-$OUTPUT.txt''',
                "output_path": "$WORKSPACE/portscan/http-$OUTPUT.txt",
                "std_path": "",
            },
            {
                "requirement": "$WORKSPACE/portscan/http-$OUTPUT.txt",
                "banner": "aquatone",
                "cmd": f"cat $WORKSPACE/portscan/http-$OUTPUT.txt | $GO_PATH/aquatone -screenshot-timeout 50000 -threads {threads} -out $WORKSPACE/portscan/$OUTPUT-aquatone",
                "output_path": "$WORKSPACE/portscan/$OUTPUT-aquatone/aquatone_report.html",
                "std_path": "$WORKSPACE/portscan/std-$OUTPUT-aquatone.std",
                "waiting": "last",
            },

            # {
            #     "requirement": "$WORKSPACE/portscan/$OUTPUT.csv",
            #     "banner": "Screenshot on ports found",
            #     "cmd": "$GO_PATH/gowitness file -s $WORKSPACE/portscan/scheme-$OUTPUT.txt -t 30 --log-level fatal --destination $WORKSPACE/portscan/screenshot/raw-gowitness/ --db $WORKSPACE/portscan/screenshot/gowitness.db",
            #     "output_path": "$WORKSPACE/portscan/screenshot/gowitness.db",
            #     "std_path": "",
            #     "post_run": "clean_gowitness",
            #     "pre_run": "get_scheme",
            #     "cleaned_output": "$WORKSPACE/portscan/screenshot-$OUTPUT.html",
            #     "waiting": "last",
            # },
            
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


class DirbScan:
    reports = [
        {
            "path": "$WORKSPACE/directory/raw-summary.txt",
            "type": "bash",
            "note": "final, diff, slack",
        },
        {
            "path": "$WORKSPACE/directory/beautify-summary.csv",
            "type": "bash",
            "note": "final, diff, slack",
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                "requirement": "$WORKSPACE/formatted/$OUTPUT-paths.txt",
                "banner": "Format fuzz URL",
                "cmd": "cat $WORKSPACE/formatted/$OUTPUT-paths.txt | unfurl -u format %s://%d%p/FUZZ | grep -v 'http:///FUZZ' > $WORKSPACE/directory/fuzz-$OUTPUT.txt",
                "output_path": "$WORKSPACE/directory/fuzz-$OUTPUT.txt",
                "std_path": "",
                "waiting": "first",
            },
            {
                "banner": "ffuf dirscan",
                "cmd": "$ALIAS_PATH/dirscan -i [[0]] -w '$DATA_PATH/wordlists/content/quick.txt' -o '$WORKSPACE/directory/raw' -p '$GO_PATH' -s '$WORKSPACE/directory'",
                "output_path": "",
                "std_path": "",
                "chunk": 5,
                "cmd_type": "list",
                "resources": "l0|$WORKSPACE/directory/fuzz-$OUTPUT.txt",
            },
            {
                "banner": "csv beautify",
                "cmd": "cat $WORKSPACE/directory/raw/* | csvcut -c 2-6 | csvlook | tee -a $WORKSPACE/directory/beautify-summary.csv",
                "output_path": "",
                "std_path": "",
                "waiting": "last",
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
