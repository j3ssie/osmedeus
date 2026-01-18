package template

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GeneratorFunc is a function that generates a value from arguments
type GeneratorFunc func(args ...string) (string, error)

// registerBuiltinGenerators registers all built-in generator functions
func (e *Engine) registerBuiltinGenerators() {
	e.generators[GenUUID] = generateUUID
	e.generators[GenCurrentDate] = generateCurrentDate
	e.generators[GenCurrentTimestamp] = generateTimestamp
	e.generators[GenGetEnvVar] = getEnvironmentVariable
	e.generators[GenConcat] = concatenateStrings
	e.generators[GenRandomInt] = generateRandomInt
	e.generators[GenRandomString] = generateRandomString
	e.generators[GenExecCmd] = executeCommand
	e.generators[GenToLower] = toLower
	e.generators[GenToUpper] = toUpper
	e.generators[GenTrim] = trimString
	e.generators[GenReplace] = replaceString
	e.generators[GenSplit] = splitString
	e.generators[GenJoin] = joinString
}

// generateUUID generates a new UUID
func generateUUID(args ...string) (string, error) {
	return uuid.New().String(), nil
}

// generateCurrentDate generates the current date with optional format
// Default format: 2006-01-02
func generateCurrentDate(args ...string) (string, error) {
	format := "2006-01-02"
	if len(args) > 0 && args[0] != "" {
		format = args[0]
	}
	return time.Now().Format(format), nil
}

// generateTimestamp generates the current Unix timestamp
func generateTimestamp(args ...string) (string, error) {
	return strconv.FormatInt(time.Now().Unix(), 10), nil
}

// getEnvironmentVariable gets an environment variable with optional default
func getEnvironmentVariable(args ...string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("getEnvVar requires at least one argument")
	}

	key := args[0]
	value := os.Getenv(key)

	if value == "" && len(args) > 1 {
		value = args[1]
	}

	return value, nil
}

// concatenateStrings concatenates multiple strings
func concatenateStrings(args ...string) (string, error) {
	return strings.Join(args, ""), nil
}

// generateRandomInt generates a random integer between min and max
func generateRandomInt(args ...string) (string, error) {
	min := 0
	max := 100

	if len(args) >= 1 {
		var err error
		min, err = strconv.Atoi(args[0])
		if err != nil {
			return "", fmt.Errorf("invalid min value: %s", args[0])
		}
	}

	if len(args) >= 2 {
		var err error
		max, err = strconv.Atoi(args[1])
		if err != nil {
			return "", fmt.Errorf("invalid max value: %s", args[1])
		}
	}

	if min > max {
		min, max = max, min
	}

	value := rand.Intn(max-min+1) + min
	return strconv.Itoa(value), nil
}

// generateRandomString generates a random alphanumeric string
func generateRandomString(args ...string) (string, error) {
	length := 16
	if len(args) > 0 {
		var err error
		length, err = strconv.Atoi(args[0])
		if err != nil {
			return "", fmt.Errorf("invalid length: %s", args[0])
		}
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b), nil
}

// executeCommand executes a shell command and returns its output
func executeCommand(args ...string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("execCmd requires a command argument")
	}

	cmd := exec.Command("sh", "-c", args[0])
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// toLower converts a string to lowercase
func toLower(args ...string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	return strings.ToLower(args[0]), nil
}

// toUpper converts a string to uppercase
func toUpper(args ...string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	return strings.ToUpper(args[0]), nil
}

// trimString trims whitespace from a string
func trimString(args ...string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	return strings.TrimSpace(args[0]), nil
}

// replaceString replaces occurrences in a string
func replaceString(args ...string) (string, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("replace requires 3 arguments: input, old, new")
	}
	return strings.ReplaceAll(args[0], args[1], args[2]), nil
}

// splitString splits a string by delimiter and returns nth element
func splitString(args ...string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("split requires at least 2 arguments: input, delimiter")
	}

	parts := strings.Split(args[0], args[1])

	index := 0
	if len(args) >= 3 {
		var err error
		index, err = strconv.Atoi(args[2])
		if err != nil {
			return "", fmt.Errorf("invalid index: %s", args[2])
		}
	}

	if index < 0 || index >= len(parts) {
		return "", nil
	}

	return parts[index], nil
}

// joinString joins strings with a delimiter
func joinString(args ...string) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("join requires at least 2 arguments")
	}

	delimiter := args[0]
	parts := args[1:]
	return strings.Join(parts, delimiter), nil
}

func init() {
	// As of Go 1.20, there is no need to seed the default random source
	// The global random generator is automatically seeded
}
