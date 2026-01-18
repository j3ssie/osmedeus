package functions

// Function name constants for easy reference and consistency
// This file serves as a central reference for all available workflow functions

// File Functions - Operations on files and directories
const (
	FnFileExists      = "fileExists"   // fileExists(path) -> bool
	FnFileLength      = "fileLength"   // fileLength(path) -> int (non-empty line count)
	FnDirLength       = "dirLength"    // dirLength(path) -> int (entry count)
	FnFileContains    = "fileContains" // fileContains(path, pattern) -> bool
	FnRegexExtract    = "regexExtract" // regexExtract(path, pattern) -> []string
	FnReadFile        = "readFile"     // readFile(path) -> string
	FnReadLines       = "readLines"    // readLines(path) -> []string
	FnRemoveFile      = "removeFile"   // removeFile(path) -> bool
	FnRemoveFolder    = "removeFolder" // removeFolder(path) -> bool
	FnRmRF            = "rm_rf"
	FnRemoveAllExcept = "remove_all_except"
	FnCreateFolder    = "createFolder" // createFolder(path) -> bool
	FnAppendFile      = "appendFile"   // appendFile(dest, source) -> bool
	FnMoveFile        = "moveFile"     // moveFile(source, dest) -> bool
	FnGlob            = "glob"         // glob(pattern) -> []string

	FnGrepStringToFile  = "grep_string_to_file"  // grep_string_to_file(dest, source, str) -> bool
	FnGrepRegexToFile   = "grep_regex_to_file"   // grep_regex_to_file(dest, source, pattern) -> bool
	FnGrepString        = "grep_string"          // grep_string(source, str) -> string
	FnGrepRegex         = "grep_regex"           // grep_regex(source, pattern) -> string
	FnRemoveBlankLines  = "remove_blank_lines"   // remove_blank_lines(path) -> bool (in-place)
)

// String Functions - String manipulation operations
const (
	FnTrim          = "trim"           // trim(str) -> string
	FnSplit         = "split"          // split(str, delim) -> []string
	FnJoin          = "join"           // join(arr, delim) -> string
	FnReplace       = "replace"        // replace(str, old, new) -> string
	FnContains      = "contains"       // contains(str, substr) -> bool
	FnStartsWith    = "startsWith"     // startsWith(str, prefix) -> bool
	FnEndsWith      = "endsWith"       // endsWith(str, suffix) -> bool
	FnToLowerCase   = "toLowerCase"    // toLowerCase(str) -> string
	FnToUpperCase   = "toUpperCase"    // toUpperCase(str) -> string
	FnMatch         = "match"          // match(str, pattern) -> bool
	FnRegexMatch    = "regex_match"    // regex_match(pattern, str) -> bool (pattern first)
	FnCutWithDelim  = "cut_with_delim" // cut_with_delim(input, delim, field) -> string (1-indexed like cut)
	FnNormalizePath = "normalize_path" // normalize_path(input) -> string (replace / | : etc with _)
	FnCleanSub      = "clean_sub"      // clean_sub(path, target?) -> bool (clean and deduplicate subdomains in file)
)

// Type Conversion Functions - Convert between types
const (
	FnParseInt   = "parseInt"   // parseInt(str) -> int
	FnParseFloat = "parseFloat" // parseFloat(str) -> float
	FnToString   = "toString"   // toString(val) -> string
	FnToBoolean  = "toBoolean"  // toBoolean(val) -> bool
)

// Utility Functions - General utility operations
const (
	FnLen        = "len"        // len(val) -> int
	FnIsEmpty    = "isEmpty"    // isEmpty(val) -> bool
	FnIsNotEmpty = "isNotEmpty" // isNotEmpty(val) -> bool
	FnPrintf     = "printf"     // printf(message) -> void (print message to stdout)
	FnCatFile    = "cat_file"   // cat_file(path) -> void (print file content to stdout)
	FnExit       = "exit"       // exit(code) -> void (exit scan with code)
	FnExecCmd    = "exec_cmd"   // exec_cmd(command) -> string (execute bash command, return stdout)
	FnSleep      = "sleep"      // sleep(seconds) -> void (pause for n seconds)
)

