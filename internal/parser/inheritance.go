package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// InheritanceResolver resolves workflow inheritance chains
type InheritanceResolver struct {
	loader *Loader
	// Track workflows being resolved to detect circular dependencies
	resolving map[string]bool
	// childPath stores the directory of the current child being resolved
	// for relative parent resolution
	childPath string
}

// NewInheritanceResolver creates a new inheritance resolver
func NewInheritanceResolver(loader *Loader) *InheritanceResolver {
	return &InheritanceResolver{
		loader:    loader,
		resolving: make(map[string]bool),
	}
}

// Resolve resolves the inheritance chain for a workflow
// Returns a new workflow with all inherited fields merged
func (r *InheritanceResolver) Resolve(child *core.Workflow) (*core.Workflow, error) {
	log := logger.Get()

	if child.Extends == "" {
		return child, nil
	}

	log.Debug("Resolving inheritance",
		zap.String("child", child.Name),
		zap.String("extends", child.Extends),
	)

	// Check for circular dependency
	if r.resolving[child.Name] {
		return nil, fmt.Errorf("circular inheritance detected: %s", child.Name)
	}
	r.resolving[child.Name] = true
	defer func() { delete(r.resolving, child.Name) }()

	// Store child's directory for relative resolution
	if child.FilePath != "" && r.childPath == "" {
		r.childPath = filepath.Dir(child.FilePath)
	}

	// Load parent workflow
	parent, err := r.loadParent(child.Extends)
	if err != nil {
		return nil, fmt.Errorf("failed to load parent workflow '%s': %w", child.Extends, err)
	}

	// Recursively resolve parent if it also extends another workflow
	if parent.Extends != "" {
		parent, err = r.Resolve(parent)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parent '%s': %w", child.Extends, err)
		}
	}

	// Validate kind compatibility
	if parent.Kind != child.Kind {
		return nil, fmt.Errorf("kind mismatch: child '%s' is %s but parent '%s' is %s",
			child.Name, child.Kind, parent.Name, parent.Kind)
	}

	// Merge parent into child
	merged, err := r.merge(parent, child)
	if err != nil {
		return nil, fmt.Errorf("failed to merge workflows: %w", err)
	}

	// Track inheritance
	merged.ResolvedFrom = child.Extends

	log.Debug("Inheritance resolved",
		zap.String("workflow", merged.Name),
		zap.String("parent", child.Extends),
	)

	return merged, nil
}

// loadParent attempts to load a parent workflow by name or path
// It loads the workflow without triggering inheritance resolution (to avoid infinite loops)
// and then manually resolves inheritance using the same resolver
func (r *InheritanceResolver) loadParent(name string) (*core.Workflow, error) {
	// Parse the parent workflow without automatic inheritance resolution
	var parentPath string

	// First, try to find the workflow file
	// Try with .yaml extension in the same directory
	if r.childPath != "" {
		yamlPath := filepath.Join(r.childPath, name+".yaml")
		if _, err := os.Stat(yamlPath); err == nil {
			parentPath = yamlPath
		} else {
			// Try with .yml extension
			ymlPath := filepath.Join(r.childPath, name+".yml")
			if _, err := os.Stat(ymlPath); err == nil {
				parentPath = ymlPath
			}
		}
	}

	// If not found in same directory, try the loader's workflows directory
	if parentPath == "" {
		// Try in the loader's workflows directory
		if strings.Contains(name, "/") || strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			// Name looks like a path
			if filepath.IsAbs(name) {
				parentPath = name
			} else if r.childPath != "" {
				parentPath = filepath.Join(r.childPath, name)
			}
		}
	}

	var parent *core.Workflow
	var err error

	if parentPath != "" {
		// Load without inheritance resolution by parsing directly
		parent, err = r.loader.parser.Parse(parentPath)
		if err != nil {
			return nil, err
		}

		// Validate (but skip the module/flow validation since inheritance might fix it)
		if parent.Kind != core.KindModule && parent.Kind != core.KindFlow {
			return nil, fmt.Errorf("invalid kind in parent: %s", parent.Kind)
		}
	} else {
		// Try to find using the loader's standard search
		// But we need to parse without inheritance resolution
		// Use the loader's internal search logic
		parent, err = r.findAndParseParent(name)
		if err != nil {
			return nil, err
		}
	}

	// If parent also extends, resolve that first using the same resolver
	if parent.Extends != "" {
		parent, err = r.Resolve(parent)
		if err != nil {
			return nil, err
		}
	}

	return parent, nil
}

