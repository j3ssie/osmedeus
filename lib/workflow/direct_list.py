import multiprocessing

cpu_cores = multiprocessing.cpu_count()
threads = str(cpu_cores * 3)


class Formatting:
    reports = [
        {
            "path": "$WORKSPACE/formatted/$OUTPUT-domains.txt",
            "type": "bash"
        },
        {
            "path": "$WORKSPACE/formatted/$OUTPUT-range.txt",
            "type": "bash"
        }
    ]
    logs = []
    commands = {
        'general': [
            {
                "requirement": "$TARGET",
                "banner": "Formatting Input",
                "cmd": "$ALIAS_PATH/format_input -i $TARGET -o '$WORKSPACE/formatted/$OUTPUT'",
                "output_path": "$WORKSPACE/formatted/$OUTPUT-domains.txt",
                "std_path": "",
                "waiting": "first",
            },
            {
                "banner": "Resolve IP",
                "requirement": "$WORKSPACE/formatted/$OUTPUT-domains.txt",
                "cmd": "cat $WORKSPACE/formatted/$OUTPUT-domains.txt | $GO_PATH/just-resolved | tee -a $WORKSPACE/formatted/$OUTPUT-range.txt",
                "output_path": "$WORKSPACE/formatted/$OUTPUT-range.txt",
                "std_path": "",
            },
            {
                "banner": "Look for domain",
                "requirement": "$WORKSPACE/formatted/$OUTPUT-domains.txt",
                "cmd": "cat $WORKSPACE/formatted/$OUTPUT-domains.txt | filter-resolved -c 50 | tee $WORKSPACE/formatted/resolved-$OUTPUT.txt",
                "output_path": "$WORKSPACE/formatted/resolved-$OUTPUT.txt",
                "std_path": "",
            },
            {
                "banner": "Resolve HTTP",
                "requirement": "$WORKSPACE/formatted/$OUTPUT-domains.txt",
                "cmd": "cat $WORKSPACE/formatted/$OUTPUT-domains.txt | $GO_PATH/httprobe -c 100 | tee $WORKSPACE/formatted/http-$OUTPUT.txt",
                "output_path": "$WORKSPACE/formatted/http-$OUTPUT.txt",
                "std_path": "",
            },
        ],
        'slow': [
            {
                "banner": "Resolve HTTP full port",
                "requirement": "$WORKSPACE/formatted/$OUTPUT-domains.txt",
                "cmd": "cat $WORKSPACE/formatted/$OUTPUT-domains.txt | $GO_PATH/httprobe -c 100 -p xlarge | tee $WORKSPACE/formatted/http-$OUTPUT.txt",
                "output_path": "$WORKSPACE/formatted/http-$OUTPUT.txt",
                "std_path": "",
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
                "cmd": "$GO_PATH/webanalyze -apps $DATA_PATH/apps.json -hosts $TARGET -output json -worker 20 | tee $WORKSPACE/fingerprint/$OUTPUT-technology.json",
                "output_path": "$WORKSPACE/fingerprint/$OUTPUT-technology.json",
                "std_path": "$WORKSPACE/fingerprint/std-$OUTPUT-technology.std",
                "post_run": "update_tech",
                "cleaned_output": "$WORKSPACE/fingerprint/formatted-tech-$OUTPUT.txt",
            },
            {
                "requirement": "$WORKSPACE/formatted/$OUTPUT-scheme.txt",
                "banner": "meg /",
                "cmd": "$GO_PATH/meg / $WORKSPACE/formatted/$OUTPUT-scheme.txt $WORKSPACE/fingerprint/responses/ -v -c 100",
                "output_path": "$WORKSPACE/fingerprint/responses/index",
                "std_path": "",
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
                "cmd": "$GO_PATH/tko-subs -data $DATA_PATH/providers-data.csv -domains $WORKSPACE/formatted/resolved-$OUTPUT.txt -output $WORKSPACE/stoscan/takeover-$TARGET-tko-subs.txt",
                "output_path": "$WORKSPACE/stoscan/takeover-$TARGET-tko-subs.txt",
                "std_path": "$WORKSPACE/stoscan/std-takeover-$TARGET-tko-subs.std",
            },
            {
                "banner": "Subjack",
                "cmd": "$GO_PATH/subjack -v -m -c $DATA_PATH/fingerprints.json -w $WORKSPACE/formatted/resolved-$OUTPUT.txt -t 100 -timeout 30 -o $WORKSPACE/stoscan/takeover-$TARGET-subjack.txt -ssl",
                "output_path": "$WORKSPACE/stoscan/takeover-$TARGET-subjack.txt",
                "std_path": "$WORKSPACE/stoscan/std-takeover-$TARGET-subjack.std"
            },
            {
                "banner": "subzy",
                "cmd": "$GO_PATH/subzy -hide_fails -https -concurrency 20 -targets $WORKSPACE/formatted/resolved-$OUTPUT.txt | tee $WORKSPACE/stoscan/takeover-$TARGET-subzy.txt",
                "output_path": "$WORKSPACE/stoscan/takeover-$TARGET-subzy.txt",
                "std_path": "$WORKSPACE/stoscan/std-takeover-$TARGET-subzy.std"
            },
            {
                "banner": "massdns resolve IP",
                "cmd": "cat $WORKSPACE/formatted/resolved-$OUTPUT.txt | $PLUGINS_PATH/massdns/bin/massdns -r $DATA_PATH/resolvers.txt -q -t A -o F -w $WORKSPACE/stoscan/all-dig-info.txt",
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
                "requirement": "$WORKSPACE/formatted/http-$OUTPUT.txt",
                "banner": "CORS Scan",
                "cmd": "python2 $PLUGINS_PATH/CORStest/corstest.py -p 50 $WORKSPACE/formatted/http-$OUTPUT.txt | tee $WORKSPACE/cors/$OUTPUT-corstest.txt",
                "output_path": "$WORKSPACE/cors/$TARGET-corstest.txt",
                "std_path": "$WORKSPACE/cors/std-$TARGET-corstest.std",
            }
        ],
    }


