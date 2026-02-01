package terminal

// Terminal symbols for different states and step types
const (
	SymbolPending = "○" // Step pending
	SymbolRunning = "⏺" // Step running
	SymbolStart   = "▶" // Step started
	SymbolSuccess = "✔" // Step succeeded
	SymbolFailed  = "⏹" // Step failed
	SymbolSkipped = "◌" // Step skipped
	SymbolInfo    = "◆" // Info message
	SymbolWarning = "⚠" // Warning
	SymbolError   = "✖" // Error
	SymbolArrow   = "▷" // Flow indicator
	SymbolBullet  = "•" // List item
	SymbolDiamond = "◇" // Module indicator

	// Step type and runner symbols
	SymbolFunction = "ƒ" // Function step
	SymbolBash     = "$" // Bash/command step
	SymbolForeach  = "∀" // Foreach step (universal quantifier)
	SymbolDocker   = "🐋" // Docker runner
	SymbolSSH      = "❄" // SSH runner

	// Decorative symbols for headers and labels
	SymbolStar      = "★" // Highlighted section
	SymbolStarEmpty = "☆" // Secondary highlight
	SymbolSparkle   = "✦" // Feature/important
	SymbolSparkle2  = "✧" // Sub-feature
	SymbolFlower    = "✿" // Special
	SymbolSun       = "☼" // Bright/positive
	SymbolSnow      = "❄" // Cool/frozen
	SymbolLightning = "ϟ" // Fast/power
	SymbolMenu      = "☰" // Menu/list
	SymbolTherefore = "∴" // Result/conclusion
	SymbolCommand   = "⌘" // Command/action
	SymbolCross     = "✢" // Marker
	SymbolAsterisk  = "＊" // Note
	SymbolHeart     = "♡" // Favorite
	SymbolDiamondSm = "❖" // Small diamond
	SymbolBowtie    = "⋈" // Join/connect
)

// StepSymbol returns the appropriate colored symbol for a step status
func StepSymbol(status string) string {
	switch status {
	case "pending":
		return Gray(SymbolPending)
	case "running":
		return Cyan(SymbolRunning)
	case "success":
		return Green(SymbolSuccess)
	case "failed":
		return Red(SymbolFailed)
	case "skipped":
		return Gray(SymbolSkipped)
	default:
		return SymbolBullet
	}
}

// StepStartSymbol returns a colored start symbol
func StepStartSymbol() string {
	return Cyan(SymbolStart)
}

// StepSuccessSymbol returns a colored success symbol
func StepSuccessSymbol() string {
	return Green(SymbolSuccess)
}

// StepFailedSymbol returns a colored failed symbol
func StepFailedSymbol() string {
	return Red(SymbolFailed)
}

// StepSkippedSymbol returns a colored skipped symbol
func StepSkippedSymbol() string {
	return Gray(SymbolSkipped)
}

// StepRunningSymbol returns a colored running symbol
func StepRunningSymbol() string {
	return Cyan(SymbolRunning)
}

// InfoSymbol returns a colored info symbol
func InfoSymbol() string {
	return Cyan(SymbolInfo)
}

// WarningSymbol returns a colored warning symbol
func WarningSymbol() string {
	return Yellow(SymbolWarning)
}

// ErrorSymbol returns a colored error symbol
func ErrorSymbol() string {
	return Red(SymbolError)
}

// FlowSymbol returns a colored flow indicator
func FlowSymbol() string {
	return Cyan(SymbolArrow)
}

// ModuleSymbol returns a colored module indicator
func ModuleSymbol() string {
	return Yellow(SymbolDiamond)
}

// SectionSymbol returns a colored section symbol
func SectionSymbol() string {
	return BoldCyan(SymbolStart)
}

// SubSectionSymbol returns a colored subsection symbol
func SubSectionSymbol() string {
	return Cyan(SymbolSparkle)
}

// ResultSymbol returns a colored result symbol
func ResultSymbol() string {
	return Yellow(SymbolTherefore)
}

// ListSymbol returns a colored list/menu symbol
func ListSymbol() string {
	return Cyan(SymbolMenu)
}

// StepTypeSymbol returns the appropriate colored symbol for a step based on its type and runner
func StepTypeSymbol(stepType, runnerType string) string {
	// Check runner type first (docker/ssh take precedence)
	switch runnerType {
	case "docker":
		return Blue(SymbolDocker)
	case "ssh":
		return Cyan(SymbolSSH)
	}

	// Check step type
	switch stepType {
	case "llm":
		return Magenta(SymbolBowtie)
	case "function":
		return Cyan(SymbolFunction)
	case "remote-bash":
		return Cyan(SymbolSSH)
	case "foreach":
		return Cyan(SymbolForeach)
	case "bash", "parallel-steps":
		return Green(SymbolBash)
	default:
		return Green(SymbolBash)
	}
}

// StepCommandPrefix returns the prefix symbol for commands based on step type
func StepCommandPrefix(stepType string) string {
	switch stepType {
	case "llm":
		return SymbolBowtie // "⋈"
	case "function":
		return SymbolFunction // "ƒ"
	case "foreach":
		return SymbolForeach // "∀"
	default:
		return SymbolBash // "$"
	}
}

// LLMSymbol returns a colored LLM step symbol (⋈ in magenta)
func LLMSymbol() string {
	return Magenta(SymbolBowtie)
}
