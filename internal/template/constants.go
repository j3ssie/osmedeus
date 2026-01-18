package template

// Generator name constants for template param generators
// These are used in workflow YAML files: params.generator: uuid
const (
	GenUUID             = "uuid"             // uuid() -> string (UUID v4)
	GenCurrentDate      = "currentDate"      // currentDate(format?) -> string (default: 2006-01-02)
	GenCurrentTimestamp = "currentTimestamp" // currentTimestamp() -> string (Unix timestamp)
	GenGetEnvVar        = "getEnvVar"        // getEnvVar(key, default?) -> string
	GenConcat           = "concat"           // concat(str1, str2, ...) -> string
	GenRandomInt        = "randomInt"        // randomInt(min?, max?) -> string (default: 0-100)
	GenRandomString     = "randomString"     // randomString(length?) -> string (default: 16)
	GenExecCmd          = "execCmd"          // execCmd(command) -> string (command output)
	GenToLower          = "toLower"          // toLower(str) -> string
	GenToUpper          = "toUpper"          // toUpper(str) -> string
	GenTrim             = "trim"             // trim(str) -> string
	GenReplace          = "replace"          // replace(str, old, new) -> string
	GenSplit            = "split"            // split(str, delim, index?) -> string
	GenJoin             = "join"             // join(delim, str1, str2, ...) -> string
)

// AllGenerators returns a list of all available generator names
func AllGenerators() []string {
	return []string{
		GenUUID,
		GenCurrentDate,
		GenCurrentTimestamp,
		GenGetEnvVar,
		GenConcat,
		GenRandomInt,
		GenRandomString,
		GenExecCmd,
		GenToLower,
		GenToUpper,
		GenTrim,
		GenReplace,
		GenSplit,
		GenJoin,
	}
}
