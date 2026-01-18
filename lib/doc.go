/*
Package lib provides the Osmedeus SDK for programmatic workflow execution.

This package exposes two main functionalities:
  - Run Workflow: Execute module workflows from YAML string content on a target
  - Function Evaluation: Evaluate JavaScript expressions with user-provided context

# Basic Usage

Run a workflow:

	workflowYAML := `
	name: simple-scan
	kind: module
	steps:
	  - name: echo-target
	    type: bash
	    command: echo "Scanning {{target}}"
	`
	result, err := lib.Run("example.com", workflowYAML, nil)
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Printf("Status: %s, Duration: %v\n", result.Status, result.Duration)

# Running with Options

Customize execution with RunOptions:

	opts := &lib.RunOptions{
	    Params:  map[string]string{"threads": "20"},
	    Tactic:  "aggressive",
	    Verbose: true,
	}
	result, err := lib.Run("test.com", workflowYAML, opts)

# Context and Timeout

Use RunWithContext for cancellation and timeout support:

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	result, err := lib.RunWithContext(ctx, "timeout.com", workflowYAML, nil)

# Function Evaluation

Evaluate JavaScript expressions using the built-in utility functions:

	// Check if a file exists
	exists, err := lib.Eval(`fileExists("/etc/passwd")`, nil)
	fmt.Printf("File exists: %v\n", exists)

	// Use context variables
	trimmed, err := lib.Eval(`trim(input)`, &lib.EvalOptions{
	    Context: map[string]interface{}{"input": "  hello  "},
	})
	fmt.Printf("Trimmed: %s\n", trimmed)

# Boolean Conditions

Evaluate conditions that return true/false:

	ok, err := lib.EvalCondition(`len(items) > 0`, &lib.EvalOptions{
	    Context: map[string]interface{}{"items": []string{"a", "b"}},
	})
	fmt.Printf("Has items: %v\n", ok)

# Available Functions

The Eval functions have access to all Osmedeus utility functions including:

File operations:

	fileExists(path)       - Check if file exists
	fileLength(path)       - Get file line count
	dirLength(path)        - Get directory entry count
	readFile(path)         - Read file contents
	readLines(path, n)     - Read first n lines
	removeFile(path)       - Delete a file
	createFolder(path)     - Create directory
	appendFile(path, data) - Append to file
	glob(pattern)          - Find files matching pattern

String operations:

	trim(s)                - Remove whitespace
	split(s, sep)          - Split string
	join(arr, sep)         - Join array
	replace(s, old, new)   - Replace substring
	contains(s, substr)    - Check substring
	startsWith(s, prefix)  - Check prefix
	endsWith(s, suffix)    - Check suffix
	toLowerCase(s)         - Convert to lowercase
	toUpperCase(s)         - Convert to uppercase
	match(s, pattern)      - Regex match
	regexExtract(s, pattern) - Extract regex groups

Type conversion:

	parseInt(s)            - Parse integer
	parseFloat(s)          - Parse float
	toString(v)            - Convert to string
	toBoolean(v)           - Convert to boolean

Utilities:

	len(v)                 - Get length
	isEmpty(v)             - Check if empty
	isNotEmpty(v)          - Check if not empty
	uuid()                 - Generate UUID
	randomString(n)        - Generate random string
	base64Encode(s)        - Base64 encode
	base64Decode(s)        - Base64 decode

Logging:

	log_info(msg)          - Log info message
	log_warn(msg)          - Log warning
	log_error(msg)         - Log error
	log_debug(msg)         - Log debug

# Error Handling

The package provides typed errors for different failure scenarios:

	result, err := lib.Run(target, yaml, nil)
	if err != nil {
	    if lib.IsParseError(err) {
	        fmt.Println("YAML parsing failed:", err)
	    } else if lib.IsValidationError(err) {
	        fmt.Println("Workflow validation failed:", err)
	    } else if lib.IsExecutionError(err) {
	        fmt.Println("Execution failed:", err)
	    }
	}

# Design Decisions

  - Silent by default: Library mode defaults to silent (no terminal output)
  - No database by default: Library mode skips database operations
  - Module-only: Only module workflows supported (not flows) for simplicity
  - Minimal config: Works with zero config using sensible defaults
  - Context support: All execution functions support context for cancellation
  - Thread-safe: Uses existing thread-safe internal packages
*/
package lib