// findAndParseParent finds and parses a parent workflow by name without triggering
// automatic inheritance resolution
func (r *InheritanceResolver) findAndParseParent(name string) (*core.Workflow, error) {
	// Try direct path first
	if strings.Contains(name, "/") || strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
		return r.loader.parser.Parse(name)
	}

	// Search in loader's workflow directory
	workflowsDir := r.loader.workflowsDir
	files, err := r.findYAMLFiles(workflowsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan workflows directory: %w", err)
	}

	// Look for exact match
	for _, file := range files {
		baseName := filepath.Base(file)
		nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		if nameWithoutExt == name || nameWithoutExt == name+"-flow" || nameWithoutExt == name+"-module" {
			return r.loader.parser.Parse(file)
		}
	}

	return nil, fmt.Errorf("workflow not found: %s", name)
}

// findYAMLFiles recursively finds all YAML files in a directory
func (r *InheritanceResolver) findYAMLFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// merge creates a new workflow by merging parent and child
// Priority: Child direct fields > Child override > Parent
func (r *InheritanceResolver) merge(parent, child *core.Workflow) (*core.Workflow, error) {
	// Start with a clone of the parent
	merged := parent.Clone()

	// Apply child's direct fields
	merged.Name = child.Name
	merged.FilePath = child.FilePath
	merged.Checksum = child.Checksum
	merged.Extends = "" // Clear extends to prevent re-resolution

	// Apply child's description if set
	if child.Description != "" {
		merged.Description = child.Description
	}

	// Apply child's tags if set
	if len(child.Tags) > 0 {
		merged.Tags = make(core.TagList, len(child.Tags))
		copy(merged.Tags, child.Tags)
	}

	// Apply child's help if set
	if child.Help != nil {
		merged.Help = child.Help.Clone()
	}

	// Apply overrides if present
	if child.Override != nil {
		if err := r.applyOverrides(merged, child.Override); err != nil {
			return nil, err
		}
	}

	return merged, nil
}

// applyOverrides applies the override section to the merged workflow
func (r *InheritanceResolver) applyOverrides(merged *core.Workflow, override *core.WorkflowOverride) error {
	// Apply param overrides
	if len(override.Params) > 0 {
		r.mergeParams(merged, override.Params)
	}

	// Apply steps override (for modules)
	if override.Steps != nil && merged.IsModule() {
		if err := r.mergeSteps(merged, override.Steps); err != nil {
			return err
		}
	}

	// Apply modules override (for flows)
	if override.Modules != nil && merged.IsFlow() {
		if err := r.mergeModules(merged, override.Modules); err != nil {
			return err
		}
	}

	// Apply triggers override (replace entirely if set)
	if len(override.Triggers) > 0 {
		merged.Triggers = make([]core.Trigger, len(override.Triggers))
		for i, t := range override.Triggers {
			merged.Triggers[i] = *t.Clone()
		}
	}

	// Apply dependencies override (merge)
	if override.Dependencies != nil {
		r.mergeDependencies(merged, override.Dependencies)
	}

	// Apply preferences override (child overrides parent)
	if override.Preferences != nil {
		r.mergePreferences(merged, override.Preferences)
	}

	// Apply runner config override (child overrides parent)
	if override.RunnerConfig != nil {
		r.mergeRunnerConfig(merged, override.RunnerConfig)
	}

	// Apply runner type override
	if override.Runner != nil {
		merged.Runner = *override.Runner
	}

	return nil
}