// Logging Functions - Log messages with level prefixes
const (
	FnLogDebug = "log_debug" // log_debug(message) -> void (print [DEBUG] message)
	FnLogInfo  = "log_info"  // log_info(message) -> void (print [INFO] message)
	FnLogWarn  = "log_warn"
	FnLogError = "log_error"
)

// HTTP and Network Functions
const (
	FnHttpRequest = "httpRequest" // httpRequest(url, method, headers, body) -> {statusCode, body, headers}
	FnHttpGet     = "http_get"    // http_get(url) -> structured JSON response
	FnHttpPost    = "http_post"   // http_post(url, body) -> structured JSON response
)

// Generation Functions - Generate random values
const (
	FnRandomString = "randomString" // randomString(length) -> string
	FnUUID         = "uuid"         // uuid() -> string (UUID v4)
)

// Encoding Functions - Encode/decode data
const (
	FnBase64Encode = "base64Encode" // base64Encode(str) -> string
	FnBase64Decode = "base64Decode" // base64Decode(str) -> string
)

// Data Query Functions - Query structured data
const (
	FnJQ         = "jq" // jq(jsonData, query) -> any (extract data using jq syntax)
	FnJQFromFile = "jq_from_file"
)

// Notification Functions - Send notifications via various channels
const (
	FnNotifyTelegram   = "notifyTelegram"   // notifyTelegram(message) -> bool
	FnSendTelegramFile = "sendTelegramFile" // sendTelegramFile(path, caption?) -> bool
	FnNotifyWebhook    = "notifyWebhook"    // notifyWebhook(message) -> bool
	FnSendWebhookEvent = "sendWebhookEvent" // sendWebhookEvent(eventType, data) -> bool
)

// CDN/Storage Functions - Cloud storage operations
const (
	FnCdnUpload   = "cdnUpload"   // cdnUpload(localPath, remotePath) -> bool
	FnCdnDownload = "cdnDownload" // cdnDownload(remotePath, localPath) -> bool
	FnCdnExists   = "cdnExists"   // cdnExists(remotePath) -> bool
	FnCdnDelete   = "cdnDelete"   // cdnDelete(remotePath) -> bool
)

// Unix Command Wrappers - Wrappers around common Unix commands
const (
	FnSortUnix         = "sortUnix"           // sortUnix(inputFile, outputFile?) -> bool (LC_ALL=C sort -u)
	FnWgetUnix         = "wgetUnix"           // wgetUnix(url, outputPath?) -> bool
	FnGitClone         = "gitClone"           // gitClone(repo, dest?) -> bool
	FnZipUnix          = "zipUnix"            // zipUnix(source, dest) -> bool (zip -r dest source)
	FnUnzipUnix        = "unzipUnix"          // unzipUnix(source, dest?) -> bool (unzip source -d dest)
	FnTarUnix          = "tarUnix"            // tarUnix(source, dest) -> bool (tar -czf dest source)
	FnUntarUnix        = "untarUnix"          // untarUnix(source, dest?) -> bool (tar -xzf source -C dest)
	FnDiffUnix         = "diffUnix"           // diffUnix(file1, file2, output?) -> string
	FnSedStringReplace = "sed_string_replace" // sed_string_replace(sed_syntax, source, dest) -> bool
	FnSedRegexReplace  = "sed_regex_replace"  // sed_regex_replace(sed_syntax, source, dest) -> bool
)

// Archive Functions - Go implementations for zip/unzip
const (
	FnZipDir   = "zip_dir"   // zip_dir(source, dest) -> bool
	FnUnzipDir = "unzip_dir" // unzip_dir(source, dest) -> bool
)

// Diff Functions - Compare files
const (
	FnExtractDiff = "extractDiff" // extractDiff(file1, file2) -> string (lines only in file2)
)

