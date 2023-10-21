package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/spf13/cobra"
)

// RootUsage base help
func RootUsage() {
	var h string
	h += ScanUsage()
	h += CloudUsage()
	h += QueueUsage()
	h += ReportUsage()
	h += UtilsUsage()
	fmt.Println(h)
}

func ScanExmaples() string {
	h := color.HiCyanString("Example Scan Commands:")
	h += color.HiBlueString("\n  ## Start a simple scan with default 'general' flow\n")
	h += "  osmedeus scan -t sample.com\n"

	h += color.HiBlueString("\n  ## Start a general scan but exclude some of the module\n")
	h += "  osmedeus scan -t sample.com -x screenshot -x spider\n"

	h += color.HiBlueString("\n  ## Start a scan directly with a module with inputs as a list of http domains like this https://sub.example.com\n")
	h += "  osmedeus scan -m content-discovery -t http-file.txt\n"

	h += color.HiBlueString("\n  ## Initiate the scan using a speed option other than the default setting\n")
	h += "  osmedeus scan -f vuln --tactic gently -t sample.com\n"
	h += "  osmedeus scan --threads-hold=10 -t sample.com\n"
	h += "  osmedeus scan -B 5 -t sample.com\n"

	h += color.HiBlueString("\n  ## Start a simple scan with other flow\n")
	h += "  osmedeus scan -f vuln -t sample.com\n"
	h += "  osmedeus scan -f extensive -t sample.com -t another.com\n"
	h += "  osmedeus scan -f urls -t list-of-urls.txt\n"

	h += color.HiBlueString("\n  ## Scan list of targets\n")
	h += "  osmedeus scan -T list_of_targets.txt\n"
	h += "  osmedeus scan -f vuln -T list-of-targets.txt\n"

	h += color.HiBlueString("\n  ## Performing static vulnerability scan and secret scan on a git repo\n")
	h += "  osmedeus scan -m repo-scan -t https://github.com/j3ssie/sample-repo\n"
	h += "  osmedeus scan -m repo-scan -t /tmp/source-code-folder\n"
	h += "  osmedeus scan -m repo-scan -T list-of-repo.txt\n"

	h += color.HiBlueString("\n  ## Scan for CIDR with file contains CIDR with the format '1.2.3.4/24'\n")
	h += "  osmedeus scan -f cidr -t list-of-ciders.txt\n"
	h += "  osmedeus scan -f cidr -t '1.2.3.4/24' # this will auto convert the single input to the file and run\n"

	h += color.HiBlueString("\n  ## Directly run on vuln scan and directory scan on list of domains\n")
	h += "  osmedeus scan -f domains -t list-of-domains.txt\n"
	h += "  osmedeus scan -f vuln-and-dirb -t list-of-domains.txt\n"

	h += color.HiBlueString("\n  ## Use a custom wordlist\n")
	h += "  osmedeus scan -t sample.com -p 'wordlists={{Data}}/wordlists/content/big.txt'\n"

	h += color.HiBlueString("\n  ## Use a custom wordlist\n")
	h += "  cat list_of_targets.txt | osmedeus scan -c 2\n"

	h += color.HiBlueString("\n  ## Start a normal scan and backup entire workflow folder to the backup folder\n")
	h += "  osmedeus scan --backup -f domains -t list-of-subdomains.txt\n"

	h += color.HiBlueString("\n  ## Start the scan with chunk inputs to review the output way more much faster\n")
	h += "  osmedeus scan --chunk --chunk-parts 20 -f cidr -t list-of-100-cidr.txt\n"

	h += color.HiBlueString("\n  ## Update the vulnerability database to the latest before starting the scan\n")
	h += "  osmedeus scan --update-vuln -f urls -t list-of-100-cidr.txt\n"

	h += color.HiBlueString("\n  ## Continuously run the scan on a target right after it finished\n")
	h += "  osmedeus utils cron --for --cmd 'osmedeus scan -t example.com'\n"

	h += color.HiBlueString("\n  ## Backing up all workspaces\n")
	h += "  ls ~/workspaces-osmedeus | osmedeus report compress\n"

	h += "\n"
	return h
}

