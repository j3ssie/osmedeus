package core

import (
    "fmt"
    "strings"

    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
)

type Tmux struct {
    ApplyAll       bool
    SelectedWindow string
    Exclude        string
    Limit          int
    Windows        []string
}

func InitTmux(options libs.Options) (Tmux, error) {
    cmd := "tmux ls"
    var tmux Tmux
    tmux.ApplyAll = options.Tmux.ApplyAll
    tmux.SelectedWindow = options.Tmux.SelectedWindow
    tmux.Exclude = options.Tmux.Exclude
    tmux.Limit = options.Tmux.Limit

    raw := utils.RunCmdWithOutput(cmd)
    if strings.Contains(raw, "command not found") || !strings.Contains(raw, "\n") {
        return tmux, fmt.Errorf("tmux program not installed")
    }

    stds := strings.Split(raw, "\n")
    for _, line := range stds {
        if strings.TrimSpace(line) == "" || !strings.Contains(line, " ") {
            continue
        }

        data := strings.Split(line, " ")
        tmux.Windows = append(tmux.Windows, strings.TrimRight(data[0], ":"))
    }
    return tmux, nil
}

func (t *Tmux) ListTmux() {
    if len(t.Windows) == 0 {
        fmt.Println("No tmux available")
        return
    }
    fmt.Println(strings.Join(t.Windows, ", "))
}

func (t *Tmux) CatchSession() string {
    var result string
    if len(t.Windows) == 0 {
        fmt.Println("No tmux available")
        return result
    }
    for _, window := range t.Windows {
        utils.DebugF("Get info of: %s", window)
        if !t.ApplyAll {
            if window != t.SelectedWindow {
                continue
            }
        }
        if strings.HasPrefix(window, t.Exclude) {
            continue
        }
        utils.InforF("Get output of %s session", window)
        cmd := fmt.Sprintf(`tmux capture-pane -pt "%s"`, window)
        raw := utils.RunCmdWithOutput(cmd)

        data := strings.Split(raw, "\n")
        if t.Limit == 0 {
            fmt.Println(raw)
        } else if t.Limit < len(data) && t.Limit > 0 {
            fmt.Println(strings.Join(data[len(data)-t.Limit:len(data)-1], "\n"))
        } else {
            fmt.Println(raw)
        }

        result += raw
    }

    return result
}
