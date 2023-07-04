package core

import (
	"fmt"
	"path"
	"time"

	"github.com/j3ssie/osmedeus/execution"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/robertkrimen/otto"
)

func (r *Runner) LoadExternalScripts() string {
	var output string
	vm := r.VM

	// special scripts
	vm.Set(Cleaning, func(call otto.FunctionCall) otto.Value {
		if r.Opt.NoClean {
			utils.InforF("Disabled Cleaning")
			return otto.Value{}
		}
		execution.Cleaning(call.Argument(0).String(), r.Reports)
		return otto.Value{}
	})

	// scripts for cleaning modules
	vm.Set(CleanAmass, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		execution.CleanAmass(src, dest)
		return otto.Value{}
	})

	vm.Set(CleanRustScan, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		execution.CleanRustScan(src, dest)
		return otto.Value{}
	})

	vm.Set(CleanGoBuster, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		execution.CleanGoBuster(src, dest)
		return otto.Value{}
	})
	vm.Set(CleanMassdns, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		execution.CleanMassdns(src, dest)
		return otto.Value{}
	})

	vm.Set(CleanSWebanalyze, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		execution.CleanSWebanalyze(src, dest)
		return otto.Value{}
	})
	vm.Set(CleanJSONDnsx, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		execution.CleanJSONDnsx(src, dest)
		return otto.Value{}
	})

	vm.Set(CleanJSONHttpx, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		execution.CleanJSONHttpx(src, dest)
		return otto.Value{}
	})

	// Deprecated
	vm.Set(CleanWebanalyze, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		args := call.ArgumentList

		techSum := path.Join(path.Dir(dest), fmt.Sprintf("tech-overview-%v.txt", r.Target["Workspace"]))
		if len(args) > 3 {
			techSum = args[2].String()
		}
		execution.CleanWebanalyze(src, dest, techSum)
		return otto.Value{}
	})

	vm.Set(CleanArjun, func(call otto.FunctionCall) otto.Value {
		// src mean folder contain arjun output
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		execution.CleanArjun(src, dest)
		return otto.Value{}
	})

	vm.Set(CleanFFUFJson, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		execution.CleanFFUFJson(src, dest)
		return otto.Value{}
	})

	vm.Set(GenNucleiReport, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		args := call.ArgumentList

		templateFile := ""
		if len(args) >= 3 {
			templateFile = args[2].String()
		}
		execution.GenNucleiReport(r.Opt, src, dest, templateFile)
		return otto.Value{}
	})

	return output
}

func (r *Runner) LoadGitScripts() string {
	var output string
	vm := r.VM
	options := r.Opt

	// Clone("git@xxx.git", "/tmp/dest")
	vm.Set(Clone, func(call otto.FunctionCall) otto.Value {
		execution.GitClone(call.Argument(0).String(), call.Argument(1).String(), false, options)
		return otto.Value{}
	})
	// like clone but delete the destination folder first
	vm.Set(FClone, func(call otto.FunctionCall) otto.Value {
		execution.GitClone(call.Argument(0).String(), call.Argument(1).String(), true, options)
		return otto.Value{}
	})

	vm.Set(PushResult, func(call otto.FunctionCall) otto.Value {
		for folder := range options.Storages {
			execution.PullResult(folder, options)
			time.Sleep(3 * time.Second)
			execution.PullResult(folder, options)
			commitMess := fmt.Sprintf("%v|%v|%v", options.Module.Name, options.Scan.ROptions["Workspace"], utils.GetCurrentDay())
			execution.PushResult(folder, commitMess, options)
		}
		return otto.Value{}
	})
	// push result but specific folder
	vm.Set(PushFolder, func(call otto.FunctionCall) otto.Value {
		folder := call.Argument(0).String()
		execution.PullResult(folder, options)
		time.Sleep(3 * time.Second)
		execution.PullResult(folder, options)
		commitMess := fmt.Sprintf("%v|%v|%v", options.Module.Name, options.Scan.ROptions["Workspace"], utils.GetCurrentDay())
		execution.PushResult(folder, commitMess, options)
		return otto.Value{}
	})

	// push result but specific folder
	vm.Set(PullFolder, func(call otto.FunctionCall) otto.Value {
		folder := call.Argument(0).String()
		execution.PullResult(folder, options)
		time.Sleep(3 * time.Second)
		execution.PullResult(folder, options)
		return otto.Value{}
	})

	vm.Set(DiffCompare, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		output := call.Argument(2).String()
		execution.DiffCompare(src, dest, output, options)
		return otto.Value{}
	})

	vm.Set(GitDiff, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		src := args[0].String()
		output := call.Argument(1).String()
		history := "1"
		if len(args) < 2 {
			history = call.Argument(2).String()
		}
		execution.GitDiff(src, output, history, options)
		return otto.Value{}
	})
	vm.Set(LoopGitDiff, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		src := args[0].String()
		output := call.Argument(1).String()
		execution.LoopGitDiff(src, output, options)
		return otto.Value{}
	})

	/* --- CDN S3 Bucket --- */

	// UploadToS3("/tmp/src", "xxx/xxx") or UploadToS3("/tmp/src", "your-cdn.s3.ap-southeast-1.amazonaws.com")
	vm.Set(UploadToS3, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		bucket := ""
		src := args[0].String()
		if len(args) > 1 {
			bucket = args[1].String()
		}

		// if bucket is empty, use the default one
		if bucket == "" {
			bucket = options.Cdn.Bucket
		}

		execution.UploadToS3(options, src, bucket)
		return otto.Value{}
	})

	// DownloadFromS3("xxx", "/tmp/dest")
	vm.Set(DownloadFromS3, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		bucket := ""
		src := args[0].String()
		dest := args[1].String()
		if len(args) > 2 {
			bucket = args[2].String()
		}

		// if bucket is empty, use the default one
		if bucket == "" {
			bucket = options.Cdn.Bucket
		}

		execution.DownloadFromS3(options, src, dest, bucket)
		return otto.Value{}
	})

	// DownloadFile("https://xxx.com", "/tmp/dest")
	vm.Set(DownloadFile, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		src := args[0].String()
		output := args[1].String()
		execution.DownloadFile(options, src, output)
		return otto.Value{}
	})

	/* --- end CDN S3 Bucket --- */

	/* --- Gitlab API --- */

	// CreateRepo("repo-name")
	// CreateRepo("repo-name", "tags")
	vm.Set(CreateRepo, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		repoName := args[0].String()
		tags := ""
		if len(args) > 1 {
			tags = args[1].String()
		}
		execution.CreateGitlabRepo(repoName, tags, options)
		return otto.Value{}
	})

	vm.Set(DeleteRepo, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		repoName := args[0].String()
		execution.DeleteRepo(repoName, 0, options)
		return otto.Value{}
	})
	vm.Set(DeleteRepoByPid, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		pid, err := args[0].ToInteger()
		if err != nil {
			return otto.Value{}
		}
		execution.DeleteRepo("", int(pid), options)
		return otto.Value{}
	})
	vm.Set(ListProjects, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		if len(args) > 0 {
			uid, err := args[0].ToInteger()
			if err == nil {
				execution.ListProjects(int(uid), options)
			}
			return otto.Value{}
		}
		execution.ListProjects(0, options)
		return otto.Value{}
	})

	/* --- end Gitlab API --- */

	// GenMarkdownReport("markdown.md", "output.html")
	vm.Set(GenMarkdownReport, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		markdownFile := args[0].String()
		outputHTML := args[1].String()
		r.GenMarkdownReport(markdownFile, outputHTML)
		return otto.Value{}
	})

	return output
}
