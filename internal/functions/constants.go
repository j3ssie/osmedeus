package functions

// Function name constants for easy reference and consistency
// This file serves as a central reference for all available workflow functions

// File Functions - Operations on files and directories
const (
	FnFileExists      = "file_exists"   // file_exists(path) -> bool
	FnFileLength      = "file_length"   // file_length(path) -> int (non-empty line count)
	FnDirLength       = "dir_length"    // dir_length(path) -> int (entry count)
	FnFileContains    = "file_contains" // file_contains(path, pattern) -> bool
	FnRegexExtract    = "regex_extract" // regex_extract(path, pattern) -> []string
	FnReadFile        = "read_file"     // read_file(path) -> string
	FnReadLines       = "read_lines"    // read_lines(path) -> []string
	FnRemoveFile      = "remove_file"   // remove_file(path) -> bool
	FnRemoveFolder    = "remove_folder" // remove_folder(path) -> bool
	FnRmRF            = "rm_rf"
	FnRemoveAllExcept = "remove_all_except"
	FnCreateFolder    = "create_folder" // create_folder(path) -> bool
	FnAppendFile      = "append_file"   // append_file(dest, source) -> bool
	FnMoveFile        = "move_file"     // move_file(source, dest) -> bool
	FnGlob            = "glob"          // glob(pattern) -> []string

	FnGrepStringToFile = "grep_string_to_file" // grep_string_to_file(dest, source, str) -> bool
	FnGrepRegexToFile  = "grep_regex_to_file"  // grep_regex_to_file(dest, source, pattern) -> bool
	FnGrepString       = "grep_string"         // grep_string(source, str) -> string
	FnGrepRegex        = "grep_regex"          // grep_regex(source, pattern) -> string
	FnRemoveBlankLines = "remove_blank_lines"  // remove_blank_lines(path) -> bool (in-place)
	FnChunkFile        = "chunk_file"          // chunk_file(input, lines_per_chunk, output) -> bool
)

// String Functions - String manipulation operations
const (
	FnTrim           = "trim"             // trim(str) -> string
	FnTrimString     = "trim_string"      // trim_string(input, substring) -> string (trim substring from both ends)
	FnTrimLeft       = "trim_left"        // trim_left(input, substring) -> string (trim substring from left/start)
	FnTrimRight      = "trim_right"       // trim_right(input, substring) -> string (trim substring from right/end)
	FnSplit          = "split"            // split(str, delim) -> []string
	FnJoin           = "join"             // join(arr, delim) -> string
	FnReplace        = "replace"          // replace(str, old, new) -> string
	FnContains       = "contains"         // contains(str, substr) -> bool
	FnStartsWith     = "starts_with"      // starts_with(str, prefix) -> bool
	FnEndsWith       = "ends_with"        // ends_with(str, suffix) -> bool
	FnToLowerCase    = "to_lower_case"    // to_lower_case(str) -> string
	FnToUpperCase    = "to_upper_case"    // to_upper_case(str) -> string
	FnMatch          = "match"            // match(str, pattern) -> bool
	FnRegexMatch     = "regex_match"      // regex_match(pattern, str) -> bool (pattern first)
	FnCutWithDelim   = "cut_with_delim"   // cut_with_delim(input, delim, field) -> string (1-indexed like cut)
	FnCut            = "cut"              // cut(input, delim, field) -> string (alias for cut_with_delim)
	FnNormalizePath  = "normalize_path"   // normalize_path(input) -> string (replace / | : etc with _)
	FnGetTargetSpace = "get_target_space" // get_target_space(input) -> string (same as {{TargetSpace}}: sanitize + truncate)
	FnCleanSub       = "clean_sub"        // clean_sub(path, target?) -> bool (clean and deduplicate subdomains in file)
)

// Type Detection Functions - Detect input types
const (
	FnGetTypes   = "get_types"   // get_types(input) -> string (file, folder, cidr, ip, url, domain, string)
	FnIsFile     = "is_file"     // is_file(path) -> bool
	FnIsDir      = "is_dir"      // is_dir(path) -> bool
	FnIsGit      = "is_git"      // is_git(path) -> bool
	FnIsURL      = "is_url"      // is_url(input) -> bool
	FnIsCompress      = "is_compress"      // is_compress(path) -> bool
	FnDetectLanguage  = "detect_language"  // detect_language(path) -> string (dominant programming language)
)

// Type Conversion Functions - Convert between types
const (
	FnParseInt   = "parse_int"   // parse_int(str) -> int
	FnParseFloat = "parse_float" // parse_float(str) -> float
	FnToString   = "to_string"   // to_string(val) -> string
	FnToBoolean  = "to_boolean"  // to_boolean(val) -> bool
)

// Utility Functions - General utility operations
const (
	FnLen            = "len"          // len(val) -> int
	FnIsEmpty        = "is_empty"     // is_empty(val) -> bool
	FnIsNotEmpty     = "is_not_empty" // is_not_empty(val) -> bool
	FnPrintf         = "printf"       // printf(message) -> void (print message to stdout)
	FnCatFile        = "cat_file"     // cat_file(path) -> void (print file content to stdout)
	FnExit           = "exit"         // exit(code) -> void (exit scan with code)
	FnExecCmd        = "exec_cmd"     // exec_cmd(command) -> string (alias for bash)
	FnBash           = "bash"
	FnSleep          = "sleep"            // sleep(seconds) -> void (pause for n seconds)
	FnCommandExists  = "command_exists"   // command_exists(command) -> bool (check if command exists in PATH)
	FnPickValid      = "pick_valid"       // pick_valid(v1, v2, ..., v10) -> any (first valid value)
	FnRunModule      = "run_module"       // run_module(module, target, params?) -> string (run osmedeus module)
	FnRunFlow        = "run_flow"         // run_flow(flow, target, params?) -> string (run osmedeus flow)
	FnExecPython     = "exec_python"      // exec_python(code) -> string (run inline Python, prefer python3)
	FnExecPythonFile = "exec_python_file" // exec_python_file(path) -> string (run Python file, prefer python3)
)

// Logging Functions - Log messages with level prefixes
const (
	FnLogDebug = "log_debug" // log_debug(message) -> void (print [DEBUG] message)
	FnLogInfo  = "log_info"  // log_info(message) -> void (print [INFO] message)
	FnLogWarn  = "log_warn"
	FnLogError = "log_error"
)

// Color Printing Functions - Print messages with colored output
const (
	FnPrintGreen  = "print_green"  // print_green(message) -> string (print in green)
	FnPrintBlue   = "print_blue"   // print_blue(message) -> string (print in blue)
	FnPrintYellow = "print_yellow" // print_yellow(message) -> string (print in yellow)
	FnPrintRed    = "print_red"    // print_red(message) -> string (print in red)
)

// Runtime Variable Functions - Set and get variables at runtime
const (
	FnSetVar = "set_var" // set_var(name, value) -> string (set runtime variable)
	FnGetVar = "get_var" // get_var(name) -> string (get runtime variable)
)

// HTTP and Network Functions
const (
	FnHttpRequest = "http_request" // http_request(url, method, headers, body) -> {statusCode, body, headers}
	FnHttpGet     = "http_get"     // http_get(url) -> structured JSON response
	FnHttpPost    = "http_post"    // http_post(url, body) -> structured JSON response
	FnGetIP       = "get_ip"       // get_ip(domain_or_url) -> string (resolved IP address)
)

// Generation Functions - Generate random values
const (
	FnRandomString = "random_string" // random_string(length) -> string
	FnUUID         = "uuid"          // uuid() -> string (UUID v4)
)

// Encoding Functions - Encode/decode data
const (
	FnBase64Encode = "base64_encode" // base64_encode(str) -> string
	FnBase64Decode = "base64_decode" // base64_decode(str) -> string
)

// Data Query Functions - Query structured data
const (
	FnJQ         = "jq" // jq(jsonData, query) -> any (extract data using jq syntax)
	FnJQFromFile = "jq_from_file"
)