// mergeParams merges parameter overrides into the workflow
func (r *InheritanceResolver) mergeParams(merged *core.Workflow, overrides map[string]*core.ParamOverride) {
	// Create a map for quick lookup
	paramMap := make(map[string]int)
	for i, p := range merged.Params {
		paramMap[p.Name] = i
	}

	// Apply overrides
	for name, override := range overrides {
		if idx, exists := paramMap[name]; exists {
			// Override existing param
			param := &merged.Params[idx]
			if override.Default != nil {
				param.Default = override.Default
			}
			if override.Type != nil {
				param.Type = *override.Type
			}
			if override.Required != nil {
				param.Required = *override.Required
			}
			if override.Generator != nil {
				param.Generator = *override.Generator
			}
		} else {
			// Add new param with override values
			newParam := core.Param{
				Name: name,
			}
			if override.Default != nil {
				newParam.Default = override.Default
			}
			if override.Type != nil {
				newParam.Type = *override.Type
			}
			if override.Required != nil {
				newParam.Required = *override.Required
			}
			if override.Generator != nil {
				newParam.Generator = *override.Generator
			}
			merged.Params = append(merged.Params, newParam)
		}
	}
}

// mergeSteps merges step overrides into the workflow based on mode
func (r *InheritanceResolver) mergeSteps(merged *core.Workflow, override *core.StepsOverride) error {
	mode := override.GetEffectiveMode()

	switch mode {
	case core.OverrideModeReplace:
		// Completely replace parent steps
		merged.Steps = make([]core.Step, len(override.Steps))
		for i, s := range override.Steps {
			merged.Steps[i] = *s.Clone()
		}

	case core.OverrideModePrepend:
		// Add child steps before parent steps
		newSteps := make([]core.Step, 0, len(override.Steps)+len(merged.Steps))
		for _, s := range override.Steps {
			newSteps = append(newSteps, *s.Clone())
		}
		newSteps = append(newSteps, merged.Steps...)
		merged.Steps = newSteps

	case core.OverrideModeAppend:
		// Add child steps after parent steps
		for _, s := range override.Steps {
			merged.Steps = append(merged.Steps, *s.Clone())
		}

	case core.OverrideModeMerge:
		// Match by name: replace matching, append new, remove specified
		if err := r.mergeStepsByName(merged, override); err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid steps override mode: %s", mode)
	}

	return nil
}

// mergeStepsByName implements the merge mode for steps
func (r *InheritanceResolver) mergeStepsByName(merged *core.Workflow, override *core.StepsOverride) error {
	// Create map of step names to indices
	stepMap := make(map[string]int)
	for i, s := range merged.Steps {
		stepMap[s.Name] = i
	}

	// Create set of steps to remove
	removeSet := make(map[string]bool)
	for _, name := range override.Remove {
		removeSet[name] = true
	}

	// Apply replacements
	for _, replacement := range override.Replace {
		if idx, exists := stepMap[replacement.Name]; exists {
			merged.Steps[idx] = *replacement.Clone()
		}
	}

	// Remove steps marked for removal
	if len(removeSet) > 0 {
		newSteps := make([]core.Step, 0, len(merged.Steps))
		for _, s := range merged.Steps {
			if !removeSet[s.Name] {
				newSteps = append(newSteps, s)
			}
		}
		merged.Steps = newSteps

		// Rebuild step map after removal
		stepMap = make(map[string]int)
		for i, s := range merged.Steps {
			stepMap[s.Name] = i
		}
	}

	// Append new steps (those not already in the workflow)
	for _, s := range override.Steps {
		if _, exists := stepMap[s.Name]; !exists {
			merged.Steps = append(merged.Steps, *s.Clone())
		}
	}

	return nil
}