// Output Functions - Save content to files
const (
	FnSaveContent = "save_content" // save_content(content, path) -> bool
	FnJSONLToCSV  = "jsonl_to_csv"
	FnCSVToJSONL  = "csv_to_jsonl"
	FnJSONLUnique = "jsonl_unique"
	FnJSONLFilter = "jsonl_filter"
)

// URL Processing Functions - URL deduplication and filtering
const (
	FnInterestingUrls = "interesting_urls" // interesting_urls(src, dest, json_field?) -> bool
)

// Markdown Functions - Markdown rendering and conversion
const (
	FnRenderMarkdownFromFile   = "render_markdown_from_file"   // render_markdown_from_file(path) -> string (rendered markdown)
	FnPrintMarkdownFromFile    = "print_markdown_from_file"    // print_markdown_from_file(path) -> void (print with syntax highlight)
	FnConvertJSONLToMarkdown   = "convert_jsonl_to_markdown"   // convert_jsonl_to_markdown(input_path, output_path) -> bool (writes markdown table to file)
	FnConvertCSVToMarkdown     = "convert_csv_to_markdown"     // convert_csv_to_markdown(path) -> string (markdown table)
	FnRenderMarkdownReport     = "render_markdown_report"      // render_markdown_report(template_path, output_path) -> bool
	FnGenerateSecurityReport   = "generate_security_report"    // generate_security_report(template_path) -> bool (output to {{Output}}/security-report.md)
)

// Database Functions - Database update and import operations
const (
	FnDBUpdate                        = "db_update"                          // db_update(table, key, field, value) -> bool
	FnDBImportAsset                   = "db_import_asset"                    // db_import_asset(workspace, json_data) -> bool
	FnDBRawInsertAsset                = "db_raw_insert_asset"                // db_raw_insert_asset(workspace, json_data) -> int (asset ID)
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

		// String Functions
		FnTrim,
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
		FnNormalizePath,
		FnCleanSub,

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
		FnSleep,

		// Logging Functions
		FnLogDebug,
		FnLogInfo,
		FnLogWarn,
		FnLogError,

		// HTTP Functions
		FnHttpRequest,
		FnHttpGet,
		FnHttpPost,

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
		FnNotifyWebhook,
		FnSendWebhookEvent,

		// CDN/Storage Functions
		FnCdnUpload,
		FnCdnDownload,
		FnCdnExists,
		FnCdnDelete,

		// Unix Command Wrappers
		FnSortUnix,
		FnWgetUnix,
		FnGitClone,
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
		FnDBRawInsertAsset,
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
	CategoryFile           = "file"
	CategoryString         = "string"
	CategoryTypeConversion = "type_conversion"
	CategoryUtility        = "utility"
	CategoryLogging        = "logging"
	CategoryHTTP           = "http"
	CategoryGeneration     = "generation"
	CategoryEncoding       = "encoding"
	CategoryDataQuery      = "data_query"
	CategoryNotification   = "notification"
	CategoryCDNStorage     = "cdn_storage"
	CategoryUnixCommands   = "unix_commands"
	CategoryArchive        = "archive"
	CategoryDiff           = "diff"
	CategoryOutput         = "output"
	CategoryURLProcessing  = "url_processing"
	CategoryMarkdown       = "markdown"
	CategoryDatabase       = "database"
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
		{CategoryHTTP, "HTTP Functions", "HTTP"},
		{CategoryGeneration, "Generation Functions", "Generation"},
		{CategoryEncoding, "Encoding Functions", "Encoding"},
		{CategoryDataQuery, "Data Query Functions", "Data Query"},
		{CategoryNotification, "Notification Functions", "Notification"},
		{CategoryCDNStorage, "CDN/Storage Functions", "CDN/Storage"},
		{CategoryUnixCommands, "Unix Command Wrappers", "Unix"},
		{CategoryArchive, "Archive Functions (Go)", "Archive"},
		{CategoryDiff, "Diff Functions", "Diff"},
		{CategoryOutput, "Output Functions", "Output"},
		{CategoryURLProcessing, "URL Processing Functions", "URL"},
		{CategoryMarkdown, "Markdown Functions", "Markdown"},
		{CategoryDatabase, "Database Functions", "Database"},
	}
}