// Notification Functions - Send notifications via various channels
const (
	FnNotifyTelegram                     = "notify_telegram"                         // notify_telegram(message) -> bool
	FnSendTelegramFile                   = "send_telegram_file"                      // send_telegram_file(path, caption?) -> bool
	FnNotifyTelegramChannel              = "notify_telegram_channel"                 // notify_telegram_channel(channel, message) -> bool
	FnSendTelegramFileChannel            = "send_telegram_file_channel"              // send_telegram_file_channel(channel, path, caption?) -> bool
	FnNotifyMessageAsFileTelegram        = "notify_message_as_file_telegram"         // notify_message_as_file_telegram(path) -> bool
	FnNotifyMessageAsFileTelegramChannel = "notify_message_as_file_telegram_channel" // notify_message_as_file_telegram_channel(channel, path) -> bool
	FnNotifyWebhook                      = "notify_webhook"                          // notify_webhook(message) -> bool
	FnSendWebhookEvent                   = "send_webhook_event"                      // send_webhook_event(eventType, data) -> bool
)

// Event Generation Functions - Generate structured events
const (
	FnGenerateEvent         = "generate_event"           // generate_event(workspace, topic, source, data_type, data) -> bool
	FnGenerateEventFromFile = "generate_event_from_file" // generate_event_from_file(workspace, topic, source, data_type, path) -> int
)

// CDN/Storage Functions - Cloud storage operations
const (
	FnCdnUpload          = "cdn_upload"            // cdn_upload(localPath, remotePath) -> bool
	FnCdnDownload        = "cdn_download"          // cdn_download(remotePath, localPath) -> bool
	FnCdnExists          = "cdn_exists"            // cdn_exists(remotePath) -> bool
	FnCdnDelete          = "cdn_delete"            // cdn_delete(remotePath) -> bool
	FnCdnSyncUpload      = "cdn_sync_upload"       // cdn_sync_upload(localDir, remotePrefix) -> object
	FnCdnSyncDownload    = "cdn_sync_download"     // cdn_sync_download(remotePrefix, localDir) -> object
	FnCdnGetPresignedURL = "cdn_get_presigned_url" // cdn_get_presigned_url(remotePath, expiryMins?) -> string
	FnCdnList            = "cdn_list"              // cdn_list(prefix?) -> []object
	FnCdnStat            = "cdn_stat"              // cdn_stat(remotePath) -> object|null
	FnCdnRead            = "cdn_read"              // cdn_read(remotePath) -> string
	FnCdnLsTree          = "cdn_ls_tree"           // cdn_ls_tree(prefix?) -> string (tree format)
)

// Unix Command Wrappers - Wrappers around common Unix commands
const (
	FnSortUnix          = "sort_unix"           // sort_unix(inputFile, outputFile?) -> bool (LC_ALL=C sort -u)
	FnWgetUnix          = "wget_unix"           // wget_unix(url, outputPath?) -> bool
	FnWget              = "wget"                // wget(url, outputPath) -> bool (pure Go, segmented download)
	FnGitClone          = "git_clone"           // git_clone(repo, dest?) -> bool
	FnGitCloneSubfolder = "git_clone_subfolder" // git_clone_subfolder(git_url, subfolder, dest) -> bool
	FnZipUnix           = "zip_unix"            // zip_unix(source, dest) -> bool (zip -r dest source)
	FnUnzipUnix         = "unzip_unix"          // unzip_unix(source, dest?) -> bool (unzip source -d dest)
	FnTarUnix           = "tar_unix"            // tar_unix(source, dest) -> bool (tar -czf dest source)
	FnUntarUnix         = "untar_unix"          // untar_unix(source, dest?) -> bool (tar -xzf source -C dest)
	FnDiffUnix          = "diff_unix"           // diff_unix(file1, file2, output?) -> string
	FnSedStringReplace  = "sed_string_replace"  // sed_string_replace(sed_syntax, source, dest) -> bool
	FnSedRegexReplace   = "sed_regex_replace"   // sed_regex_replace(sed_syntax, source, dest) -> bool
)

// Installer Functions - Download and install packages
const (
	FnGoGetter           = "go_getter"             // go_getter(url, dest) -> bool
	FnGoGetterWithSSHKey = "go_getter_with_sshkey" // go_getter_with_sshkey(ssh_key_path, git_url, dest) -> bool
	FnNixInstall         = "nix_install"           // nix_install(package, dest?) -> bool
	FnFilepathInstaller  = "filepath_installer"    // filepath_installer(local_path, tool_name, dest?) -> bool
)

// Environment Functions - Environment variable operations
const (
	FnOsGetenv = "os_getenv" // os_getenv(name) -> string
	FnOsSetenv = "os_setenv" // os_setenv(name, value) -> bool
)

// LLM Functions - Invoke LLM from workflows
const (
	FnLLMInvoke        = "llm_invoke"        // llm_invoke(message) -> string (simple direct message)
	FnLLMInvokeCustom  = "llm_invoke_custom" // llm_invoke_custom(message, body_json) -> string (custom POST body with {{message}} placeholder)
	FnLLMConversations = "llm_conversations" // llm_conversations(msg1, msg2, ...) -> string (multi-turn with "role:content" format)
)

// Archive Functions - Go implementations for zip/unzip
const (
	FnZipDir    = "zip_dir"    // zip_dir(source, dest) -> bool
	FnUnzipDir  = "unzip_dir"  // unzip_dir(source, dest) -> bool
	FnExtractTo = "extract_to" // extract_to(source, dest) -> bool (auto-detect .zip, .tar.gz, .tar.bz2, .tar.xz, .tgz; removes dest first)
)

// Diff Functions - Compare files
const (
	FnExtractDiff = "extract_diff" // extract_diff(file1, file2) -> string (lines only in file2)
)

// Output Functions - Save content to files
const (
	FnSaveContent = "save_content" // save_content(content, path) -> bool
	FnJSONLToCSV  = "jsonl_to_csv"
	FnCSVToJSONL  = "csv_to_jsonl"
	FnJSONLUnique = "jsonl_unique"
	FnJSONLFilter = "jsonl_filter"
)

// URL Processing Functions - URL deduplication, filtering, and parsing
const (
	FnInterestingUrls = "interesting_urls" // interesting_urls(src, dest, json_field?) -> bool
	FnGetParentURL    = "get_parent_url"   // get_parent_url(url) -> string (strips last path component)
	FnParseURL        = "parse_url"        // parse_url(url, format) -> string (format directives like unfurl)
	FnQueryReplace    = "query_replace"    // query_replace(url, value, mode?) -> string (replace all query param values)
	FnPathReplace     = "path_replace"     // path_replace(url, value, position?) -> string (replace path segment at position)
)

// Markdown Functions - Markdown rendering and conversion
const (
	FnRenderMarkdownFromFile = "render_markdown_from_file" // render_markdown_from_file(path) -> string (rendered markdown)
	FnPrintMarkdownFromFile  = "print_markdown_from_file"  // print_markdown_from_file(path) -> void (print with syntax highlight)
	FnConvertJSONLToMarkdown = "convert_jsonl_to_markdown" // convert_jsonl_to_markdown(input_path, output_path) -> bool (writes markdown table to file)
	FnConvertCSVToMarkdown   = "convert_csv_to_markdown"   // convert_csv_to_markdown(path) -> string (markdown table)
	FnRenderMarkdownReport   = "render_markdown_report"    // render_markdown_report(template_path, output_path) -> bool
	FnGenerateSecurityReport = "generate_security_report"  // generate_security_report(template_path) -> bool (output to {{Output}}/security-report.md)
)