// mergeModules merges module overrides into the workflow based on mode
func (r *InheritanceResolver) mergeModules(merged *core.Workflow, override *core.ModulesOverride) error {
	mode := override.GetEffectiveMode()

	switch mode {
	case core.OverrideModeReplace:
		// Completely replace parent modules
		merged.Modules = make([]core.ModuleRef, len(override.Modules))
		for i, m := range override.Modules {
			merged.Modules[i] = *m.Clone()
		}

	case core.OverrideModePrepend:
		// Add child modules before parent modules
		newModules := make([]core.ModuleRef, 0, len(override.Modules)+len(merged.Modules))
		for _, m := range override.Modules {
			newModules = append(newModules, *m.Clone())
		}
		newModules = append(newModules, merged.Modules...)
		merged.Modules = newModules

	case core.OverrideModeAppend:
		// Add child modules after parent modules
		for _, m := range override.Modules {
			merged.Modules = append(merged.Modules, *m.Clone())
		}

	case core.OverrideModeMerge:
		// Match by name: replace matching, append new, remove specified
		if err := r.mergeModulesByName(merged, override); err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid modules override mode: %s", mode)
	}

	return nil
}

// mergeModulesByName implements the merge mode for modules
func (r *InheritanceResolver) mergeModulesByName(merged *core.Workflow, override *core.ModulesOverride) error {
	// Create map of module names to indices
	moduleMap := make(map[string]int)
	for i, m := range merged.Modules {
		moduleMap[m.Name] = i
	}

	// Create set of modules to remove
	removeSet := make(map[string]bool)
	for _, name := range override.Remove {
		removeSet[name] = true
	}

	// Apply replacements
	for _, replacement := range override.Replace {
		if idx, exists := moduleMap[replacement.Name]; exists {
			merged.Modules[idx] = *replacement.Clone()
		}
	}

	// Remove modules marked for removal
	if len(removeSet) > 0 {
		newModules := make([]core.ModuleRef, 0, len(merged.Modules))
		for _, m := range merged.Modules {
			if !removeSet[m.Name] {
				newModules = append(newModules, m)
			}
		}
		merged.Modules = newModules

		// Rebuild module map after removal
		moduleMap = make(map[string]int)
		for i, m := range merged.Modules {
			moduleMap[m.Name] = i
		}
	}

	// Append new modules (those not already in the workflow)
	for _, m := range override.Modules {
		if _, exists := moduleMap[m.Name]; !exists {
			merged.Modules = append(merged.Modules, *m.Clone())
		}
	}

	return nil
}

// mergeDependencies merges dependency overrides (union of all items)
func (r *InheritanceResolver) mergeDependencies(merged *core.Workflow, override *core.Dependencies) {
	if merged.Dependencies == nil {
		merged.Dependencies = &core.Dependencies{}
	}

	// Merge commands (deduplicated)
	if len(override.Commands) > 0 {
		cmdSet := make(map[string]bool)
		for _, c := range merged.Dependencies.Commands {
			cmdSet[c] = true
		}
		for _, c := range override.Commands {
			if !cmdSet[c] {
				merged.Dependencies.Commands = append(merged.Dependencies.Commands, c)
				cmdSet[c] = true
			}
		}
	}

	// Merge files (deduplicated)
	if len(override.Files) > 0 {
		fileSet := make(map[string]bool)
		for _, f := range merged.Dependencies.Files {
			fileSet[f] = true
		}
		for _, f := range override.Files {
			if !fileSet[f] {
				merged.Dependencies.Files = append(merged.Dependencies.Files, f)
				fileSet[f] = true
			}
		}
	}

	// Merge variables (deduplicated by name)
	if len(override.Variables) > 0 {
		varMap := make(map[string]int)
		for i, v := range merged.Dependencies.Variables {
			varMap[v.Name] = i
		}
		for _, v := range override.Variables {
			if idx, exists := varMap[v.Name]; exists {
				// Override existing variable
				merged.Dependencies.Variables[idx] = v
			} else {
				// Add new variable
				merged.Dependencies.Variables = append(merged.Dependencies.Variables, v)
				varMap[v.Name] = len(merged.Dependencies.Variables) - 1
			}
		}
	}

	// Merge target types (deduplicated)
	if len(override.TargetTypes) > 0 {
		typeSet := make(map[core.TargetType]bool)
		for _, t := range merged.Dependencies.TargetTypes {
			typeSet[t] = true
		}
		for _, t := range override.TargetTypes {
			if !typeSet[t] {
				merged.Dependencies.TargetTypes = append(merged.Dependencies.TargetTypes, t)
				typeSet[t] = true
			}
		}
	}

	// Merge function conditions (deduplicated)
	if len(override.FunctionsConditions) > 0 {
		condSet := make(map[string]bool)
		for _, c := range merged.Dependencies.FunctionsConditions {
			condSet[c] = true
		}
		for _, c := range override.FunctionsConditions {
			if !condSet[c] {
				merged.Dependencies.FunctionsConditions = append(merged.Dependencies.FunctionsConditions, c)
				condSet[c] = true
			}
		}
	}
}

