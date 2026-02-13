package functions

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
)

// GojaRuntime wraps the Goja JavaScript interpreter with VM pooling.
// Uses a pool of Goja VMs for parallel execution without global mutex.
type GojaRuntime struct {
	pool *VMPool    // Pool of configured VMs
	mu   sync.Mutex // Only used for custom function registration
}

// vmFunc provides VM access to all function implementations.
// This wrapper is needed because goja.FunctionCall doesn't provide direct VM access
// like otto.FunctionCall.Otto did.
type vmFunc struct {
	vm      *goja.Runtime
	runtime *GojaRuntime
}

// getContext retrieves the VMContext for the current VM
func (vf *vmFunc) getContext() *VMContext {
	return getVMContext(vf.vm)
}

// NewGojaRuntime creates a new Goja runtime with VM pooling
func NewGojaRuntime() *GojaRuntime {
	r := &GojaRuntime{}
	// Create pool with function registration callback
	r.pool = NewVMPool(r.registerFunctionsOnVM)
	return r
}

// Backward compatibility alias
var NewOttoRuntime = NewGojaRuntime

// OttoRuntime is an alias for backward compatibility
type OttoRuntime = GojaRuntime

// registerFunctionsOnVM registers all built-in functions on a given VM.
// This is called by the pool when creating new VMs.
func (r *GojaRuntime) registerFunctionsOnVM(vm *goja.Runtime) {
	vf := &vmFunc{vm: vm, runtime: r}

	// File functions
	_ = vm.Set(FnFileExists, vf.fileExists)
	_ = vm.Set(FnFileLength, vf.fileLength)
	_ = vm.Set(FnDirLength, vf.dirLength)
	_ = vm.Set(FnFileContains, vf.fileContains)
	_ = vm.Set(FnRegexExtract, vf.regexExtract)
	_ = vm.Set(FnReadFile, vf.readFile)
	_ = vm.Set(FnReadLines, vf.readLines)
	_ = vm.Set(FnRemoveFile, vf.removeFile)
	_ = vm.Set(FnRemoveFolder, vf.removeFolder)
	_ = vm.Set(FnRmRF, vf.rmRF)
	_ = vm.Set(FnRemoveAllExcept, vf.removeAllExcept)
	_ = vm.Set(FnCreateFolder, vf.createFolder)
	_ = vm.Set(FnAppendFile, vf.appendFile)
	_ = vm.Set(FnMoveFile, vf.moveFile)
	_ = vm.Set(FnGlob, vf.glob)
	_ = vm.Set(FnGrepStringToFile, vf.grepStringToFile)
	_ = vm.Set(FnGrepRegexToFile, vf.grepRegexToFile)
	_ = vm.Set(FnGrepString, vf.grepString)
	_ = vm.Set(FnGrepRegex, vf.grepRegex)
	_ = vm.Set(FnRemoveBlankLines, vf.removeBlankLines)
	_ = vm.Set(FnChunkFile, vf.chunkFile)

	// String functions
	_ = vm.Set(FnTrim, vf.trim)
	_ = vm.Set(FnTrimString, vf.trimString)
	_ = vm.Set(FnTrimLeft, vf.trimLeft)
	_ = vm.Set(FnTrimRight, vf.trimRight)
	_ = vm.Set(FnSplit, vf.split)
	_ = vm.Set(FnJoin, vf.join)
	_ = vm.Set(FnReplace, vf.replace)
	_ = vm.Set(FnContains, vf.contains)
	_ = vm.Set(FnStartsWith, vf.startsWith)
	_ = vm.Set(FnEndsWith, vf.endsWith)
	_ = vm.Set(FnToLowerCase, vf.toLowerCase)
	_ = vm.Set(FnToUpperCase, vf.toUpperCase)
	_ = vm.Set(FnMatch, vf.match)
	_ = vm.Set(FnRegexMatch, vf.regexMatch)
	_ = vm.Set(FnCutWithDelim, vf.cutWithDelim)
	_ = vm.Set(FnCut, vf.cutWithDelim) // alias for cut_with_delim
	_ = vm.Set(FnNormalizePath, vf.normalizePath)
	_ = vm.Set(FnGetTargetSpace, vf.getTargetSpace)
	_ = vm.Set(FnCleanSub, vf.cleanSub)

	// Type detection functions
	_ = vm.Set(FnGetTypes, vf.getTypes)
	_ = vm.Set(FnIsFile, vf.isFile)
	_ = vm.Set(FnIsDir, vf.isDir)
	_ = vm.Set(FnIsGit, vf.isGit)
	_ = vm.Set(FnIsURL, vf.isURLFunc)
	_ = vm.Set(FnIsCompress, vf.isCompress)
	_ = vm.Set(FnDetectLanguage, vf.detectLanguage)

	// Type conversion
	_ = vm.Set(FnParseInt, vf.parseInt)
	_ = vm.Set(FnParseFloat, vf.parseFloat)
	_ = vm.Set(FnToString, vf.toString)
	_ = vm.Set(FnToBoolean, vf.toBoolean)

	// Utility functions
	_ = vm.Set(FnLen, vf.length)
	_ = vm.Set(FnIsEmpty, vf.isEmpty)
	_ = vm.Set(FnIsNotEmpty, vf.isNotEmpty)
	_ = vm.Set(FnPrintf, vf.printf)
	_ = vm.Set(FnCatFile, vf.catFile)
	_ = vm.Set(FnExit, vf.exit)
	_ = vm.Set(FnSkip, vf.skip)
	_ = vm.Set(FnExecCmd, vf.execCmd)
	_ = vm.Set(FnBash, vf.bash)
	_ = vm.Set(FnSleep, vf.sleep)
	_ = vm.Set(FnCommandExists, vf.commandExists)
	_ = vm.Set(FnPickValid, vf.pickValid)
	_ = vm.Set(FnRunModule, vf.runModule)
	_ = vm.Set(FnRunFlow, vf.runFlow)
	_ = vm.Set(FnExecPython, vf.execPython)
	_ = vm.Set(FnExecPythonFile, vf.execPythonFile)

	// Logging functions
	_ = vm.Set(FnLogDebug, vf.logDebug)
	_ = vm.Set(FnLogInfo, vf.logInfo)
	_ = vm.Set(FnLogWarn, vf.logWarn)
	_ = vm.Set(FnLogError, vf.logError)

	// Color printing functions
	_ = vm.Set(FnPrintGreen, vf.printGreen)
	_ = vm.Set(FnPrintBlue, vf.printBlue)
	_ = vm.Set(FnPrintYellow, vf.printYellow)
	_ = vm.Set(FnPrintRed, vf.printRed)

	// Runtime variable functions
	_ = vm.Set(FnSetVar, vf.setVar)
	_ = vm.Set(FnGetVar, vf.getVar)

	// HTTP and network functions
	_ = vm.Set(FnHttpRequest, vf.httpRequest)
	_ = vm.Set(FnHttpGet, vf.httpGet)
	_ = vm.Set(FnHttpPost, vf.httpPost)
	_ = vm.Set(FnGetIP, vf.getIP)

	// LLM functions
	_ = vm.Set(FnLLMInvoke, vf.llmInvoke)
	_ = vm.Set(FnLLMInvokeCustom, vf.llmInvokeCustom)
	_ = vm.Set(FnLLMConversations, vf.llmConversations)

	// Generation functions
	_ = vm.Set(FnRandomString, vf.randomString)
	_ = vm.Set(FnUUID, vf.uuidFunc)

	// Encoding functions
	_ = vm.Set(FnBase64Encode, vf.base64Encode)
	_ = vm.Set(FnBase64Decode, vf.base64Decode)

	// Data query functions
	_ = vm.Set(FnJQ, vf.jq)
	_ = vm.Set(FnJQFromFile, vf.jqFromFile)

	// Notification functions
	_ = vm.Set(FnNotifyTelegram, vf.notifyTelegram)
	_ = vm.Set(FnSendTelegramFile, vf.sendTelegramFile)
	_ = vm.Set(FnNotifyTelegramChannel, vf.notifyTelegramChannel)
	_ = vm.Set(FnSendTelegramFileChannel, vf.sendTelegramFileChannel)
	_ = vm.Set(FnNotifyMessageAsFileTelegram, vf.notifyMessageAsFileTelegram)
	_ = vm.Set(FnNotifyMessageAsFileTelegramChannel, vf.notifyMessageAsFileTelegramChannel)
	_ = vm.Set(FnNotifyWebhook, vf.notifyWebhook)
	_ = vm.Set(FnSendWebhookEvent, vf.sendWebhookEvent)

	// Event generation functions
	_ = vm.Set(FnGenerateEvent, vf.generateEvent)
	_ = vm.Set(FnGenerateEventFromFile, vf.generateEventFromFile)

	// CDN/Storage functions
	_ = vm.Set(FnCdnUpload, vf.cdnUpload)
	_ = vm.Set(FnCdnDownload, vf.cdnDownload)
	_ = vm.Set(FnCdnExists, vf.cdnExists)
	_ = vm.Set(FnCdnDelete, vf.cdnDelete)
	_ = vm.Set(FnCdnSyncUpload, vf.cdnSyncUpload)
	_ = vm.Set(FnCdnSyncDownload, vf.cdnSyncDownload)
	_ = vm.Set(FnCdnGetPresignedURL, vf.cdnGetPresignedURL)
	_ = vm.Set(FnCdnList, vf.cdnList)
	_ = vm.Set(FnCdnStat, vf.cdnStat)
	_ = vm.Set(FnCdnRead, vf.cdnRead)
	_ = vm.Set(FnCdnLsTree, vf.cdnLsTree)

	// Unix command wrappers
	_ = vm.Set(FnSortUnix, vf.sortUnix)
	_ = vm.Set(FnWgetUnix, vf.wgetUnix)
	_ = vm.Set(FnWget, vf.wget)
	_ = vm.Set(FnGitClone, vf.gitClone)
	_ = vm.Set(FnGitCloneSubfolder, vf.gitCloneSubfolder)
	_ = vm.Set(FnZipUnix, vf.zipUnix)
	_ = vm.Set(FnUnzipUnix, vf.unzipUnix)
	_ = vm.Set(FnTarUnix, vf.tarUnix)
	_ = vm.Set(FnUntarUnix, vf.untarUnix)
	_ = vm.Set(FnDiffUnix, vf.diffUnix)
	_ = vm.Set(FnSedStringReplace, vf.sedStringReplace)
	_ = vm.Set(FnSedRegexReplace, vf.sedRegexReplace)

	// Archive functions (Go implementations)
	_ = vm.Set(FnZipDir, vf.zipDir)
	_ = vm.Set(FnUnzipDir, vf.unzipDir)
	_ = vm.Set(FnExtractTo, vf.extractTo)

	// Diff functions
	_ = vm.Set(FnExtractDiff, vf.extractDiff)

	// Output functions
	_ = vm.Set(FnSaveContent, vf.saveContent)
	_ = vm.Set(FnJSONLToCSV, vf.jsonlToCSV)
	_ = vm.Set(FnCSVToJSONL, vf.csvToJSONL)
	_ = vm.Set(FnJSONLUnique, vf.jsonlUnique)
	_ = vm.Set(FnJSONLFilter, vf.jsonlFilter)

	// URL processing functions
	_ = vm.Set(FnInterestingUrls, vf.interestingUrls)
	_ = vm.Set(FnGetParentURL, vf.getParentURL)
	_ = vm.Set(FnParseURL, vf.parseURL)
	_ = vm.Set(FnQueryReplace, vf.queryReplace)
	_ = vm.Set(FnPathReplace, vf.pathReplace)

	// Markdown functions
	_ = vm.Set(FnRenderMarkdownFromFile, vf.renderMarkdownFromFile)
	_ = vm.Set(FnPrintMarkdownFromFile, vf.printMarkdownFromFile)
	_ = vm.Set(FnConvertJSONLToMarkdown, vf.convertJSONLToMarkdown)
	_ = vm.Set(FnConvertCSVToMarkdown, vf.convertCSVToMarkdown)
	_ = vm.Set(FnRenderMarkdownReport, vf.renderMarkdownReport)
	_ = vm.Set(FnGenerateSecurityReport, vf.generateSecurityReport)

	// Database functions
	_ = vm.Set(FnDBUpdate, vf.dbUpdate)
	_ = vm.Set(FnDBImportAsset, vf.dbImportAsset)
	_ = vm.Set(FnDBQuickImportAsset, vf.dbQuickImportAsset)
	_ = vm.Set(FnDBRawInsertAsset, vf.dbRawInsertAsset)
	_ = vm.Set(FnDBPartialImportAsset, vf.dbPartialImportAsset)
	_ = vm.Set(FnDBPartialImportAssetFile, vf.dbPartialImportAssetFile)
	_ = vm.Set(FnDBTotalURLs, vf.dbTotalURLs)
	_ = vm.Set(FnDBTotalSubdomains, vf.dbTotalSubdomains)
	_ = vm.Set(FnDBTotalAssets, vf.dbTotalAssets)
	_ = vm.Set(FnDBTotalVulns, vf.dbTotalVulns)
	_ = vm.Set(FnDBVulnCritical, vf.dbVulnCritical)
	_ = vm.Set(FnDBVulnHigh, vf.dbVulnHigh)
	_ = vm.Set(FnDBVulnMedium, vf.dbVulnMedium)
	_ = vm.Set(FnDBVulnLow, vf.dbVulnLow)
	_ = vm.Set(FnDBTotalIPs, vf.dbTotalIPs)
	_ = vm.Set(FnDBTotalLinks, vf.dbTotalLinks)
	_ = vm.Set(FnDBTotalContent, vf.dbTotalContent)
	_ = vm.Set(FnDBTotalArchive, vf.dbTotalArchive)
	_ = vm.Set(FnRuntimeExport, vf.runtimeExport)
	_ = vm.Set(FnDBRegisterArtifact, vf.dbRegisterArtifact)
	_ = vm.Set(FnStoreArtifact, vf.storeArtifact)
	_ = vm.Set(FnDBSelectAssets, vf.dbSelectAssets)
	_ = vm.Set(FnDBSelectAssetsFiltered, vf.dbSelectAssetsFiltered)
	_ = vm.Set(FnDBSelectVulnerabilities, vf.dbSelectVulnerabilities)
	_ = vm.Set(FnDBSelectVulnerabilitiesFiltered, vf.dbSelectVulnerabilitiesFiltered)
	_ = vm.Set(FnDBSelect, vf.dbSelect)
	_ = vm.Set(FnDBSelectToFile, vf.dbSelectToFile)
	_ = vm.Set(FnDBSelectToJSONL, vf.dbSelectToJSONL)

	// Workspace stats SELECT functions (no arguments, use current workspace context)
	_ = vm.Set(FnDBSelectTotalSubdomains, vf.dbSelectTotalSubdomains)
	_ = vm.Set(FnDBSelectTotalURLs, vf.dbSelectTotalURLs)
	_ = vm.Set(FnDBSelectTotalAssets, vf.dbSelectTotalAssets)
	_ = vm.Set(FnDBSelectTotalVulns, vf.dbSelectTotalVulns)
	_ = vm.Set(FnDBSelectVulnCritical, vf.dbSelectVulnCritical)
	_ = vm.Set(FnDBSelectVulnHigh, vf.dbSelectVulnHigh)
	_ = vm.Set(FnDBSelectVulnMedium, vf.dbSelectVulnMedium)
	_ = vm.Set(FnDBSelectVulnLow, vf.dbSelectVulnLow)

	// JSONL import functions
	_ = vm.Set(FnDBImportAssetFromFile, vf.dbImportAssetFromFile)
	_ = vm.Set(FnDBImportVuln, vf.dbImportVuln)
	_ = vm.Set(FnDBImportVulnFromFile, vf.dbImportVulnFromFile)

	// DNS and custom asset import functions
	_ = vm.Set(FnDBImportDNSAsset, vf.dbImportDNSAsset)
	_ = vm.Set(FnDBImportCustomAsset, vf.dbImportCustomAsset)

	// SARIF import functions
	_ = vm.Set(FnDBImportSARIF, vf.dbImportSARIF)
	_ = vm.Set(FnConvertSARIFToMarkdown, vf.convertSARIFToMarkdown)

	// Database diff functions
	_ = vm.Set(FnDBAssetDiff, vf.dbAssetDiff)
	_ = vm.Set(FnDBVulnDiff, vf.dbVulnDiff)
	_ = vm.Set(FnDBAssetDiffToFile, vf.dbAssetDiffToFile)
	_ = vm.Set(FnDBVulnDiffToFile, vf.dbVulnDiffToFile)
	_ = vm.Set(FnDBSelectRuns, vf.dbSelectRuns)
	_ = vm.Set(FnDBSelectRunByUUID, vf.dbSelectRunByUUID)

	// Event log management functions
	_ = vm.Set(FnDBResetEventLogs, vf.dbResetEventLogs)

	// Installer functions
	_ = vm.Set(FnGoGetter, vf.goGetter)
	_ = vm.Set(FnGoGetterWithSSHKey, vf.goGetterWithSSHKey)
	_ = vm.Set(FnNixInstall, vf.nixInstall)
	_ = vm.Set(FnFilepathInstaller, vf.filepathInstaller)

	// Environment functions
	_ = vm.Set(FnOsGetenv, vf.osGetenv)
	_ = vm.Set(FnOsSetenv, vf.osSetenv)

	// SSH functions
	_ = vm.Set(FnSSHExec, vf.sshExec)
	_ = vm.Set(FnSSHRsync, vf.sshRsync)

	// Console for debugging
	_ = vm.Set("console", map[string]interface{}{
		"log": func(call goja.FunctionCall) goja.Value {
			fmt.Println(call.Argument(0).String())
			return goja.Undefined()
		},
	})
}