// Database Functions - Database update and import operations
const (
	FnDBUpdate                        = "db_update"                          // db_update(table, key, field, value) -> bool
	FnDBImportAsset                   = "db_import_asset"                    // db_import_asset(workspace, json_data) -> bool
	FnDBQuickImportAsset              = "db_quick_import_asset"              // db_quick_import_asset(workspace, asset_value, asset_type?) -> bool
	FnDBRawInsertAsset                = "db_raw_insert_asset"                // db_raw_insert_asset(workspace, json_data) -> int (asset ID)
	FnDBPartialImportAsset            = "db_partial_import_asset"            // db_partial_import_asset(workspace, asset_type, asset_value) -> bool
	FnDBPartialImportAssetFile        = "db_partial_import_asset_file"       // db_partial_import_asset_file(workspace, asset_type, file_path) -> int (count)
	FnDBTotalURLs                     = "db_total_urls"                      // db_total_urls(file_path) -> int (count lines, update workspace)
	FnDBTotalSubdomains               = "db_total_subdomains"                // db_total_subdomains(file_path) -> int
	FnDBTotalAssets                   = "db_total_assets"                    // db_total_assets(file_path) -> int
	FnDBTotalVulns                    = "db_total_vulns"                     // db_total_vulns(file_path) -> int
	FnDBVulnCritical                  = "db_vuln_critical"                   // db_vuln_critical(file_path) -> int
	FnDBVulnHigh                      = "db_vuln_high"                       // db_vuln_high(file_path) -> int
	FnDBVulnMedium                    = "db_vuln_medium"                     // db_vuln_medium(file_path) -> int
	FnDBVulnLow                       = "db_vuln_low"                        // db_vuln_low(file_path) -> int
	FnDBTotalIPs                      = "db_total_ips"                       // db_total_ips(file_path) -> int
	FnDBTotalLinks                    = "db_total_links"                     // db_total_links(file_path) -> int
	FnDBTotalContent                  = "db_total_content"                   // db_total_content(file_path) -> int
	FnDBTotalArchive                  = "db_total_archive"                   // db_total_archive(file_path) -> int
	FnRuntimeExport                   = "runtime_export"                     // runtime_export() -> bool (export scan+workspace to run-state.json)
	FnDBRegisterArtifact              = "register_artifact"                  // register_artifact(path, type?) -> bool (register file as scan artifact)
	FnStoreArtifact                   = "store_artifact"                     // store_artifact(path, type?) -> bool (store file as scan artifact)
	FnDBSelectAssets                  = "db_select_assets"                   // db_select_assets(workspace, format) -> string
	FnDBSelectAssetsFiltered          = "db_select_assets_filtered"          // db_select_assets_filtered(workspace, status_code, asset_type, format) -> string
	FnDBSelectVulnerabilities         = "db_select_vulnerabilities"          // db_select_vulnerabilities(workspace, format) -> string
	FnDBSelectVulnerabilitiesFiltered = "db_select_vulnerabilities_filtered" // db_select_vulnerabilities_filtered(workspace, severity, asset_value, format) -> string
	FnDBSelect                        = "db_select"                          // db_select(sql_query, format) -> string
	FnDBSelectToFile                  = "db_select_to_file"                  // db_select_to_file(sql_query, dest) -> bool
	FnDBSelectToJSONL                 = "db_select_to_jsonl"                 // db_select_to_jsonl(sql_query, fields, dest) -> bool

	// SELECT functions - read workspace stats without arguments (uses current workspace context)
	FnDBSelectTotalSubdomains = "db_select_total_subdomains" // db_select_total_subdomains() -> int
	FnDBSelectTotalURLs       = "db_select_total_urls"       // db_select_total_urls() -> int
	FnDBSelectTotalAssets     = "db_select_total_assets"     // db_select_total_assets() -> int
	FnDBSelectTotalVulns      = "db_select_total_vulns"      // db_select_total_vulns() -> int
	FnDBSelectVulnCritical    = "db_select_vuln_critical"    // db_select_vuln_critical() -> int
	FnDBSelectVulnHigh        = "db_select_vuln_high"        // db_select_vuln_high() -> int
	FnDBSelectVulnMedium      = "db_select_vuln_medium"      // db_select_vuln_medium() -> int
	FnDBSelectVulnLow         = "db_select_vuln_low"         // db_select_vuln_low() -> int

	// JSONL import functions - import data from JSONL files
	FnDBImportAssetFromFile = "db_import_asset_from_file" // db_import_asset_from_file(workspace, file_path) -> int (count)
	FnDBImportVuln          = "db_import_vuln"            // db_import_vuln(workspace, json_data) -> bool
	FnDBImportVulnFromFile  = "db_import_vuln_from_file"  // db_import_vuln_from_file(workspace, file_path) -> int (count)

	// SARIF import functions
	FnDBImportSARIF          = "db_import_sarif"            // db_import_sarif(workspace, file_path) -> map (stats)
	FnConvertSARIFToMarkdown = "convert_sarif_to_markdown"  // convert_sarif_to_markdown(input_path, output_path) -> bool

	// Diff functions - asset and vulnerability change tracking
	FnDBAssetDiff       = "db_asset_diff"         // db_asset_diff(workspace) -> string (JSONL)
	FnDBVulnDiff        = "db_vuln_diff"          // db_vuln_diff(workspace) -> string (JSONL)
	FnDBAssetDiffToFile = "db_asset_diff_to_file" // db_asset_diff_to_file(workspace, dest) -> bool
	FnDBVulnDiffToFile  = "db_vuln_diff_to_file"  // db_vuln_diff_to_file(workspace, dest) -> bool

	// Run status functions - query run records
	FnDBSelectRuns      = "run_status"         // run_status(workspace, format) -> string
	FnDBSelectRunByUUID = "run_status_by_uuid" // run_status_by_uuid(uuid, format) -> string

	// Event log management functions
	FnDBResetEventLogs = "db_reset_event_logs" // db_reset_event_logs(workspace?, topic_pattern?) -> {reset: int, total: int}
)

// AllFunctions returns a list of all available function names
func AllFunctions() []string {
	return []string{
		// File Functions
		FnFileExists,
		FnFileLength,
		FnDirLength,
		FnFileContains,
		FnRegexExtract,
		FnReadFile,
		FnReadLines,
		FnRemoveFile,
		FnRemoveFolder,
		FnRmRF,
		FnRemoveAllExcept,
		FnCreateFolder,
		FnAppendFile,
		FnMoveFile,
		FnGlob,
		FnGrepStringToFile,
		FnGrepRegexToFile,
		FnGrepString,
		FnGrepRegex,
		FnRemoveBlankLines,
		FnChunkFile,

		// String Functions
		FnTrim,
		FnTrimString,
		FnTrimLeft,
		FnTrimRight,
		FnSplit,
		FnJoin,
		FnReplace,
		FnContains,
		FnStartsWith,
		FnEndsWith,
		FnToLowerCase,
		FnToUpperCase,
		FnMatch,
		FnRegexMatch,
		FnCutWithDelim,
		FnCut,
		FnNormalizePath,
		FnGetTargetSpace,
		FnCleanSub,

		// Type Detection Functions
		FnGetTypes,
		FnIsFile,
		FnIsDir,
		FnIsGit,
		FnIsURL,
		FnIsCompress,
		FnDetectLanguage,

		// Type Conversion Functions
		FnParseInt,
		FnParseFloat,
		FnToString,
		FnToBoolean,

		// Utility Functions
		FnLen,
		FnIsEmpty,
		FnIsNotEmpty,
		FnPrintf,
		FnCatFile,
		FnExit,
		FnExecCmd,
		FnBash,
		FnSleep,
		FnCommandExists,
		FnPickValid,
		FnRunModule,
		FnRunFlow,
		FnExecPython,
		FnExecPythonFile,

		// Logging Functions
		FnLogDebug,
		FnLogInfo,
		FnLogWarn,
		FnLogError,

		// Color Printing Functions
		FnPrintGreen,
		FnPrintBlue,
		FnPrintYellow,
		FnPrintRed,

		// Runtime Variable Functions
		FnSetVar,
		FnGetVar,

		// HTTP Functions
		FnHttpRequest,
		FnHttpGet,
		FnHttpPost,
		FnGetIP,

		// LLM Functions
		FnLLMInvoke,
		FnLLMInvokeCustom,
		FnLLMConversations,

		// Generation Functions
		FnRandomString,
		FnUUID,

		// Encoding Functions
		FnBase64Encode,
		FnBase64Decode,

		// Data Query Functions
		FnJQ,
		FnJQFromFile,

		// Notification Functions
		FnNotifyTelegram,
		FnSendTelegramFile,
		FnNotifyTelegramChannel,
		FnSendTelegramFileChannel,
		FnNotifyMessageAsFileTelegram,
		FnNotifyMessageAsFileTelegramChannel,
		FnNotifyWebhook,
		FnSendWebhookEvent,

		// Event Generation Functions
		FnGenerateEvent,
		FnGenerateEventFromFile,

		// CDN/Storage Functions
		FnCdnUpload,
		FnCdnDownload,
		FnCdnExists,
		FnCdnDelete,
		FnCdnSyncUpload,
		FnCdnSyncDownload,
		FnCdnGetPresignedURL,
		FnCdnList,
		FnCdnStat,
		FnCdnRead,
		FnCdnLsTree,

		// Unix Command Wrappers
		FnSortUnix,
		FnWgetUnix,
		FnWget,
		FnGitClone,
		FnGitCloneSubfolder,
		FnZipUnix,
		FnUnzipUnix,
		FnTarUnix,
		FnUntarUnix,
		FnDiffUnix,
		FnSedStringReplace,
		FnSedRegexReplace,

		// Archive Functions (Go implementations)
		FnZipDir,
		FnUnzipDir,
		FnExtractTo,

		// Diff Functions
		FnExtractDiff,

		// Output Functions
		FnSaveContent,
		FnJSONLToCSV,
		FnCSVToJSONL,
		FnJSONLUnique,
		FnJSONLFilter,

		// URL Processing Functions
		FnInterestingUrls,
		FnGetParentURL,
		FnParseURL,
		FnQueryReplace,
		FnPathReplace,

		// Markdown Functions
		FnRenderMarkdownFromFile,
		FnPrintMarkdownFromFile,
		FnConvertJSONLToMarkdown,
		FnConvertCSVToMarkdown,
		FnRenderMarkdownReport,
		FnGenerateSecurityReport,

		// Database Functions
		FnDBUpdate,
		FnDBImportAsset,
		FnDBQuickImportAsset,
		FnDBRawInsertAsset,
		FnDBPartialImportAsset,
		FnDBPartialImportAssetFile,
		FnDBTotalURLs,
		FnDBTotalSubdomains,
		FnDBTotalAssets,
		FnDBTotalVulns,
		FnDBVulnCritical,
		FnDBVulnHigh,
		FnDBVulnMedium,
		FnDBVulnLow,
		FnDBTotalIPs,
		FnDBTotalLinks,
		FnDBTotalContent,
		FnDBTotalArchive,
		FnRuntimeExport,
		FnDBRegisterArtifact,
		FnStoreArtifact,
		FnDBSelectAssets,
		FnDBSelectAssetsFiltered,
		FnDBSelectVulnerabilities,
		FnDBSelectVulnerabilitiesFiltered,
		FnDBSelect,
		FnDBSelectToFile,
		FnDBSelectToJSONL,
		FnDBSelectTotalSubdomains,
		FnDBSelectTotalURLs,
		FnDBSelectTotalAssets,
		FnDBSelectTotalVulns,
		FnDBSelectVulnCritical,
		FnDBSelectVulnHigh,
		FnDBSelectVulnMedium,
		FnDBSelectVulnLow,

		// JSONL import functions
		FnDBImportAssetFromFile,
		FnDBImportVuln,
		FnDBImportVulnFromFile,

		// SARIF import functions
		FnDBImportSARIF,
		FnConvertSARIFToMarkdown,

		// Diff functions
		FnDBAssetDiff,
		FnDBVulnDiff,
		FnDBAssetDiffToFile,
		FnDBVulnDiffToFile,
		FnDBSelectRuns,
		FnDBSelectRunByUUID,

		// Event log management functions
		FnDBResetEventLogs,

		// Installer functions
		FnGoGetter,
		FnGoGetterWithSSHKey,
		FnNixInstall,
		FnFilepathInstaller,

		// Environment functions
		FnOsGetenv,
		FnOsSetenv,
	}
}

