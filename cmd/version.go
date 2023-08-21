package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	jsoniter "github.com/json-iterator/go"

	//"github.com/mackerelio/go-osstat/cpu"
	//"github.com/mackerelio/go-osstat/memory"
	"github.com/spf13/cobra"
	//"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func init() {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show core version",
		Long:  core.Banner(),
		RunE:  runVersion,
	}
	versionCmd.Flags().BoolVarP(&options.Verbose, "verbose", "V", false, "Show stat info too")
	versionCmd.Flags().BoolVar(&options.JsonOutput, "json", false, "Output as JSON")
	RootCmd.AddCommand(versionCmd)
}

func runVersion(_ *cobra.Command, _ []string) error {
	if options.JsonOutput {
		fmt.Println(PrintStat())
		return nil
	}

	if !options.Verbose {
		fmt.Printf("osmedeus %s by %s\n", libs.VERSION, libs.AUTHOR)
	} else {
		statInfo := PrintStat()
		fmt.Printf("osmedeus %s by %s -- %s\n", libs.VERSION, libs.AUTHOR, statInfo)
	}
	return nil
}

// StatData overview struct
type StatData struct {
	CPU     string `json:"cpu"`
	Mem     string `json:"mem"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// PrintStat print status
func PrintStat() string {
	data := GetStat()
	if data.CPU.Idle == 0.0 {
		return strings.TrimSpace(utils.Emojif(":thought_balloon:", "not responding"))
	}
	var cpu string
	cpuUsage := 100.0 - data.CPU.Idle
	if cpuUsage <= 20.0 {
		cpu = utils.Emojif(":green_circle:", "  cpu: %0.2f", cpuUsage)
	} else if (cpuUsage > 20.0) && (cpuUsage <= 50.0) {
		cpu = utils.Emojif(":green_circle:", "  cpu: %0.2f", cpuUsage)
	} else if (cpuUsage > 50.0) && (cpuUsage <= 80.0) {
		cpu = utils.Emojif(":orange_circle:", "  cpu: %0.2f", cpuUsage)
	} else {
		cpu = utils.Emojif(":red_circle:", "  cpu: %0.2f", cpuUsage)
	}

	var mem string
	memUsage := 100.0 - (data.Mem.Free/data.Mem.Total)*100
	if memUsage <= 20.0 {
		mem = utils.Emojif(":green_circle:", "  mem: %0.2f", memUsage)
	} else if (memUsage > 20.0) && (memUsage <= 50.0) {
		mem = utils.Emojif(":green_circle:", "  mem: %0.2f", memUsage)
	} else if (memUsage > 50.0) && (memUsage <= 80.0) {
		mem = utils.Emojif(":orange_circle:", "  mem: %0.2f", memUsage)
	} else {
		mem = utils.Emojif(":red_circle:", "  mem: %0.2f", memUsage)
	}

	name, _ := os.Hostname()
	if options.JsonOutput {
		stat := StatData{
			CPU:     fmt.Sprintf("%v", cpuUsage),
			Mem:     fmt.Sprintf("%v", memUsage),
			Name:    name,
			Version: fmt.Sprintf("osmedeus %s by %s", libs.VERSION, libs.AUTHOR),
		}
		if data, err := jsoniter.MarshalToString(stat); err == nil {
			return data
		}
	}

	return fmt.Sprintf("%s: %12s - %s", name, strings.TrimSpace(cpu), strings.TrimSpace(mem))
}

type ServerStatData struct {
	CPU struct {
		System float64
		User   float64
		Idle   float64
	}
	Mem struct {
		Total  float64
		Used   float64
		Free   float64
		Cached float64
	}
}

// GetStat get stat data
// func GetStat() ServerStatData {
// 	var stat ServerStatData

// 	before, err := cpu.Get()
// 	if err != nil {
// 		return stat
// 	}
// 	time.Sleep(time.Duration(1) * time.Second)
// 	after, err := cpu.Get()
// 	if err != nil {
// 		return stat
// 	}
// 	total := float64(after.Total - before.Total)
// 	stat.CPU.User = float64(after.User-before.User) / total * 100
// 	stat.CPU.System = float64(after.System-before.System) / total * 100
// 	stat.CPU.Idle = float64(after.Idle-before.Idle) / total * 100
// 	// memory part
// 	memory, err := memory.Get()
// 	if err != nil {
// 		return stat
// 	}
// 	stat.Mem.Total = float64(memory.Total+memory.SwapTotal) / (1024 * 1024 * 1024)
// 	stat.Mem.Used = float64(memory.Used+memory.SwapUsed) / (1024 * 1024 * 1024)
// 	stat.Mem.Used = float64(memory.Used+memory.SwapUsed) / (1024 * 1024 * 1024)
// 	stat.Mem.Cached = float64(memory.Cached) / (1024 * 1024 * 1024)
// 	stat.Mem.Free = float64(memory.Free+memory.SwapFree) / (1024 * 1024 * 1024)
// 	return stat
// }

// GetStat gets stat data
func GetStat() ServerStatData {
	var stat ServerStatData

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	time.Sleep(time.Duration(1) * time.Second)

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	totalAlloc := float64(after.TotalAlloc - before.TotalAlloc)
	sys := float64(after.Sys - before.Sys)
	stat.CPU.User = totalAlloc / sys * 100
	stat.CPU.System = sys / sys * 100
	stat.CPU.Idle = 100 - stat.CPU.User - stat.CPU.System

	// Memory part
	memory, err := mem.VirtualMemory()
	if err != nil {
		return stat
	}
	stat.Mem.Total = float64(memory.Total) / (1024 * 1024 * 1024)
	stat.Mem.Used = float64(memory.Used) / (1024 * 1024 * 1024)
	stat.Mem.Cached = float64(memory.Cached) / (1024 * 1024 * 1024)
	stat.Mem.Free = float64(memory.Free) / (1024 * 1024 * 1024)

	return stat
}
