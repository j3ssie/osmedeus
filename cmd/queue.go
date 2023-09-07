package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
)

func init() {
	var queueCmd = &cobra.Command{
		Use:     "queue",
		Short:   "Running the scan with input from queue file",
		Aliases: []string{"queq", "quee", "queu", "que"},
		Long:    core.Banner(),
		RunE:    runQueue,
	}
	queueCmd.PersistentFlags().StringVarP(&options.Queue.QueueFile, "queue-file", "Q", fmt.Sprintf("~/.%s/queue/queue-mimic.txt", libs.BINARY), "File contain list of target to simulate the queue")
	queueCmd.PersistentFlags().BoolVar(&options.Queue.Add, "add", false, "Add new input to the queue file")
	queueCmd.PersistentFlags().BoolVarP(&options.Queue.InputAsFile, "as-file", "F", false, "treat input as a file")
	queueCmd.PersistentFlags().StringVar(&options.Queue.RawCommand, "cmd", "", "Raw Command to run")
	queueCmd.SetHelpFunc(QueueHelp)
	RootCmd.AddCommand(queueCmd)
	queueCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if options.FullHelp {
			cmd.Help()
			os.Exit(0)
		}
	}
}

func runQueue(_ *cobra.Command, _ []string) error {
	options.Queue.QueueFile = utils.NormalizePath(options.Queue.QueueFile)
	options.Queue.QueueFolder = path.Dir(options.Queue.QueueFile)
	if !utils.FolderExists(options.Queue.QueueFolder) {
		utils.MakeDir(options.Queue.QueueFolder)
	}

	if options.Queue.Add {
		addInput()
		return nil
	}

	if !utils.FileExists(options.Queue.QueueFile) {
		utils.WriteToFile(options.Queue.QueueFile, "")
	}

	content := utils.ReadingFileUnique(options.Queue.QueueFile)
	if len(content) == 0 {
		utils.WarnF("Queue file is empty: %v", options.Queue.QueueFile)
		utils.WarnF("Consider to add a input to it:" + color.HiGreenString(" osmedeus queue --add -t example.com"))
	} else {
		utils.InforF("Queue file is not empty: %v", color.HiCyanString(options.Queue.QueueFile))
		utils.InforF("Consider to delete it if you want a fresh scan")
	}

	core.QueueWatcher(options)
	return nil
}

func addInput() {
	utils.InforF("Adding new input to the queue file: %v", color.HiCyanString(options.Queue.QueueFile))
	// osmedeus queue --add -t example.com
	if options.Queue.RawCommand == "" {
		utils.WriteToFile(options.Queue.QueueFile, strings.Join(options.Scan.Inputs, "\n"))
		return
	}

	// osmedeus queue --add -t /tmp/cidr --cmd "osmedeus -t {{Input}} -m recon -w"
	for _, target := range options.Scan.Inputs {
		queueInput := libs.InputFormat{
			Input:       target,
			Command:     options.Queue.RawCommand,
			InputAsFile: options.Queue.InputAsFile,
		}
		if line, ok := jsoniter.MarshalToString(queueInput); ok == nil {
			utils.AppendToContent(options.Queue.QueueFile, line)
		}
	}
}