// Execute executes a JavaScript expression with context.
// Uses VM pooling for parallel execution without global mutex.
// Note: Uses full variable loading because functions like render_markdown_report()
// access variables via vm.Get() internally, not just from the expression text.
func (r *GojaRuntime) Execute(expr string, ctx map[string]interface{}) (interface{}, error) {
	// Get VM from pool (no global lock!)
	vmCtx := r.pool.Get()
	defer r.pool.Put(vmCtx)

	// Set context fields on this VM's context
	vmCtx.SetContext(ctx)

	// Set all context variables on the VM
	// Cannot use lazy loading here because functions may access variables
	// via vm.Get() internally (e.g., render_markdown_report reads Target, Output, etc.)
	if err := vmCtx.SetVariables(ctx); err != nil {
		return nil, fmt.Errorf("error setting variables: %w", err)
	}

	// Execute expression
	result, err := vmCtx.Run(expr)
	if err != nil {
		return nil, fmt.Errorf("error executing expression: %w", err)
	}

	// Export result to Go value (goja.Export() returns interface{} directly, no error)
	exported := result.Export()

	return exported, nil
}

// EvaluateCondition evaluates a boolean condition.
// Uses VM pooling for parallel execution without global mutex.
// Employs lazy variable loading - only sets variables actually referenced in the condition.
func (r *GojaRuntime) EvaluateCondition(condition string, ctx map[string]interface{}) (bool, error) {
	// Get VM from pool (no global lock!)
	vmCtx := r.pool.Get()
	defer r.pool.Put(vmCtx)

	// Use lazy loading - only set variables referenced in the condition
	// This is 50-80% faster for simple conditions with large contexts
	if err := vmCtx.SetVariablesLazy(ctx, condition); err != nil {
		return false, fmt.Errorf("error setting variables: %w", err)
	}

	// Execute condition
	result, err := vmCtx.Run(condition)
	if err != nil {
		return false, fmt.Errorf("error evaluating condition: %w", err)
	}

	// goja.ToBoolean() returns bool directly, no error
	boolResult := result.ToBoolean()

	return boolResult, nil
}

// Register registers a custom function on all VMs.
// Note: This only affects newly created VMs from the pool.
// For consistent behavior, register functions before first use.
func (r *GojaRuntime) Register(name string, fn interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Get a VM, register the function, then return it
	// Note: This is a limitation - custom functions only work on VMs that call this
	vmCtx := r.pool.Get()
	defer r.pool.Put(vmCtx)
	return vmCtx.VM().Set(name, fn)
}

// Clone returns the same runtime since VM pooling handles parallelism.
// The runtime is safe for concurrent use.
func (r *GojaRuntime) Clone() *GojaRuntime {
	return r
}

// GetPool returns the underlying VMPool.
// This allows external callers (e.g., scheduler) to use VMs with all utility functions registered.
func (r *GojaRuntime) GetPool() *VMPool {
	return r.pool
}