// FunctionRegistry returns all function metadata organized by category
func FunctionRegistry() map[string][]FunctionInfo {
	return map[string][]FunctionInfo{
		CategoryFile: {
			{FnFileExists, "fileExists(path)", "Check if file exists", "bool", "fileExists('/tmp/test.txt')"},
			{FnFileLength, "fileLength(path)", "Count non-empty lines in file", "int", "fileLength('{{Output}}/subdomains.txt')"},
			{FnDirLength, "dirLength(path)", "Count entries in directory", "int", "dirLength('{{Output}}/screenshots')"},
			{FnFileContains, "fileContains(path, pattern)", "Check if file contains pattern", "bool", "fileContains('{{Output}}/urls.txt', 'admin')"},
			{FnRegexExtract, "regexExtract(path, pattern)", "Extract matching lines from file", "[]string", "regexExtract('{{Output}}/urls.txt', '.*api.*')"},
			{FnReadFile, "readFile(path)", "Read entire file contents", "string", "readFile('{{Output}}/config.json')"},
			{FnReadLines, "readLines(path)", "Read file as array of lines", "[]string", "readLines('{{Output}}/subdomains.txt')"},
			{FnRemoveFile, "removeFile(path)", "Delete a file", "bool", "removeFile('{{Output}}/temp.txt')"},
			{FnRemoveFolder, "removeFolder(path)", "Delete folder recursively", "bool", "removeFolder('{{Output}}/cache')"},
			{FnRmRF, "rm_rf(path)", "Delete file or folder recursively", "bool", "rm_rf('{{Output}}/tmp')"},
			{FnRemoveAllExcept, "remove_all_except(folder, keep_file)", "Remove everything under folder except keep_file", "bool", "remove_all_except('{{Output}}', '{{Output}}/keep.txt')"},
			{FnCreateFolder, "createFolder(path)", "Create folder recursively", "bool", "createFolder('{{Output}}/new-folder')"},
			{FnAppendFile, "appendFile(dest, source)", "Append source file content into destination file", "bool", "appendFile('{{Output}}/all.txt', '{{Output}}/part.txt')"},
			{FnMoveFile, "moveFile(source, dest)", "Move file from source to destination (rename or copy+delete)", "bool", "moveFile('{{Output}}/raw.txt', '{{Output}}/processed.txt')"},
			{FnGlob, "glob(pattern)", "List filenames matching glob pattern", "[]string", "glob('{{Output}}/*.txt')"},
			{FnGrepStringToFile, "grep_string_to_file(dest, source, str)", "Write lines containing string to destination file", "bool", "grep_string_to_file('{{Output}}/out.txt', '{{Output}}/in.txt', 'admin')"},
			{FnGrepRegexToFile, "grep_regex_to_file(dest, source, pattern)", "Write lines matching regex to destination file", "bool", "grep_regex_to_file('{{Output}}/out.txt', '{{Output}}/in.txt', '.*api.*')"},
			{FnGrepString, "grep_string(source, str)", "Return lines containing string", "string", "grep_string('{{Output}}/in.txt', 'admin')"},
			{FnGrepRegex, "grep_regex(source, pattern)", "Return lines matching regex", "string", "grep_regex('{{Output}}/in.txt', '.*api.*')"},
			{FnRemoveBlankLines, "remove_blank_lines(path)", "Remove blank lines from file in-place", "bool", "remove_blank_lines('{{Output}}/urls.txt')"},
		},
		CategoryString: {
			{FnTrim, "trim(str)", "Trim whitespace", "string", "trim('  hello  ')"},
			{FnSplit, "split(str, delim)", "Split string by delimiter", "[]string", "split('a,b,c', ',')"},
			{FnJoin, "join(arr, delim)", "Join array with delimiter", "string", "join(['a','b','c'], ',')"},
			{FnReplace, "replace(str, old, new)", "Replace all occurrences", "string", "replace('hello', 'l', 'L')"},
			{FnContains, "contains(str, substr)", "Check if string contains substring", "bool", "contains('hello', 'ell')"},
			{FnStartsWith, "startsWith(str, prefix)", "Check if string starts with prefix", "bool", "startsWith('hello', 'he')"},
			{FnEndsWith, "endsWith(str, suffix)", "Check if string ends with suffix", "bool", "endsWith('hello.txt', '.txt')"},
			{FnToLowerCase, "toLowerCase(str)", "Convert to lowercase", "string", "toLowerCase('HELLO')"},
			{FnToUpperCase, "toUpperCase(str)", "Convert to uppercase", "string", "toUpperCase('hello')"},
			{FnMatch, "match(str, pattern)", "Check if string matches regex", "bool", "match('test123', '[0-9]+')"},
			{FnRegexMatch, "regex_match(pattern, str)", "Check if string matches regex (pattern first)", "bool", "regex_match('[0-9]+', 'test123')"},
			{FnCutWithDelim, "cut_with_delim(input, delim, field)", "Extract field by delimiter (1-indexed)", "string", "cut_with_delim('a:b:c', ':', 2)"},
			{FnNormalizePath, "normalize_path(input)", "Replace special chars with underscore", "string", "normalize_path('test/path:file')"},
			{FnCleanSub, "clean_sub(path, target?)", "Clean and deduplicate subdomains in file, optionally filter by target domain", "bool", "clean_sub('{{Output}}/subdomains.txt', 'example.com')"},
		},
		CategoryTypeConversion: {
			{FnParseInt, "parseInt(str)", "Parse string to integer", "int", "parseInt('42')"},
			{FnParseFloat, "parseFloat(str)", "Parse string to float", "float", "parseFloat('3.14')"},
			{FnToString, "toString(val)", "Convert value to string", "string", "toString(123)"},
			{FnToBoolean, "toBoolean(val)", "Convert value to boolean", "bool", "toBoolean('true')"},
		},
		CategoryUtility: {
			{FnLen, "len(val)", "Get length of string or array", "int", "len('hello')"},
			{FnIsEmpty, "isEmpty(val)", "Check if value is empty", "bool", "isEmpty('')"},
			{FnIsNotEmpty, "isNotEmpty(val)", "Check if value is not empty", "bool", "isNotEmpty('test')"},
			{FnPrintf, "printf(message)", "Print message to stdout", "void", "printf('Scan started')"},
			{FnCatFile, "cat_file(path)", "Print file content to stdout", "void", "cat_file('{{Output}}/results.txt')"},
			{FnExit, "exit(code)", "Exit scan with code", "void", "exit(1)"},
			{FnExecCmd, "exec_cmd(command)", "Execute bash command and return output", "string", "exec_cmd('whoami')"},
			{FnSleep, "sleep(seconds)", "Pause for n seconds", "void", "sleep(5)"},
		},
		CategoryLogging: {
			{FnLogDebug, "log_debug(message)", "Log debug message with [DEBUG] prefix", "void", "log_debug('Processing target')"},
			{FnLogInfo, "log_info(message)", "Log info message with [INFO] prefix", "void", "log_info('Scan completed')"},
			{FnLogWarn, "log_warn(message)", "Log warning message with [WARN] prefix", "void", "log_warn('Timeout hit')"},
			{FnLogError, "log_error(message)", "Log error message with [ERROR] prefix", "void", "log_error('Request failed')"},
		},
		CategoryHTTP: {
			{FnHttpRequest, "httpRequest(url, method, headers, body)", "Make HTTP request", "object", "httpRequest('https://api.example.com', 'GET', {}, '')"},
			{FnHttpGet, "http_get(url)", "HTTP GET request with structured response", "object", "http_get('https://api.example.com/data')"},
			{FnHttpPost, "http_post(url, body)", "HTTP POST request with structured response", "object", "http_post('https://api.example.com', '{\"key\":\"value\"}')"},
		},
		CategoryGeneration: {
			{FnRandomString, "randomString(length)", "Generate random alphanumeric string", "string", "randomString(16)"},
			{FnUUID, "uuid()", "Generate UUID v4", "string", "uuid()"},
		},
		CategoryEncoding: {
			{FnBase64Encode, "base64Encode(str)", "Encode string to base64", "string", "base64Encode('hello')"},
			{FnBase64Decode, "base64Decode(str)", "Decode base64 string", "string", "base64Decode('aGVsbG8=')"},
		},
		CategoryDataQuery: {
			{FnJQ, "jq(jsonData, query)", "Extract data using jq syntax", "any", "jq('{\"name\":\"test\"}', '.name')"},
			{FnJQFromFile, "jq_from_file(path, query)", "Extract data using jq from JSON file", "any", "jq_from_file('{{Output}}/data.json', '.name')"},
		},
		CategoryNotification: {
			{FnNotifyTelegram, "notifyTelegram(message)", "Send message to Telegram", "bool", "notifyTelegram('Scan finished for {{Target}}')"},
			{FnSendTelegramFile, "sendTelegramFile(path, caption?)", "Send file to Telegram", "bool", "sendTelegramFile('{{Output}}/report.pdf', 'Scan report')"},
			{FnNotifyWebhook, "notifyWebhook(message)", "Send message to all webhooks", "bool", "notifyWebhook('Scan finished for {{Target}}')"},
			{FnSendWebhookEvent, "sendWebhookEvent(eventType, data)", "Send event to all webhooks", "bool", "sendWebhookEvent('scan_complete', {target: '{{Target}}'})"},
		},
		CategoryCDNStorage: {
			{FnCdnUpload, "cdnUpload(localPath, remotePath)", "Upload file to cloud storage", "bool", "cdnUpload('{{Output}}/report.zip', 'scans/{{Target}}/report.zip')"},
			{FnCdnDownload, "cdnDownload(remotePath, localPath)", "Download file from cloud storage", "bool", "cdnDownload('wordlists/common.txt', '/tmp/common.txt')"},
			{FnCdnExists, "cdnExists(remotePath)", "Check if file exists in cloud storage", "bool", "cdnExists('scans/{{Target}}/report.zip')"},
			{FnCdnDelete, "cdnDelete(remotePath)", "Delete file from cloud storage", "bool", "cdnDelete('scans/{{Target}}/old-report.zip')"},
		},
		CategoryUnixCommands: {
			{FnSortUnix, "sortUnix(input, output?)", "Sort file with LC_ALL=C sort -u", "bool", "sortUnix('{{Output}}/urls.txt')"},
			{FnWgetUnix, "wgetUnix(url, output?)", "Download file with wget", "bool", "wgetUnix('https://example.com/file.txt', '/tmp/file.txt')"},
			{FnGitClone, "gitClone(repo, dest?)", "Clone git repository (shallow)", "bool", "gitClone('https://github.com/user/repo', '/tmp/repo')"},
			{FnZipUnix, "zipUnix(source, dest)", "Create zip archive (zip -r)", "bool", "zipUnix('{{Output}}', '{{Output}}/archive.zip')"},
			{FnUnzipUnix, "unzipUnix(source, dest?)", "Extract zip archive (unzip)", "bool", "unzipUnix('/tmp/archive.zip', '/tmp/extracted')"},
			{FnTarUnix, "tarUnix(source, dest)", "Create tar.gz archive (tar -czf)", "bool", "tarUnix('{{Output}}', '{{Output}}/archive.tar.gz')"},
			{FnUntarUnix, "untarUnix(source, dest?)", "Extract tar.gz archive (tar -xzf)", "bool", "untarUnix('/tmp/archive.tar.gz', '/tmp/extracted')"},
			{FnDiffUnix, "diffUnix(file1, file2, output?)", "Compare files with diff command", "string", "diffUnix('old.txt', 'new.txt', 'diff.txt')"},
			{FnSedStringReplace, "sed_string_replace(sed_syntax, source, dest)", "String replacement with sed s/old/new/g syntax", "bool", "sed_string_replace('s/http/https/g', '{{Output}}/urls.txt', '{{Output}}/urls-fixed.txt')"},
			{FnSedRegexReplace, "sed_regex_replace(sed_syntax, source, dest)", "Regex replacement with sed s/pattern/repl/g syntax", "bool", "sed_regex_replace('s/[0-9]+/NUM/g', '{{Output}}/data.txt', '{{Output}}/data-clean.txt')"},
		},
		CategoryArchive: {
			{FnZipDir, "zip_dir(source, dest)", "Zip directory using Go archive/zip", "bool", "zip_dir('{{Output}}', '{{Output}}/archive.zip')"},
			{FnUnzipDir, "unzip_dir(source, dest)", "Unzip archive using Go archive/zip", "bool", "unzip_dir('/tmp/archive.zip', '/tmp/extracted')"},
		},
		CategoryDiff: {
			{FnExtractDiff, "extractDiff(file1, file2)", "Lines only in file2 (new content)", "string", "extractDiff('{{Output}}/old-subs.txt', '{{Output}}/new-subs.txt')"},
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
		},
		CategoryMarkdown: {
			{FnRenderMarkdownFromFile, "render_markdown_from_file(path)", "Render markdown with terminal styling", "string", "render_markdown_from_file('{{Output}}/report.md')"},
			{FnPrintMarkdownFromFile, "print_markdown_from_file(path)", "Print markdown with syntax highlighting", "void", "print_markdown_from_file('{{Output}}/summary.md')"},
			{FnConvertJSONLToMarkdown, "convert_jsonl_to_markdown(input_path, output_path)", "Convert JSONL to markdown table and write to file", "bool", "convert_jsonl_to_markdown('{{Output}}/assets.jsonl', '{{Output}}/assets.md')"},
			{FnConvertCSVToMarkdown, "convert_csv_to_markdown(path)", "Convert CSV to markdown table", "string", "convert_csv_to_markdown('{{Output}}/data.csv')"},
			{FnRenderMarkdownReport, "render_markdown_report(template_path, output_path)", "Render markdown template with osm-func blocks", "bool", "render_markdown_report('{{Templates}}/report.md', '{{Output}}/report.md')"},
			{FnGenerateSecurityReport, "generate_security_report(template_path)", "Generate security report from template to {{Output}}/security-report.md and register as artifact", "bool", "generate_security_report('{{MarkdownTemplates}}/security-report-template.md')"},
		},
		CategoryDatabase: {
			{FnDBRegisterArtifact, "register_artifact(path, type?)", "Register file as scan artifact", "bool", "register_artifact('{{Output}}/nuclei.json', 'nuclei')"},
			{FnStoreArtifact, "store_artifact(path)", "Store file as run artifact for current workspace", "bool", "store_artifact('{{Output}}/report.md')"},
			{FnDBUpdate, "db_update(table, key, field, value)", "Update database field", "bool", "db_update('workspaces', '{{Workspace}}', 'status', 'completed')"},
			{FnDBImportAsset, "db_import_asset(workspace, json)", "Import asset from JSON (upsert)", "bool", "db_import_asset('{{Workspace}}', '{\"asset_value\":\"sub.example.com\"}')"},
			{FnDBRawInsertAsset, "db_raw_insert_asset(workspace, json)", "Insert asset from JSON (pure insert)", "int", "db_raw_insert_asset('{{Workspace}}', '{\"asset_value\":\"api.example.com\"}')"},
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
		},
	}
}
