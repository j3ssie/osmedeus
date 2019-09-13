
class Formatting:
    reports = [
        {
            "path": "$WORKSPACE/formatted/$OUTPUT-domains.txt",
            "type": "bash"
        },
        {
            "path": "$WORKSPACE/formatted/ip-$OUTPUT.txt",
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
                "banner": "massdns resolve IP",
                "requirement": "$TARGET",
                "cmd": "cat $WORKSPACE/formatted/$OUTPUT-domains.txt | $PLUGINS_PATH/massdns/bin/massdns -r $PLUGINS_PATH/massdns/lists/resolvers.txt -q -t A -o S -w $WORKSPACE/formatted/raw-massdns-$OUTPUT.txt",
                "output_path": "$WORKSPACE/formatted/raw-massdns-$OUTPUT.txt",
                "std_path": "",
                "post_run": "clean_massdns",
                "cleaned_output": "$WORKSPACE/formatted/ip-$OUTPUT.txt",
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
                "requirement": "$WORKSPACE/formatted/ip-$OUTPUT.txt",
                "banner": "CORS Scan",
                "cmd": "python2 $PLUGINS_PATH/CORStest/corstest.py -p 50 $WORKSPACE/formatted/http-$OUTPUT.txt| tee $WORKSPACE/cors/$OUTPUT-corstest.txt",
                "output_path": "$WORKSPACE/cors/$TARGET-corstest.txt",
                "std_path": "$WORKSPACE/cors/std-$TARGET-corstest.std",
            }
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
                "requirement": "$WORKSPACE/formatted/ip-$OUTPUT.txt",
                "banner": "Masscan 65535 ports",
                "cmd": "$ALIAS_PATH/portscan -i $WORKSPACE/formatted/ip-$OUTPUT.txt -o '$WORKSPACE/portscan/$OUTPUT' -s '$WORKSPACE/portscan/summary.txt' -p '$PLUGINS_PATH'",
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