func ScanUsage() string {
	h := ScanExmaples()
	h += color.HiCyanString("\nScan Usage:\n")
	h += "  osmedeus scan -f [flowName] -t [target] \n"
	h += "  osmedeus scan -m [modulePath] -T [targetsFile] \n"
	h += "  osmedeus scan -f /path/to/flow.yaml -t [target] \n"
	h += "  osmedeus scan -m /path/to/module.yaml -t [target] --params 'port=9200'\n"
	h += "  osmedeus scan -m /path/to/module.yaml -t [target] -l /tmp/log.log\n"
	h += "  osmedeus scan --tactic aggressive -m module -t [target] \n"
	h += "  cat targets | osmedeus scan -f sample\n"

	h += color.HiCyanString("\nPractical Scan Usage:\n")
	h += "  osmedeus scan -T list_of_targets.txt -W custom_workspaces\n"
	h += "  osmedeus scan -t target.com -w workspace_name --debug\n"
	h += "  osmedeus scan -f general -t sample.com\n"
	h += "  osmedeus scan --tactic aggressive -f general -t sample.com\n"
	h += "  osmedeus scan -f extensive -t sample.com -t another.com\n"
	h += "  cat list_of_urls.txt | osmedeus scan -f urls\n"
	h += "  osmedeus scan --threads-hold=15 -f cidr -t 1.2.3.4/24\n"
	h += "  osmedeus scan -m ~/.osmedeus/core/workflow/test/dirbscan.yaml -t list_of_urls.txt\n"
	h += "  osmedeus scan --wfFolder ~/custom-workflow/ -f your-custom-workflow -t list_of_urls.txt\n"
	h += "  osmedeus scan --chunk --chunk-part 40 -c 2 -f cidr -t list-of-cidr.txt\n"
	return h
}

func UtilsUsage() string {
	h := color.HiCyanString("\nUtilities Usage:\n")
	h += color.HiBlueString("  ## Health Utility\n")
	h += "  osmedeus health \n"
	h += "  osmedeus health git\n"
	h += "  osmedeus health cloud\n"
	h += "  osmedeus version --json \n"
	h += "\n"

	h += color.HiBlueString("  ## Set the base threads hold\n")
	h += "  osmedeus config set --threads-hold=10\n"
	h += "\n"

	h += color.HiBlueString("  ## Update utilities\n")
	h += "  osmedeus update \n"
	h += "  osmedeus update --vuln\n"
	h += "  osmedeus update --force --clean \n"
	h += "  osmedeus update --force --update-url https://very-long-url/premium.sh\n"
	h += "\n"

	h += color.HiBlueString("  ## Workflow utilities\n")
	h += "  osmedeus workflow list \n"
	h += "  osmedeus workflow view -f general\n"
	h += "  osmedeus workflow view -v -f general\n"
	h += "\n"

	h += color.HiBlueString("  ## Tmux utilities\n")
	h += "  osmedeus utils tmux ls \n"
	h += "  osmedeus utils tmux logs -A -l 10 \n"
	h += "\n"

	h += color.HiBlueString("  ## Process utilities\n")
	h += "  osmedeus utils ps \n"
	h += "  osmedeus utils ps --proc 'jaeles' \n"
	h += "\n"

	h += color.HiBlueString("  ## List all the sub proccess running by osmedeus\n")
	h += "  osmedeus utils ps --osm \n"
	h += "\n"

	h += color.HiBlueString("  ## Kill all the sub proccess running by osmedeus\n")
	h += "  osmedeus utils ps --osm --kill \n"
	h += "\n"

	h += color.HiBlueString("  ## Cron utilities\n")
	h += "  osmedeus utils cron --cmd 'osmdeus scan -t example.com' --sch 60\n"
	h += "  osmedeus utils cron --for --cmd 'osmedeus scan -t example.com'\n"

	return h
}

func ConfigUsage() string {
	h := color.HiCyanString("\nConfig Usage:\n")
	h += "  osmedeus config [action] [OPTIONS] \n"
	h += "  osmedeus config init -p https://github.com/j3ssie/osmedeus-plugins\n"
	h += "  osmedeus config --user newusser --pass newpassword\n"
	h += "  osmedeus config clean \n"
	h += "  osmedeus config delete -t woskapce \n"
	h += "  osmedeus config delete -w workspace_name \n"
	h += "  osmedeus config set --threads-hold=10 \n"
	return h
}

func QueueUsage() string {
	h := color.HiCyanString("\nQueue Usage:\n")
	h += "  osmedeus queue -Q /tmp/queue-file.txt -c 2\n"
	h += "  osmedeus queue --add -t example.com -Q /tmp/queue-file.txt \n"
	return h
}