// mergePreferences merges preference overrides (child overrides parent)
func (r *InheritanceResolver) mergePreferences(merged *core.Workflow, override *core.Preferences) {
	if merged.Preferences == nil {
		merged.Preferences = &core.Preferences{}
	}

	if override.DisableNotifications != nil {
		merged.Preferences.DisableNotifications = override.DisableNotifications
	}
	if override.DisableLogging != nil {
		merged.Preferences.DisableLogging = override.DisableLogging
	}
	if override.HeuristicsCheck != nil {
		merged.Preferences.HeuristicsCheck = override.HeuristicsCheck
	}
	if override.CIOutputFormat != nil {
		merged.Preferences.CIOutputFormat = override.CIOutputFormat
	}
	if override.Silent != nil {
		merged.Preferences.Silent = override.Silent
	}
	if override.Repeat != nil {
		merged.Preferences.Repeat = override.Repeat
	}
	if override.RepeatWaitTime != nil {
		merged.Preferences.RepeatWaitTime = override.RepeatWaitTime
	}
	if override.EmptyTarget != nil {
		merged.Preferences.EmptyTarget = override.EmptyTarget
	}
}

// mergeRunnerConfig merges runner config overrides (child overrides parent)
func (r *InheritanceResolver) mergeRunnerConfig(merged *core.Workflow, override *core.RunnerConfig) {
	if merged.RunnerConfig == nil {
		merged.RunnerConfig = &core.RunnerConfig{}
	}

	// Override non-empty fields
	if override.Image != "" {
		merged.RunnerConfig.Image = override.Image
	}
	if override.Network != "" {
		merged.RunnerConfig.Network = override.Network
	}
	if override.Persistent {
		merged.RunnerConfig.Persistent = override.Persistent
	}
	if override.Host != "" {
		merged.RunnerConfig.Host = override.Host
	}
	if override.Port != 0 {
		merged.RunnerConfig.Port = override.Port
	}
	if override.User != "" {
		merged.RunnerConfig.User = override.User
	}
	if override.KeyFile != "" {
		merged.RunnerConfig.KeyFile = override.KeyFile
	}
	if override.Password != "" {
		merged.RunnerConfig.Password = override.Password
	}
	if override.WorkDir != "" {
		merged.RunnerConfig.WorkDir = override.WorkDir
	}

	// Merge env (child overrides parent for same keys)
	if len(override.Env) > 0 {
		if merged.RunnerConfig.Env == nil {
			merged.RunnerConfig.Env = make(map[string]string)
		}
		for k, v := range override.Env {
			merged.RunnerConfig.Env[k] = v
		}
	}

	// Merge volumes (append, deduplicated)
	if len(override.Volumes) > 0 {
		volSet := make(map[string]bool)
		for _, v := range merged.RunnerConfig.Volumes {
			volSet[v] = true
		}
		for _, v := range override.Volumes {
			if !volSet[v] {
				merged.RunnerConfig.Volumes = append(merged.RunnerConfig.Volumes, v)
				volSet[v] = true
			}
		}
	}
}
