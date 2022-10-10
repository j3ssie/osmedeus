package execution

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
	gops "github.com/mitchellh/go-ps"
	"github.com/panjf2000/ants"
	"github.com/shirou/gopsutil/process"
	"github.com/spf13/cast"
	"strings"
	"sync"
)

func ListAllOsmedeusProcess() (pids []int) {
	processes, err := gops.Processes()
	if err != nil {
		return pids
	}
	var allProcess []OSProcess

	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(20, func(i interface{}) {
		defer wg.Done()

		ps := i.(gops.Process)
		pid := ps.Pid()
		proc, _ := process.NewProcess(cast.ToInt32(pid))
		cmd, _ := proc.Cmdline()
		if !strings.Contains(cmd, libs.BINARY) {
			return
		}

		osProcess := OSProcess{
			PID:     pid,
			Command: cmd,
		}
		fmt.Printf("pid:%v %s %v\n", color.HiCyanString("%v", osProcess.PID), color.HiMagentaString("--"), osProcess.Command)
		allProcess = append(allProcess, osProcess)
	}, ants.WithPreAlloc(true))
	defer p.Release()

	for _, ps := range processes {
		wg.Add(1)
		_ = p.Invoke(ps)

	}
	wg.Wait()

	return pids
}

type OSProcess struct {
	PID     int    `json:"pid"`
	Command string `json:"command"`
}

func GetOsmProcess(processName string) []OSProcess {
	if processName == "" {
		processName = libs.BINARY
	}
	var results []OSProcess
	processes, err := gops.Processes()
	if err != nil {
		return results
	}

	for _, ps := range processes {
		pid := ps.Pid()
		binary := ps.Executable()

		if strings.ToLower(binary) != strings.ToLower(processName) {
			continue
		}

		proc, _ := process.NewProcess(cast.ToInt32(pid))
		cmd, _ := proc.Cmdline()

		if strings.Contains(cmd, fmt.Sprintf("%s utils ps", libs.BINARY)) {
			continue
		}

		osmProcess := OSProcess{
			PID:     pid,
			Command: cmd,
		}
		results = append(results, osmProcess)
	}

	return results
}