// FunctionInfo describes a utility function with its metadata
type FunctionInfo struct {
	Name        string // Function name (e.g., "fileExists")
	Signature   string // Full signature (e.g., "fileExists(path)")
	Description string // Human-readable description
	ReturnType  string // Return type (e.g., "bool", "string")
	Example     string // Example usage
}

// Category keys for function registry
const (
	CategoryFile            = "file"
	CategoryString          = "string"
	CategoryTypeConversion  = "type_conversion"
	CategoryUtility         = "utility"
	CategoryLogging         = "logging"
	CategoryColorPrinting   = "color_printing"
	CategoryRuntimeVars     = "runtime_vars"
	CategoryHTTP            = "http"
	CategoryGeneration      = "generation"
	CategoryEncoding        = "encoding"
	CategoryDataQuery       = "data_query"
	CategoryNotification    = "notification"
	CategoryEventGeneration = "event_generation"
	CategoryCDNStorage      = "cdn_storage"
	CategoryUnixCommands    = "unix_commands"
	CategoryArchive         = "archive"
	CategoryDiff            = "diff"
	CategoryOutput          = "output"
	CategoryURLProcessing   = "url_processing"
	CategoryMarkdown        = "markdown"
	CategoryDatabase        = "database"
	CategoryInstaller       = "installer"
	CategoryEnvironment     = "environment"
	CategoryTypeDetection   = "type_detection"
	CategoryLLM             = "llm"
)

// CategoryInfo provides display metadata for a function category
type CategoryInfo struct {
	Key        string
	Title      string
	ShortTitle string // Short version for table display
}

// CategoryOrder returns the ordered list of categories with display titles
func CategoryOrder() []CategoryInfo {
	return []CategoryInfo{
		{CategoryFile, "File Functions", "File"},
		{CategoryString, "String Functions", "String"},
		{CategoryTypeConversion, "Type Conversion", "Type"},
		{CategoryUtility, "Utility Functions", "Utility"},
		{CategoryLogging, "Logging Functions", "Logging"},
		{CategoryColorPrinting, "Color Printing Functions", "Color"},
		{CategoryRuntimeVars, "Runtime Variable Functions", "Runtime Vars"},
		{CategoryHTTP, "HTTP Functions", "HTTP"},
		{CategoryLLM, "LLM Functions", "LLM"},
		{CategoryGeneration, "Generation Functions", "Generation"},
		{CategoryEncoding, "Encoding Functions", "Encoding"},
		{CategoryDataQuery, "Data Query Functions", "Data Query"},
		{CategoryNotification, "Notification Functions", "Notification"},
		{CategoryEventGeneration, "Event Generation Functions", "Event"},
		{CategoryCDNStorage, "CDN/Storage Functions", "CDN/Storage"},
		{CategoryUnixCommands, "Unix Command Wrappers", "Unix"},
		{CategoryArchive, "Archive Functions (Go)", "Archive"},
		{CategoryDiff, "Diff Functions", "Diff"},
		{CategoryOutput, "Output Functions", "Output"},
		{CategoryURLProcessing, "URL Processing Functions", "URL"},
		{CategoryMarkdown, "Markdown Functions", "Markdown"},
		{CategoryDatabase, "Database Functions", "Database"},
		{CategoryInstaller, "Installer Functions", "Installer"},
		{CategoryEnvironment, "Environment Functions", "Environment"},
		{CategoryTypeDetection, "Type Detection Functions", "Type Detection"},
	}
}

