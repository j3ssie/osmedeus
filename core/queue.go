package core

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	jsoniter "github.com/json-iterator/go"
	"github.com/panjf2000/ants"
	"github.com/thoas/go-funk"
	"os"
	"strings"
	"sync"
)

func QueueWatcher(options libs.Options) {
	queueFile := options.Queue.QueueFile

	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(options.Concurrency, func(i interface{}) {
		RunTheScan(i.(string), options)
		wg.Done()
	}, ants.WithPreAlloc(true))
	defer p.Release()

	data := utils.ReadingFileUnique(queueFile)

	/* Start the watcher */

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	utils.InforF("Starting to watch the queue file: %v", color.HiMagentaString(queueFile))

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op.String() == "WRITE" {
					utils.DebugF(color.HiMagentaString("modified file: %v -- %v", event.Name, event.Op))
					target := GetNewLine(queueFile)
					wg.Add(1)
					_ = p.Invoke(strings.TrimSpace(target))
				}

				if event.Op.String() == "REMOVE" {
					utils.ErrorF("Queue file removed, exiting ...")
					os.Exit(-1)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				utils.ErrorF("error: %v", err)
			}
		}
	}()

	err = watcher.Add(queueFile)
	if err != nil {
		panic(err)
	}

	// Just to trigger the file events
	utils.WriteToFile(queueFile, strings.Join(append(data, "\n"), "\n"))

	// Block main goroutine forever.
	<-make(chan struct{})
}

func RunTheScan(target string, options libs.Options) error {
	if strings.TrimSpace(target) == "" {
		return fmt.Errorf("target is empty")
	}
	utils.InforF("Picking the target from the queue: %v", color.CyanString(target))

	var inputFormat libs.InputFormat
	if ok := jsoniter.UnmarshalFromString(target, &inputFormat); ok == nil {
		utils.DebugF("Parsing the input in JSON format: %v", color.CyanString(target))
		cmd := CommandBuilder(inputFormat)
		utils.InforF("Running the command: %v", color.CyanString(cmd))
		utils.RunOSCommand(cmd)
		return nil
	}

	runner, err := InitRunner(target, options)
	if err != nil {
		utils.ErrorF("Error init runner with: %s", target)
		return err
	}
	runner.Start()

	return nil
}

func GetNewLine(queueFile string) string {
	data := utils.ReadingLines(queueFile)
	if len(data) == 0 {
		return ""
	}

	target := data[0]
	data = funk.DropString(data, 1)
	utils.DebugF("Getting the target from the queue file: %v -- %v", queueFile, color.CyanString(target))
	utils.WriteToFile(queueFile, strings.Join(data, "\n"))
	return target
}

func CommandBuilder(inputFormat libs.InputFormat) (command string) {
	if inputFormat.Command == "" {
		inputFormat.Command = fmt.Sprintf("%v scan -t %v", libs.BINARY, inputFormat.Input)
		if inputFormat.InputAsFile {
			inputFormat.Command = fmt.Sprintf("%v scan -T %v", libs.BINARY, inputFormat.Input)
		}

		if inputFormat.Flow == "" {
			inputFormat.Command += " -f " + inputFormat.Flow
		}

		// append the modules
		if len(inputFormat.Modules) > 0 {
			for _, item := range inputFormat.Modules {
				inputFormat.Command += " -m " + item
			}
		}

		if len(inputFormat.Params) > 0 {
			for _, item := range inputFormat.Params {
				inputFormat.Command += " -p " + item
			}
		}

		inputFormat.Command += " " + inputFormat.Extra
	}

	// formatting the command if there is any input in it
	inputFormat.Command = strings.ReplaceAll(inputFormat.Command, "{{.input}}", inputFormat.Input)
	return command
}