func CloudUsage() string {
	h := color.HiCyanString("\nProvider Usage:\n")
	h += "  osmedeus provider wizard \n"
	h += "  osmedeus provider validate \n"
	h += "  osmedeus provider build --token xxx --rebuild --ic\n"

	h += "  osmedeus provider health --debug \n"
	h += "  osmedeus provider health --for \n"
	h += "  osmedeus provider create --name 'sample' \n"
	h += "  osmedeus provider delete --id 34317111 --id 34317112 \n"
	h += "  osmedeus provider list \n"

	h += color.HiCyanString("\nCloud Usage:\n")
	h += "  osmedeus cloud -f [flowName] -t [target] \n"
	h += "  osmedeus cloud -f [flowName] -T [targetFile] --no-del\n"
	h += "  osmedeus cloud -m [modulePath] -t [target] \n"
	h += "  osmedeus cloud -c 5 -f [flowName] -T [targetsFile] \n"
	h += "  osmedeus cloud --token xxx -c 5 -f [flowName] -T [targetsFile] \n"
	h += "  osmedeus cloud --chunk -c 5 -f [flowName] -t [targetsFile] \n"

	return h
}

func ReportUsage() string {
	h := color.HiCyanString("\nReport Usage:\n")
	h += "  osmedeus report list\n"
	h += "  osmedeus report extract -t target.com.tar.gz\n"
	h += "  osmedeus report extract -t target.com.tar.gz --dest .\n"
	h += "  osmedeus report compress -t target.com\n"
	h += "  osmedeus report view --raw -t target.com\n"
	h += "  osmedeus report view --static -t target.com\n"
	h += "  osmedeus report view --static --ip 0 -t target.com\n"
	return h
}

func ServerUsage() string {
	h := color.HiCyanString("\nServer Usage:\n")
	h += "  osmedeus server --port 5000\n"
	h += "  osmedeus server --disable-ssl\n"
	h += "  osmedeus server -A --disable-ssl\n"
	return h
}

func QueueHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(core.Banner())
	fmt.Println(cmd.UsageString())
	h := QueueUsage()
	fmt.Println(h)
	printDocs(cmd)
}

// ScanHelp scan help message
func ScanHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(core.Banner())
	if options.FullHelp {
		fmt.Println(cmd.UsageString())
	}
	h := ScanUsage()
	fmt.Println(h)
	printDocs(cmd)
}

// CloudHelp scan help message
func CloudHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(core.Banner())
	if options.FullHelp {
		fmt.Println(cmd.UsageString())
	}
	h := CloudUsage()
	fmt.Println(h)
	printDocs(cmd)
}

// ServerHelp scan help message
func ServerHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(core.Banner())
	if options.FullHelp {
		fmt.Println(cmd.UsageString())
	}
	h := ServerUsage()
	fmt.Println(h)
	printDocs(cmd)
}

// ConfigHelp config help message
func ConfigHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(core.Banner())
	if options.FullHelp {
		fmt.Println(cmd.UsageString())
	}
	h := ConfigUsage()
	fmt.Println(h)
	printDocs(cmd)
}

// UtilsHelp utils help message
func UtilsHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(core.Banner())
	if options.FullHelp {
		fmt.Println(cmd.UsageString())
	}
	h := UtilsUsage()
	fmt.Println(h)
	printDocs(cmd)
}

// ReportHelp utils help message
func ReportHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(core.Banner())
	if options.FullHelp {
		fmt.Println(cmd.UsageString())
	}
	h := ReportUsage()
	fmt.Println(h)
	printDocs(cmd)
}

// RootHelp print help message
func RootHelp(cmd *cobra.Command, _ []string) {
	fmt.Println(core.Banner())
	if options.FullHelp {
		fmt.Println(cmd.UsageString())
	}
	RootUsage()
	printDocs(cmd)
}

func printDocs(cmd *cobra.Command) {
	if !options.FullHelp {
		if cmd.Use == libs.BINARY {
			fmt.Printf("💡 For full help message, please run: %s\n", color.GreenString("osmedeus --hh"))
		} else {
			fmt.Printf("💡 For full help message, please run: %s or %s\n", color.GreenString("osmedeus --hh"), color.GreenString("osmedeus "+cmd.Use+" --hh"))
		}
	}
	fmt.Printf("📖 Documentation can be found here: %s\n", color.GreenString(libs.DOCS))
}