class ParamFinding:
    reports = [
        {
            "path": "$WORKSPACE/params/summary-$OUTPUT.txt",
            "type": "bash",
            "note": "final, diff, slack",
        },
    ]
    logs = []
    commands = {
        'general': [
            {
                "banner": "Arjun param finding",
                "cmd": "$ALIAS_PATH/paramfinder -i '[[0]]' -o '$WORKSPACE/params/raw' -s '$WORKSPACE/params/summary-$OUTPUT.txt' -p '$PLUGINS_PATH'",
                "output_path": "",
                "std_path": "",
                "chunk": 5,
                "cmd_type": "list",
                "resources": "l0|$WORKSPACE/formatted/$OUTPUT-paths.txt",
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
                "requirement": "$WORKSPACE/formatted/$OUTPUT-domains.txt",
                "banner": "waybackurls",
                "cmd": "cat $WORKSPACE/probing/resolved-$OUTPUT.txt | $GO_PATH/waybackurls | tee $WORKSPACE/links/raw-wayback-$OUTPUT.txt",
                "output_path": "$WORKSPACE/links/raw-wayback-$OUTPUT.txt",
                "std_path": "$WORKSPACE/links/std-wayback-$OUTPUT.std",
                "post_run": "clean_waybackurls",
                "cleaned_output": "$WORKSPACE/links/waybackurls-$OUTPUT.txt",
            },
            {
                "requirement": "$WORKSPACE/formatted/http-$OUTPUT.txt",
                "pre_run": "get_domains",
                "banner": "linkfinder",
                "cmd": "$ALIAS_PATH/linkfinding -i '[[0]]' -o '$WORKSPACE/links/raw/' -s '$WORKSPACE/links/summary-$OUTPUT.txt' -p '$PLUGINS_PATH'",
                "output_path": "$WORKSPACE/links/raw/[[0]]-$OUTPUT.txt",
                "std_path": "",
                "chunk": 5,
                "cmd_type": "list",
                "resources": "l0|$WORKSPACE/formatted/http-$OUTPUT.txt",
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
                "requirement": "$WORKSPACE/formatted/http-$OUTPUT.txt",
                "banner": "aquatone",
                "cmd": f"cat $WORKSPACE/formatted/http-$OUTPUT.txt | $GO_PATH/aquatone -screenshot-timeout 50000 -threads {threads} -out $WORKSPACE/screenshot/$OUTPUT-aquatone",
                "output_path": "$WORKSPACE/screenshot/$OUTPUT-aquatone/aquatone_report.html",
                "std_path": "$WORKSPACE/screenshot/std-$OUTPUT-aquatone.std"
            },
            {
                "banner": "gowitness",
                "cmd": f"$GO_PATH/gowitness file -s $WORKSPACE/formatted/http-$OUTPUT.txt -t {threads} --timeout 10  --log-level fatal --destination  $WORKSPACE/screenshot/raw-gowitness/ --db $WORKSPACE/screenshot/gowitness.db",
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
                "requirement": "$WORKSPACE/formatted/http-$OUTPUT.txt",
                "banner": "csv beautify",
                "cmd": "cat $WORKSPACE/directory/raw/* | csvcut -c 2-6 | csvlook | tee $WORKSPACE/directory/beautify-summary.csv",
                "output_path": "",
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
            "path": "$WORKSPACE/portscan/beautify-$OUTPUT.txt",
            "type": "bash",
            "note": "final, slack, diff",
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
                "requirement": "$WORKSPACE/formatted/$OUTPUT-range.txt",
                "banner": "Masscan 65535 ports",
                "cmd": "$ALIAS_PATH/portscan -i $WORKSPACE/formatted/$OUTPUT-range.txt -o '$WORKSPACE/portscan/$OUTPUT' -s '$WORKSPACE/portscan/summary.txt' -p '$PLUGINS_PATH'",
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
                "requirement": "$WORKSPACE/formatted/$OUTPUT-domains.txt",
                "banner": "Nmap all port",
                "cmd": "$ALIAS_PATH/vulnscan -i [[0]] -o '$WORKSPACE/vulnscan/details/[[0]]' -s '$WORKSPACE/vulnscan/summary-$OUTPUT.csv' -p '$PLUGINS_PATH'",
                "output_path": "$WORKSPACE/vulnscan/details/[[0]].txt",
                "std_path": "$WORKSPACE/vulnscan/details/std-[[0]].std",
                "chunk": 3,
                "delay": 500,  # decrease this in high traffic network
                "cmd_type": "list",
                "resources": "l0|$WORKSPACE/formatted/$OUTPUT-domains.txt",
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


class GitScan:
    reports = [
        {
            "path": "$WORKSPACE/gitscan/$OUTPUT-repo.txt",
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
                "cmd": "$ALIAS_PATH/format_input -i '$TARGET' -o '$WORKSPACE/gitscan/$OUTPUT'",
                "output_path": "$WORKSPACE/gitscan/$OUTPUT-repo.txt",
                "std_path": "",
                "waiting": "first",
            },
            {
                "requirement": "$WORKSPACE/gitscan/$OUTPUT-repo.txt",
                "banner": "Git recon repo",
                "cmd": "$ALIAS_PATH/gitrecon -r $WORKSPACE/gitscan/$OUTPUT-repo.txt -o '$WORKSPACE/gitscan/$OUTPUT' -k '$GITHUB_API_KEY' -p '$PLUGIN_PATH' ",
                "output_path": "$WORKSPACE/gitscan/$TARGET-gitrob.txt",
                "std_path": "$WORKSPACE/gitscan/std-$TARGET-gitrob.std"
            },
            {
                "requirement": "$WORKSPACE/gitscan/$OUTPUT-user.txt",
                "banner": "Git recon user",
                "cmd": "$ALIAS_PATH/gitrecon -u $WORKSPACE/gitscan/$OUTPUT-repo.txt -o '$WORKSPACE/gitscan/$OUTPUT' -k '$GITHUB_API_KEY' -p '$PLUGIN_PATH' ",
                "output_path": "$WORKSPACE/gitscan/$TARGET-gitrob.txt",
                "std_path": "$WORKSPACE/gitscan/std-$TARGET-gitrob.std"
            }
        ],
    }