// FunctionRegistry returns all function metadata organized by category
func FunctionRegistry() map[string][]FunctionInfo {
	return map[string][]FunctionInfo{
		CategoryFile: {
			{FnFileExists, "file_exists(path)", "Check if file exists", "bool", "file_exists('/tmp/test.txt')"},
			{FnFileLength, "file_length(path)", "Count non-empty lines in file", "int", "file_length('{{Output}}/subdomains.txt')"},
			{FnDirLength, "dir_length(path)", "Count entries in directory", "int", "dir_length('{{Output}}/screenshots')"},
			{FnFileContains, "file_contains(path, pattern)", "Check if file contains pattern", "bool", "file_contains('{{Output}}/urls.txt', 'admin')"},
			{FnRegexExtract, "regex_extract(path, pattern)", "Extract matching lines from file", "[]string", "regex_extract('{{Output}}/urls.txt', '.*api.*')"},
			{FnReadFile, "read_file(path)", "Read entire file contents", "string", "read_file('{{Output}}/config.json')"},
			{FnReadLines, "read_lines(path)", "Read file as array of lines", "[]string", "read_lines('{{Output}}/subdomains.txt')"},
			{FnRemoveFile, "remove_file(path)", "Delete a file", "bool", "remove_file('{{Output}}/temp.txt')"},
			{FnRemoveFolder, "remove_folder(path)", "Delete folder recursively", "bool", "remove_folder('{{Output}}/cache')"},
			{FnRmRF, "rm_rf(path)", "Delete file or folder recursively", "bool", "rm_rf('{{Output}}/tmp')"},
			{FnRemoveAllExcept, "remove_all_except(folder, keep_file)", "Remove everything under folder except keep_file", "bool", "remove_all_except('{{Output}}', '{{Output}}/keep.txt')"},
			{FnCreateFolder, "create_folder(path)", "Create folder recursively", "bool", "create_folder('{{Output}}/new-folder')"},
			{FnAppendFile, "append_file(dest, source)", "Append source file content into destination file", "bool", "append_file('{{Output}}/all.txt', '{{Output}}/part.txt')"},
			{FnMoveFile, "move_file(source, dest)", "Move file from source to destination (rename or copy+delete)", "bool", "move_file('{{Output}}/raw.txt', '{{Output}}/processed.txt')"},
			{FnGlob, "glob(pattern)", "List filenames matching glob pattern", "[]string", "glob('{{Output}}/*.txt')"},
			{FnGrepStringToFile, "grep_string_to_file(dest, source, str)", "Write lines containing string to destination file", "bool", "grep_string_to_file('{{Output}}/out.txt', '{{Output}}/in.txt', 'admin')"},
			{FnGrepRegexToFile, "grep_regex_to_file(dest, source, pattern)", "Write lines matching regex to destination file", "bool", "grep_regex_to_file('{{Output}}/out.txt', '{{Output}}/in.txt', '.*api.*')"},
			{FnGrepString, "grep_string(source, str)", "Return lines containing string", "string", "grep_string('{{Output}}/in.txt', 'admin')"},
			{FnGrepRegex, "grep_regex(source, pattern)", "Return lines matching regex", "string", "grep_regex('{{Output}}/in.txt', '.*api.*')"},
			{FnRemoveBlankLines, "remove_blank_lines(path)", "Remove blank lines from file in-place", "bool", "remove_blank_lines('{{Output}}/urls.txt')"},
			{FnChunkFile, "chunk_file(input, lines_per_chunk, output)", "Split file into chunks and write manifest of chunk paths", "bool", "chunk_file('{{Output}}/urls.txt', 100, '{{Output}}/url_chunks.txt')"},
		},
		CategoryString: {
			{FnTrim, "trim(str)", "Trim whitespace", "string", "trim('  hello  ')"},
			{FnTrimString, "trim_string(input, substring)", "Trim substring from both ends", "string", "trim_string('  hello  ', ' ')"},
			{FnTrimLeft, "trim_left(input, substring)", "Trim substring from left/start", "string", "trim_left('///path/to/file', '/')"},
			{FnTrimRight, "trim_right(input, substring)", "Trim substring from right/end", "string", "trim_right('example.com///', '/')"},
			{FnSplit, "split(str, delim)", "Split string by delimiter", "[]string", "split('a,b,c', ',')"},
			{FnJoin, "join(arr, delim)", "Join array with delimiter", "string", "join(['a','b','c'], ',')"},
			{FnReplace, "replace(str, old, new)", "Replace all occurrences", "string", "replace('hello', 'l', 'L')"},
			{FnContains, "contains(str, substr)", "Check if string contains substring", "bool", "contains('hello', 'ell')"},
			{FnStartsWith, "starts_with(str, prefix)", "Check if string starts with prefix", "bool", "starts_with('hello', 'he')"},
			{FnEndsWith, "ends_with(str, suffix)", "Check if string ends with suffix", "bool", "ends_with('hello.txt', '.txt')"},
			{FnToLowerCase, "to_lower_case(str)", "Convert to lowercase", "string", "to_lower_case('HELLO')"},
			{FnToUpperCase, "to_upper_case(str)", "Convert to uppercase", "string", "to_upper_case('hello')"},
			{FnMatch, "match(str, pattern)", "Check if string matches regex", "bool", "match('test123', '[0-9]+')"},
			{FnRegexMatch, "regex_match(pattern, str)", "Check if string matches regex (pattern first)", "bool", "regex_match('[0-9]+', 'test123')"},
			{FnCutWithDelim, "cut_with_delim(input, delim, field)", "Extract field by delimiter (1-indexed)", "string", "cut_with_delim('a:b:c', ':', 2)"},
			{FnCut, "cut(input, delim, field)", "Extract field by delimiter (1-indexed, alias for cut_with_delim)", "string", "cut('a:b:c', ':', 2)"},
			{FnNormalizePath, "normalize_path(input)", "Replace special chars with underscore", "string", "normalize_path('test/path:file')"},
			{FnGetTargetSpace, "get_target_space(input)", "Normalize to path-friendly format (same as {{TargetSpace}})", "string", "get_target_space('https://example.com/path')"},
			{FnCleanSub, "clean_sub(path, target?)", "Clean and deduplicate subdomains in file, optionally filter by target domain", "bool", "clean_sub('{{Output}}/subdomains.txt', 'example.com')"},
		},
		CategoryTypeConversion: {
			{FnParseInt, "parse_int(str)", "Parse string to integer", "int", "parse_int('42')"},
			{FnParseFloat, "parse_float(str)", "Parse string to float", "float", "parse_float('3.14')"},
			{FnToString, "to_string(val)", "Convert value to string", "string", "to_string(123)"},
			{FnToBoolean, "to_boolean(val)", "Convert value to boolean", "bool", "to_boolean('true')"},
		},
		CategoryUtility: {
			{FnLen, "len(val)", "Get length of string or array", "int", "len('hello')"},
			{FnIsEmpty, "is_empty(val)", "Check if value is empty", "bool", "is_empty('')"},
			{FnIsNotEmpty, "is_not_empty(val)", "Check if value is not empty", "bool", "is_not_empty('test')"},
			{FnPrintf, "printf(message)", "Print message to stdout", "void", "printf('Scan started')"},
			{FnCatFile, "cat_file(path)", "Print file content to stdout", "void", "cat_file('{{Output}}/results.txt')"},
			{FnExit, "exit(code)", "Exit scan with code", "void", "exit(1)"},
			{FnBash, "bash(command)", "Execute bash command and return output", "string", "bash('whoami')"},
			{FnExecCmd, "exec_cmd(command)", "Alias for bash(command)", "string", "exec_cmd('whoami')"},
			{FnSleep, "sleep(seconds)", "Pause for n seconds", "void", "sleep(5)"},
			{FnCommandExists, "command_exists(command)", "Check if command exists in PATH", "bool", "command_exists('nmap')"},
			{FnPickValid, "pick_valid(v1, v2, ..., v10)", "Return first valid value from up to 10 arguments", "any", "pick_valid('', '', 'hello', 'world')"},
			{FnRunModule, "run_module(module, target, params?)", "Run osmedeus module as subprocess, optional comma-separated key=value params", "string", "run_module('subdomain', 'example.com', 'threads=10,deep=true')"},
			{FnRunFlow, "run_flow(flow, target, params?)", "Run osmedeus flow as subprocess, optional comma-separated key=value params", "string", "run_flow('general', 'example.com')"},
			{FnExecPython, "exec_python(code)", "Run inline Python code via python3 -c (falls back to python)", "string", "exec_python('print(2+2)')"},
			{FnExecPythonFile, "exec_python_file(path)", "Run a Python file via python3 (falls back to python)", "string", "exec_python_file('/tmp/script.py')"},
		},
		CategoryLogging: {
			{FnLogDebug, "log_debug(message)", "Log debug message with [DEBUG] prefix", "void", "log_debug('Processing target')"},
			{FnLogInfo, "log_info(message)", "Log info message with [INFO] prefix", "void", "log_info('Scan completed')"},
			{FnLogWarn, "log_warn(message)", "Log warning message with [WARN] prefix", "void", "log_warn('Timeout hit')"},
			{FnLogError, "log_error(message)", "Log error message with [ERROR] prefix", "void", "log_error('Request failed')"},
		},
		CategoryColorPrinting: {
			{FnPrintGreen, "print_green(message)", "Print message in green color", "string", "print_green('Success!')"},
			{FnPrintBlue, "print_blue(message)", "Print message in blue color", "string", "print_blue('Processing {{Target}}')"},
			{FnPrintYellow, "print_yellow(message)", "Print message in yellow color", "string", "print_yellow('Warning: Rate limit')"},
			{FnPrintRed, "print_red(message)", "Print message in red color", "string", "print_red('Error occurred')"},
		},
		CategoryRuntimeVars: {
			{FnSetVar, "set_var(name, value)", "Set a runtime variable for later retrieval", "string", "set_var('api_url', 'https://api.example.com')"},
			{FnGetVar, "get_var(name)", "Get a runtime variable value", "string", "get_var('api_url')"},
		},
		CategoryHTTP: {
			{FnHttpRequest, "http_request(url, method, headers, body)", "Make HTTP request", "object", "http_request('https://api.example.com', 'GET', {}, '')"},
			{FnHttpGet, "http_get(url)", "HTTP GET request with structured response", "object", "http_get('https://api.example.com/data')"},
			{FnHttpPost, "http_post(url, body)", "HTTP POST request with structured response", "object", "http_post('https://api.example.com', '{\"key\":\"value\"}')"},
			{FnGetIP, "get_ip(domain_or_url)", "Resolve domain/URL to IP address (auto-parses URL hostname)", "string", "get_ip('https://example.com/path')"},
		},
		CategoryLLM: {
			{FnLLMInvoke, "llm_invoke(message)", "Simple LLM call with direct message, returns response content", "string", "llm_invoke('Analyze security posture of {{Target}}')"},
			{FnLLMInvokeCustom, "llm_invoke_custom(message, body_json)", "LLM call with custom POST body ({{message}} placeholder)", "string", "llm_invoke_custom('Summarize: {{Target}}', '{\"model\":\"gpt-4\",\"messages\":[{\"role\":\"user\",\"content\":\"{{message}}\"}]}')"},
			{FnLLMConversations, "llm_conversations(msg1, msg2, ...)", "Multi-turn conversation with 'role:content' format messages", "string", "llm_conversations('system:Be brief', 'user:Analyze {{Target}}')"},
		},
		CategoryGeneration: {
			{FnRandomString, "random_string(length)", "Generate random alphanumeric string", "string", "random_string(16)"},
			{FnUUID, "uuid()", "Generate UUID v4", "string", "uuid()"},
		},
		CategoryEncoding: {
			{FnBase64Encode, "base64_encode(str)", "Encode string to base64", "string", "base64_encode('hello')"},
			{FnBase64Decode, "base64_decode(str)", "Decode base64 string", "string", "base64_decode('aGVsbG8=')"},
		},
		CategoryDataQuery: {
			{FnJQ, "jq(jsonData, query)", "Extract data using jq syntax", "any", "jq('{\"name\":\"test\"}', '.name')"},
			{FnJQFromFile, "jq_from_file(path, query)", "Extract data using jq from JSON file", "any", "jq_from_file('{{Output}}/data.json', '.name')"},
		},
		CategoryNotification: {
			{FnNotifyTelegram, "notify_telegram(message)", "Send markdown message to Telegram", "bool", "notify_telegram('Scan finished for {{Target}}')"},
			{FnSendTelegramFile, "send_telegram_file(path, caption?)", "Send file to Telegram (supports ~ and $HOME paths)", "bool", "send_telegram_file('~/reports/scan.pdf', 'Scan report')"},
			{FnNotifyTelegramChannel, "notify_telegram_channel(channel, message)", "Send markdown message to specific Telegram channel (#name or numeric ID)", "bool", "notify_telegram_channel('#alerts', 'New finding!')"},
			{FnSendTelegramFileChannel, "send_telegram_file_channel(channel, path, caption?)", "Send file to specific Telegram channel (supports ~ and $HOME paths)", "bool", "send_telegram_file_channel('#reports', '~/reports/scan.pdf', 'Scan report')"},
			{FnNotifyMessageAsFileTelegram, "notify_message_as_file_telegram(path)", "Read file content and send as markdown message to Telegram", "bool", "notify_message_as_file_telegram('~/reports/summary.md')"},
			{FnNotifyMessageAsFileTelegramChannel, "notify_message_as_file_telegram_channel(channel, path)", "Read file content and send as markdown message to Telegram channel", "bool", "notify_message_as_file_telegram_channel('#alerts', '~/reports/summary.md')"},
			{FnNotifyWebhook, "notify_webhook(message)", "Send message to all webhooks", "bool", "notify_webhook('Scan finished for {{Target}}')"},
			{FnSendWebhookEvent, "send_webhook_event(eventType, data)", "Send event to all webhooks", "bool", "send_webhook_event('scan_complete', {target: '{{Target}}'})"},
		},
		CategoryEventGeneration: {
			{FnGenerateEvent, "generate_event(workspace, topic, source, data_type, data)", "Generate structured event with metadata", "bool", "generate_event('{{Workspace}}', 'discovery', 'subdomain-scan', 'domain', 'api.example.com')"},
			{FnGenerateEventFromFile, "generate_event_from_file(workspace, topic, source, data_type, path)", "Generate events from file (one per line)", "int", "generate_event_from_file('{{Workspace}}', 'discovery', 'amass', 'subdomain', '{{Output}}/subdomains.txt')"},
		},
		CategoryCDNStorage: {
			{FnCdnUpload, "cdn_upload(localPath, remotePath)", "Upload file to cloud storage", "bool", "cdn_upload('{{Output}}/report.zip', 'scans/{{Target}}/report.zip')"},
			{FnCdnDownload, "cdn_download(remotePath, localPath)", "Download file from cloud storage", "bool", "cdn_download('wordlists/common.txt', '/tmp/common.txt')"},
			{FnCdnExists, "cdn_exists(remotePath)", "Check if file exists in cloud storage", "bool", "cdn_exists('scans/{{Target}}/report.zip')"},
			{FnCdnDelete, "cdn_delete(remotePath)", "Delete file from cloud storage", "bool", "cdn_delete('scans/{{Target}}/old-report.zip')"},
			{FnCdnSyncUpload, "cdn_sync_upload(localDir, remotePrefix)", "Sync local directory to cloud storage (delta)", "JSON string", "cdn_sync_upload('{{Output}}', 'scans/{{Target}}/')"},
			{FnCdnSyncDownload, "cdn_sync_download(remotePrefix, localDir)", "Sync cloud storage to local directory (delta)", "JSON string", "cdn_sync_download('base-setup/', '{{BaseFolder}}')"},
			{FnCdnGetPresignedURL, "cdn_get_presigned_url(remotePath, expiryMins?)", "Generate presigned URL for file access", "string", "cdn_get_presigned_url('report.zip', 60)"},
			{FnCdnList, "cdn_list(pattern?)", "List files with metadata from cloud storage (supports glob patterns)", "JSON string", "cdn_list('scans/')"},
			{FnCdnStat, "cdn_stat(remotePath)", "Get file metadata from cloud storage", "JSON string", "cdn_stat('scans/target/report.zip')"},
			{FnCdnRead, "cdn_read(remotePath)", "Read file content from cloud storage", "string", "cdn_read('config/settings.yaml')"},
			{FnCdnLsTree, "cdn_ls_tree(prefix?, depth?)", "List files from cloud storage in tree format", "string", "cdn_ls_tree('scans/', 2)"},
		},
		CategoryUnixCommands: {
			{FnSortUnix, "sort_unix(input, output?)", "Sort file with LC_ALL=C sort -u", "bool", "sort_unix('{{Output}}/urls.txt')"},
			{FnWgetUnix, "wget_unix(url, output?)", "Download file with wget", "bool", "wget_unix('https://example.com/file.txt', '/tmp/file.txt')"},
			{FnWget, "wget(url, output)", "Download file with pure Go (segmented parallel download)", "bool", "wget('https://example.com/large-file.tar.gz', '/tmp/file.tar.gz')"},
			{FnGitClone, "git_clone(repo, dest?)", "Clone git repository (shallow)", "bool", "git_clone('https://github.com/user/repo', '/tmp/repo')"},
			{FnGitCloneSubfolder, "git_clone_subfolder(git_url, subfolder, dest)", "Clone repo and extract specific subfolder (falls back to ZIP for GitHub)", "bool", "git_clone_subfolder('https://github.com/projectdiscovery/nuclei-templates', 'http', '/tmp/nuclei-http')"},
			{FnZipUnix, "zip_unix(source, dest)", "Create zip archive (zip -r)", "bool", "zip_unix('{{Output}}', '{{Output}}/archive.zip')"},
			{FnUnzipUnix, "unzip_unix(source, dest?)", "Extract zip archive (unzip)", "bool", "unzip_unix('/tmp/archive.zip', '/tmp/extracted')"},
			{FnTarUnix, "tar_unix(source, dest)", "Create tar.gz archive (tar -czf)", "bool", "tar_unix('{{Output}}', '{{Output}}/archive.tar.gz')"},
			{FnUntarUnix, "untar_unix(source, dest?)", "Extract tar.gz archive (tar -xzf)", "bool", "untar_unix('/tmp/archive.tar.gz', '/tmp/extracted')"},
			{FnDiffUnix, "diff_unix(file1, file2, output?)", "Compare files with diff command", "string", "diff_unix('old.txt', 'new.txt', 'diff.txt')"},
			{FnSedStringReplace, "sed_string_replace(sed_syntax, source, dest)", "String replacement with sed s/old/new/g syntax", "bool", "sed_string_replace('s/http/https/g', '{{Output}}/urls.txt', '{{Output}}/urls-fixed.txt')"},
			{FnSedRegexReplace, "sed_regex_replace(sed_syntax, source, dest)", "Regex replacement with sed s/pattern/repl/g syntax", "bool", "sed_regex_replace('s/[0-9]+/NUM/g', '{{Output}}/data.txt', '{{Output}}/data-clean.txt')"},
		},
		CategoryArchive: {
			{FnZipDir, "zip_dir(source, dest)", "Zip directory using Go archive/zip", "bool", "zip_dir('{{Output}}', '{{Output}}/archive.zip')"},
			{FnUnzipDir, "unzip_dir(source, dest)", "Unzip archive using Go archive/zip", "bool", "unzip_dir('/tmp/archive.zip', '/tmp/extracted')"},
			{FnExtractTo, "extract_to(source, dest)", "Auto-detect archive format (.zip, .tar.gz, .tar.bz2, .tar.xz, .tgz) and extract to dest (removes dest first)", "bool", "extract_to('/tmp/repo.tar.gz', '/tmp/repo')"},
		},
		CategoryDiff: {
			{FnExtractDiff, "extract_diff(file1, file2)", "Lines only in file2 (new content)", "string", "extract_diff('{{Output}}/old-subs.txt', '{{Output}}/new-subs.txt')"},
		},
		CategoryOutput: {
			{FnSaveContent, "save_content(content, path)", "Save string content to file", "bool", "save_content('hello', '{{Output}}/greeting.txt')"},
			{FnJSONLToCSV, "jsonl_to_csv(source, dest)", "Convert JSONL file to CSV", "bool", "jsonl_to_csv('{{Output}}/assets.jsonl', '{{Output}}/assets.csv')"},
			{FnCSVToJSONL, "csv_to_jsonl(source, dest)", "Convert CSV file to JSONL", "bool", "csv_to_jsonl('{{Output}}/assets.csv', '{{Output}}/assets.jsonl')"},
			{FnJSONLUnique, "jsonl_unique(source, dest, fields)", "Deduplicate JSONL by hashing selected fields", "bool", "jsonl_unique('{{Output}}/httpx.jsonl', '{{Output}}/httpx.unique.jsonl', ['status','words','lines'])"},
			{FnJSONLFilter, "jsonl_filter(source, dest, fields)", "Filter JSONL to selected fields (comma or array)", "bool", "jsonl_filter('{{Output}}/httpx.jsonl', '{{Output}}/httpx.filtered.jsonl', 'host,status,hash.body_sha256')"},
		},
		CategoryURLProcessing: {
			{FnInterestingUrls, "interesting_urls(src, dest, json_field?)", "Deduplicate URLs by hostname+path+params, filter static files and noise patterns", "bool", "interesting_urls('{{Output}}/all-urls.txt', '{{Output}}/interesting-urls.txt', 'url')"},
			{FnGetParentURL, "get_parent_url(url)", "Strip last path component and return parent directory URL", "string", "get_parent_url('https://example.com/path/file.php')"},
			{FnParseURL, "parse_url(url, format)", "Format URL using directives: %s(scheme) %d(domain) %S(subdomain) %r(root) %t(tld) %P(port) %p(path) %e(ext) %q(query) %f(fragment) %a(authority)", "string", "parse_url('https://sub.example.com/path', '%S.%r')"},
			{FnQueryReplace, "query_replace(url, value, mode?)", "Replace all query param values; mode: 'replace' (default) or 'append'", "string", "query_replace('https://example.com?a=1&b=2', 'test')"},
			{FnPathReplace, "path_replace(url, value, position?)", "Replace path segment at position (1-indexed); 0 replaces all", "string", "path_replace('https://example.com/a/b/c', 'new', 2)"},
		},
		CategoryMarkdown: {
			{FnRenderMarkdownFromFile, "render_markdown_from_file(path)", "Render markdown with terminal styling", "string", "render_markdown_from_file('{{Output}}/report.md')"},
			{FnPrintMarkdownFromFile, "print_markdown_from_file(path)", "Print markdown with syntax highlighting", "void", "print_markdown_from_file('{{Output}}/summary.md')"},
			{FnConvertJSONLToMarkdown, "convert_jsonl_to_markdown(input_path, output_path)", "Convert JSONL to markdown table and write to file", "bool", "convert_jsonl_to_markdown('{{Output}}/assets.jsonl', '{{Output}}/assets.md')"},
			{FnConvertCSVToMarkdown, "convert_csv_to_markdown(path)", "Convert CSV to markdown table", "string", "convert_csv_to_markdown('{{Output}}/data.csv')"},
			{FnRenderMarkdownReport, "render_markdown_report(template_path, output_path)", "Render markdown template with osm-func blocks", "bool", "render_markdown_report('{{Templates}}/report.md', '{{Output}}/report.md')"},
			{FnGenerateSecurityReport, "generate_security_report(template_path)", "Generate security report from template to {{Output}}/security-report.md and register as artifact", "bool", "generate_security_report('{{MarkdownTemplates}}/security-report-template.md')"},
			{FnConvertSARIFToMarkdown, "convert_sarif_to_markdown(input_path, output_path)", "Convert SARIF file to markdown table with severity, location, title, description", "bool", "convert_sarif_to_markdown('{{Output}}/semgrep.sarif', '{{Output}}/semgrep.md')"},
		},
		CategoryDatabase: {
			{FnDBRegisterArtifact, "register_artifact(path, type?)", "Register file as scan artifact", "bool", "register_artifact('{{Output}}/nuclei.json', 'nuclei')"},
			{FnStoreArtifact, "store_artifact(path)", "Store file as run artifact for current workspace", "bool", "store_artifact('{{Output}}/report.md')"},
			{FnDBUpdate, "db_update(table, key, field, value)", "Update database field", "bool", "db_update('workspaces', '{{Workspace}}', 'status', 'completed')"},
			{FnDBImportAsset, "db_import_asset(workspace, json)", "Import asset from JSON (upsert)", "bool", "db_import_asset('{{Workspace}}', '{\"asset_value\":\"sub.example.com\"}')"},
			{FnDBQuickImportAsset, "db_quick_import_asset(workspace, asset_value, asset_type?)", "Quick import asset without JSON, creates db.new.asset event for new assets", "bool", "db_quick_import_asset('{{Workspace}}', 'sub.example.com', 'domain')"},
			{FnDBRawInsertAsset, "db_raw_insert_asset(workspace, json)", "Insert asset from JSON (pure insert)", "int", "db_raw_insert_asset('{{Workspace}}', '{\"asset_value\":\"api.example.com\"}')"},
			{FnDBPartialImportAsset, "db_partial_import_asset(workspace, asset_type, asset_value)", "Import asset with only workspace/type/value (no JSON)", "bool", "db_partial_import_asset('{{Workspace}}', 'domain', 'sub.example.com')"},
			{FnDBPartialImportAssetFile, "db_partial_import_asset_file(workspace, asset_type, file_path)", "Import assets from file line-by-line with type", "int", "db_partial_import_asset_file('{{Workspace}}', 'domain', '{{Output}}/subdomains.txt')"},
			{FnDBTotalURLs, "db_total_urls(path)", "Count lines, update workspace URLs", "int", "db_total_urls('{{Output}}/urls.txt')"},
			{FnDBTotalSubdomains, "db_total_subdomains(path)", "Count lines, update workspace subdomains", "int", "db_total_subdomains('{{Output}}/subdomains.txt')"},
			{FnDBTotalAssets, "db_total_assets(path)", "Count lines, update workspace assets", "int", "db_total_assets('{{Output}}/assets.txt')"},
			{FnDBTotalVulns, "db_total_vulns(path)", "Count lines, update workspace vulns", "int", "db_total_vulns('{{Output}}/vulns.txt')"},
			{FnDBVulnCritical, "db_vuln_critical(path)", "Count critical vulns", "int", "db_vuln_critical('{{Output}}/nuclei.json')"},
			{FnDBVulnHigh, "db_vuln_high(path)", "Count high vulns", "int", "db_vuln_high('{{Output}}/nuclei.json')"},
			{FnDBVulnMedium, "db_vuln_medium(path)", "Count medium vulns", "int", "db_vuln_medium('{{Output}}/nuclei.json')"},
			{FnDBVulnLow, "db_vuln_low(path)", "Count low vulns", "int", "db_vuln_low('{{Output}}/nuclei.json')"},
			{FnDBTotalIPs, "db_total_ips(path)", "Count lines, update workspace IPs (+=, 0 to reset)", "int", "db_total_ips('{{Output}}/ips.txt')"},
			{FnDBTotalLinks, "db_total_links(path)", "Count lines, update workspace links (+=, 0 to reset)", "int", "db_total_links('{{Output}}/links.txt')"},
			{FnDBTotalContent, "db_total_content(path)", "Count lines, update workspace content (+=, 0 to reset)", "int", "db_total_content('{{Output}}/content.txt')"},
			{FnDBTotalArchive, "db_total_archive(path)", "Count lines, update workspace archive (+=, 0 to reset)", "int", "db_total_archive('{{Output}}/archive.txt')"},
			{FnRuntimeExport, "runtime_export()", "Export scan+workspace to run-state.json", "bool", "runtime_export()"},
			{FnDBSelectAssets, "db_select_assets(workspace, format)", "Select assets (markdown/jsonl)", "string", "db_select_assets('{{Workspace}}', 'markdown')"},
			{FnDBSelectAssetsFiltered, "db_select_assets_filtered(workspace, status_code, asset_type, format)", "Select assets with filters", "string", "db_select_assets_filtered('{{Workspace}}', '200', 'subdomain', 'jsonl')"},
			{FnDBSelectVulnerabilities, "db_select_vulnerabilities(workspace, format)", "Select vulnerabilities (markdown/jsonl)", "string", "db_select_vulnerabilities('{{Workspace}}', 'markdown')"},
			{FnDBSelectVulnerabilitiesFiltered, "db_select_vulnerabilities_filtered(workspace, severity, asset_value, format)", "Select vulns with filters", "string", "db_select_vulnerabilities_filtered('{{Workspace}}', 'critical', '', 'jsonl')"},
			{FnDBSelect, "db_select(sql_query, format)", "Execute SELECT query (markdown/jsonl)", "string", "db_select('SELECT * FROM assets LIMIT 10', 'markdown')"},
			{FnDBSelectToFile, "db_select_to_file(sql_query, dest)", "Execute SELECT and write markdown to file", "bool", "db_select_to_file('SELECT * FROM assets', '{{Output}}/assets.md')"},
			{FnDBSelectToJSONL, "db_select_to_jsonl(sql_query, fields, dest)", "Execute SELECT and write JSONL with specified fields to file", "bool", "db_select_to_jsonl('SELECT * FROM assets', 'asset_value,status_code', '{{Output}}/assets.jsonl')"},
			{FnDBSelectTotalSubdomains, "db_select_total_subdomains()", "Get total subdomains from workspace", "int", "db_select_total_subdomains()"},
			{FnDBSelectTotalURLs, "db_select_total_urls()", "Get total URLs from workspace", "int", "db_select_total_urls()"},
			{FnDBSelectTotalAssets, "db_select_total_assets()", "Get total assets from workspace", "int", "db_select_total_assets()"},
			{FnDBSelectTotalVulns, "db_select_total_vulns()", "Get total vulns from workspace", "int", "db_select_total_vulns()"},
			{FnDBSelectVulnCritical, "db_select_vuln_critical()", "Get critical vuln count from workspace", "int", "db_select_vuln_critical()"},
			{FnDBSelectVulnHigh, "db_select_vuln_high()", "Get high vuln count from workspace", "int", "db_select_vuln_high()"},
			{FnDBSelectVulnMedium, "db_select_vuln_medium()", "Get medium vuln count from workspace", "int", "db_select_vuln_medium()"},
			{FnDBSelectVulnLow, "db_select_vuln_low()", "Get low vuln count from workspace", "int", "db_select_vuln_low()"},
			{FnDBImportAssetFromFile, "db_import_asset_from_file(workspace, file_path)", "Import assets from JSONL file (httpx format)", "int", "db_import_asset_from_file('{{Workspace}}', '{{Output}}/httpx.jsonl')"},
			{FnDBImportVuln, "db_import_vuln(workspace, json_data)", "Import single vulnerability from JSON (nuclei format)", "bool", "db_import_vuln('{{Workspace}}', '{\"template-id\":\"...\",\"info\":{\"name\":\"...\",\"severity\":\"high\"}}')"},
			{FnDBImportVulnFromFile, "db_import_vuln_from_file(workspace, file_path)", "Import vulnerabilities from JSONL file (nuclei format)", "int", "db_import_vuln_from_file('{{Workspace}}', '{{Output}}/nuclei.jsonl')"},
			{FnDBImportSARIF, "db_import_sarif(workspace, file_path)", "Import vulnerabilities from SARIF file (Semgrep, Trivy, etc.)", "map", "db_import_sarif('{{Workspace}}', '{{Output}}/semgrep.sarif')"},
			{FnDBAssetDiff, "db_asset_diff(workspace)", "Get asset diff as JSONL string", "string", "db_asset_diff('{{Workspace}}')"},
			{FnDBVulnDiff, "db_vuln_diff(workspace)", "Get vulnerability diff as JSONL string", "string", "db_vuln_diff('{{Workspace}}')"},
			{FnDBAssetDiffToFile, "db_asset_diff_to_file(workspace, dest)", "Write asset diff to JSONL file", "bool", "db_asset_diff_to_file('{{Workspace}}', '{{Output}}/asset-diff.jsonl')"},
			{FnDBVulnDiffToFile, "db_vuln_diff_to_file(workspace, dest)", "Write vulnerability diff to JSONL file", "bool", "db_vuln_diff_to_file('{{Workspace}}', '{{Output}}/vuln-diff.jsonl')"},
			{FnDBSelectRuns, "run_status(workspace, format)", "Query run records by workspace. Format: markdown or jsonl", "string", "run_status('{{Workspace}}', 'markdown')"},
			{FnDBSelectRunByUUID, "run_status_by_uuid(uuid, format)", "Query run record by UUID. Format: markdown or jsonl", "string", "run_status_by_uuid('abc-123', 'jsonl')"},
			{FnDBResetEventLogs, "db_reset_event_logs(workspace?, topic_pattern?)", "Reset processed event logs to unprocessed state with optional filters", "object", "db_reset_event_logs('example.com', 'db.*')"},
		},
		CategoryInstaller: {
			{FnGoGetter, "go_getter(url, dest)", "Download files/repos using go-getter", "bool", "go_getter('https://github.com/user/repo.git?ref=main', '{{Output}}/repo')"},
			{FnGoGetterWithSSHKey, "go_getter_with_sshkey(ssh_key_path, git_url, dest)", "Clone git repo via SSH with auto-encoded key", "bool", "go_getter_with_sshkey('~/.ssh/id_rsa', 'git@github.com:user/private-repo.git', '{{Output}}/repo')"},
			{FnNixInstall, "nix_install(package, dest?)", "Install package via Nix", "bool", "nix_install('nuclei', '{{Binaries}}')"},
			{FnFilepathInstaller, "filepath_installer(local_path, tool_name, dest?)", "Install local archive or binary to binaries folder", "bool", "filepath_installer('/tmp/nuclei.tar.gz', 'nuclei', '{{Binaries}}')"},
		},
		CategoryEnvironment: {
			{FnOsGetenv, "os_getenv(name)", "Get environment variable", "string", "os_getenv('HOME')"},
			{FnOsSetenv, "os_setenv(name, value)", "Set environment variable", "bool", "os_setenv('API_KEY', 'secret')"},
		},
		CategoryTypeDetection: {
			{FnGetTypes, "get_types(input)", "Detect input type (file, folder, cidr, ip, url, domain, string)", "string", "get_types('192.168.1.0/24')"},
			{FnIsFile, "is_file(path)", "Check if path is an existing regular file", "bool", "is_file('{{Output}}/results.txt')"},
			{FnIsDir, "is_dir(path)", "Check if path is an existing directory", "bool", "is_dir('{{Output}}/screenshots')"},
			{FnIsGit, "is_git(path)", "Check if path is a git repository (contains .git folder)", "bool", "is_git('{{Output}}/repo')"},
			{FnIsURL, "is_url(input)", "Check if input is an HTTP/HTTPS URL", "bool", "is_url('https://example.com')"},
			{FnIsCompress, "is_compress(path)", "Check if path has a compressed file extension (.zip, .tar.gz, .tgz, .gz, .tar.bz2, .tar.xz)", "bool", "is_compress('archive.tar.gz')"},
			{FnDetectLanguage, "detect_language(path)", "Detect dominant programming language of a source code folder by file extensions and shebangs", "string", "detect_language('{{Output}}/repo')"},
		},
	}
}
