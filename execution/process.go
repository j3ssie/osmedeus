package execution

import (
    "fmt"
    "github.com/j3ssie/osmedeus/libs"
    gops "github.com/mitchellh/go-ps"
    "github.com/shirou/gopsutil/process"
    "github.com/spf13/cast"
    "strings"
)

func ListProcess() {
    processes, err := gops.Processes()
    if err != nil {
        return
    }

    for _, ps := range processes {
        pid := ps.Pid()
        binary := ps.Executable()
        // should be change to osm process
        if binary != "HotKey" {
            continue
        }

        proc, _ := process.NewProcess(cast.ToInt32(pid))
        //spew.Dump(proc)
        cmd, _ := proc.Cmdline()

        fmt.Printf("pid -- %v : %v \n", pid, cmd)
        //return
    }
}

type OsmProcess struct {
    PID     int    `json:"pid"`
    Command string `json:"command"`
}

func GetOsmProcess(processName string) []OsmProcess {
    //var out string
    if processName == "" {
        processName = libs.BINARY
    }
    var results []OsmProcess
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

        osmProcess := OsmProcess{
            PID:     pid,
            Command: cmd,
        }
        results = append(results, osmProcess)
    }

    return results
}
