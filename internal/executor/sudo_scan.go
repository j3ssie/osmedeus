package executor

import (
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// scanStepsForSudo returns true if any step (or nested step) contains
// a command that invokes sudo.
func scanStepsForSudo(steps []core.Step) bool {
	for i := range steps {
		if stepContainsSudo(&steps[i]) {
			return true
		}
	}
	return false
}

// stepContainsSudo checks a single step and its nested children
// for sudo command usage.
func stepContainsSudo(step *core.Step) bool {
	// Check direct command fields
	if containsSudoCommand(step.Command) {
		return true
	}
	for _, cmd := range step.Commands {
		if containsSudoCommand(cmd) {
			return true
		}
	}
	for _, cmd := range step.ParallelCommands {
		if containsSudoCommand(cmd) {
			return true
		}
	}

	// Check parallel nested steps
	if len(step.ParallelSteps) > 0 && scanStepsForSudo(step.ParallelSteps) {
		return true
	}

	// Check foreach nested step
	if step.Step != nil && stepContainsSudo(step.Step) {
		return true
	}

	return false
}

// containsSudoCommand detects "sudo " at word boundaries — start of string
// or after a pipe/semicolon/ampersand/whitespace. This avoids false
// positives on words like "pseudo" or "sudoku".
func containsSudoCommand(cmd string) bool {
	if cmd == "" {
		return false
	}

	// Check if the command starts with "sudo "
	if strings.HasPrefix(cmd, "sudo ") {
		return true
	}

	// Check for sudo after common shell separators: |, ;, &&, ||, `, $(
	// We look for any of these chars followed by optional whitespace and "sudo "
	for i := 0; i < len(cmd)-5; i++ {
		c := cmd[i]
		if c == '|' || c == ';' || c == '&' || c == '`' || c == '(' {
			// Skip whitespace after the separator
			j := i + 1
			for j < len(cmd) && (cmd[j] == ' ' || cmd[j] == '\t') {
				j++
			}
			if strings.HasPrefix(cmd[j:], "sudo ") {
				return true
			}
		}
	}

	return false
}
