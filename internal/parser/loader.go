package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// CacheEntry holds a cached workflow with metadata for invalidation
type CacheEntry struct {
	Workflow *core.Workflow
	FilePath string    // Absolute path for mtime check
	ModTime  time.Time // File modification time when cached
}

// Loader loads and caches workflows
type Loader struct {
	workflowsDir string
	modulesDir   string
	parser       *Parser
	cache        map[string]*CacheEntry
	mu           sync.RWMutex
}

// NewLoader creates a new workflow loader
func NewLoader(workflowsDir string) *Loader {
	return &Loader{
		workflowsDir: workflowsDir,
		modulesDir:   filepath.Join(workflowsDir, "modules"),
		parser:       NewParser(),
		cache:        make(map[string]*CacheEntry),
	}
}

// isCacheValid checks if a cache entry is still valid by comparing file mtime
func (l *Loader) isCacheValid(entry *CacheEntry) bool {
	if entry == nil || entry.FilePath == "" {
		return false
	}
	info, err := os.Stat(entry.FilePath)
	if err != nil {
		return false // File gone or inaccessible
	}
	return !info.ModTime().After(entry.ModTime)
}

// LoadWorkflow loads a single workflow by name or path
// If name looks like a path (contains separator or ends with .yaml/.yml), it loads by path
// Otherwise, it searches for the workflow by name in the workflows directory
func (l *Loader) LoadWorkflow(name string) (*core.Workflow, error) {
	log := logger.Get()

	log.Debug("LoadWorkflow called",
		zap.String("name", name),
		zap.String("workflows_dir", l.workflowsDir),
	)

	// If name looks like a path (contains path separator or ends with .yaml/.yml)
	if strings.Contains(name, string(filepath.Separator)) ||
		strings.Contains(name, "/") ||
		strings.HasSuffix(name, ".yaml") ||
		strings.HasSuffix(name, ".yml") {
		log.Debug("Loading workflow by path", zap.String("path", name))
		return l.LoadWorkflowByPath(name)
	}

	// Check cache first with mtime validation
	l.mu.RLock()
	if entry, ok := l.cache[name]; ok {
		if l.isCacheValid(entry) {
			l.mu.RUnlock()
			log.Debug("Workflow loaded from cache (mtime valid)", zap.String("name", name))
			return entry.Workflow, nil
		}
		log.Debug("Cache entry invalid (file modified), will re-parse", zap.String("name", name))
	}
	l.mu.RUnlock()

	log.Debug("Workflow not in cache, searching directory",
		zap.String("name", name),
		zap.String("dir", l.workflowsDir),
	)

	// Search recursively in the workflows directory
	files, err := l.findYAMLFiles(l.workflowsDir, true)
	if err != nil {
		return nil, fmt.Errorf("failed to scan workflows directory: %w", err)
	}

	log.Debug("Found YAML files", zap.Int("count", len(files)))

	// Look for exact match first (name.yaml or name.yml)
	for _, file := range files {
		baseName := filepath.Base(file)
		nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		if nameWithoutExt == name {
			log.Debug("Found exact match", zap.String("file", file))
			return l.loadAndCache(name, file)
		}
	}

	// Try with -flow or -module suffix
	for _, file := range files {
		baseName := filepath.Base(file)
		nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
		if nameWithoutExt == name+"-flow" || nameWithoutExt == name+"-module" {
			log.Debug("Found with suffix", zap.String("file", file))
			return l.loadAndCache(name, file)
		}
	}

	log.Debug("Workflow not found", zap.String("name", name))
	return nil, fmt.Errorf("workflow not found: %s", name)
}

// LoadWorkflowByPath loads a workflow from a specific path
func (l *Loader) LoadWorkflowByPath(path string) (*core.Workflow, error) {
	// If path is absolute, use it directly
	if !filepath.IsAbs(path) {
		// Check if file exists relative to CWD first
		if _, err := os.Stat(path); err == nil {
			// File exists relative to CWD, make it absolute
			absPath, err := filepath.Abs(path)
			if err == nil {
				path = absPath
			}
		} else {
			// File doesn't exist relative to CWD, try relative to workflowsDir
			path = filepath.Join(l.workflowsDir, path)
		}
	}

	// Get name from path for caching (handle both .yaml and .yml)
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, ".yaml")
	name = strings.TrimSuffix(name, ".yml")

	return l.loadAndCache(name, path)
}

// loadAndCache loads a workflow and caches it
func (l *Loader) loadAndCache(name, path string) (*core.Workflow, error) {
	log := logger.Get()

	log.Debug("Parsing workflow file",
		zap.String("name", name),
		zap.String("path", path),
	)

	workflow, err := l.parser.Parse(path)
	if err != nil {
		log.Debug("Failed to parse workflow", zap.Error(err))
		return nil, err
	}

	log.Debug("Workflow parsed",
		zap.String("name", workflow.Name),
		zap.String("kind", string(workflow.Kind)),
		zap.Int("steps", len(workflow.Steps)),
	)

	// Resolve inheritance if workflow extends another
	if workflow.Extends != "" {
		log.Debug("Resolving workflow inheritance",
			zap.String("workflow", workflow.Name),
			zap.String("extends", workflow.Extends),
		)

		resolver := NewInheritanceResolver(l)
		workflow, err = resolver.Resolve(workflow)
		if err != nil {
			log.Debug("Inheritance resolution failed", zap.Error(err))
			return nil, fmt.Errorf("inheritance resolution: %w", err)
		}

		log.Debug("Inheritance resolved",
			zap.String("workflow", workflow.Name),
			zap.String("resolved_from", workflow.ResolvedFrom),
		)
	}

	// Validate workflow (after inheritance resolution)
	if err := l.parser.Validate(workflow); err != nil {
		log.Debug("Workflow validation failed", zap.Error(err))
		return nil, err
	}

	log.Debug("Workflow validated, caching",
		zap.String("cache_key", name),
	)

	// Get file modification time for cache invalidation
	absPath, _ := filepath.Abs(path)
	var modTime time.Time
	if info, err := os.Stat(absPath); err == nil {
		modTime = info.ModTime()
	}

	// Cache the workflow with metadata
	entry := &CacheEntry{
		Workflow: workflow,
		FilePath: absPath,
		ModTime:  modTime,
	}
	l.mu.Lock()
	l.cache[name] = entry
	l.mu.Unlock()

	return workflow, nil
}

// LoadAllWorkflows loads all workflows from the configured directories recursively
func (l *Loader) LoadAllWorkflows() ([]*core.Workflow, error) {
	var workflows []*core.Workflow

	// Recursively find all YAML files
	files, err := l.findYAMLFiles(l.workflowsDir, true)
	if err != nil {
		return nil, fmt.Errorf("failed to scan workflows directory: %w", err)
	}

	for _, file := range files {
		w, err := l.LoadWorkflowByPath(file)
		if err != nil {
			// Log warning but continue
			continue
		}
		workflows = append(workflows, w)
	}

	return workflows, nil
}

// shouldSkipPath returns true for paths that are obviously not workflow files
func shouldSkipPath(path string) bool {
	// Skip hidden directories (except the workflow root itself)
	parts := strings.Split(path, string(os.PathSeparator))
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." {
			return true // Skip .github/, .gitlab/, .git/, etc.
		}
	}
	return false
}

// isWorkflowYAML checks if a YAML file contains kind: module or kind: flow
func isWorkflowYAML(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	if !strings.Contains(string(content), "kind:") {
		return false
	}
	kindPattern := regexp.MustCompile(`(?m)^kind:\s*['"]?(module|flow)['"]?\s*$`)
	return kindPattern.Match(content)
}

// findYAMLFiles finds all YAML files in a directory
func (l *Loader) findYAMLFiles(dir string, recursive bool) ([]string, error) {
	var files []string

	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
				if shouldSkipPath(path) {
					return nil // Skip hidden directories
				}
				if !isWorkflowYAML(path) {
					return nil // Skip non-workflow YAML files
				}
				files = append(files, path)
			}
			return nil
		})
		return files, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml")) {
			path := filepath.Join(dir, entry.Name())
			if shouldSkipPath(path) {
				continue // Skip hidden directories
			}
			if !isWorkflowYAML(path) {
				continue // Skip non-workflow YAML files
			}
			files = append(files, path)
		}
	}

	return files, nil
}

// ReloadWorkflows clears cache and reloads all workflows
func (l *Loader) ReloadWorkflows() error {
	l.mu.Lock()
	l.cache = make(map[string]*CacheEntry)
	l.mu.Unlock()

	_, err := l.LoadAllWorkflows()
	return err
}

// GetWorkflow returns a cached workflow by name
func (l *Loader) GetWorkflow(name string) (*core.Workflow, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	entry, ok := l.cache[name]
	if !ok || entry == nil {
		return nil, false
	}
	return entry.Workflow, true
}

// GetAllCached returns all cached workflows
func (l *Loader) GetAllCached() []*core.Workflow {
	l.mu.RLock()
	defer l.mu.RUnlock()

	workflows := make([]*core.Workflow, 0, len(l.cache))
	for _, entry := range l.cache {
		if entry != nil && entry.Workflow != nil {
			workflows = append(workflows, entry.Workflow)
		}
	}
	return workflows
}

// ListAllWorkflows recursively scans the workflow directory and returns
// all workflows categorized by their kind (flow or module)
func (l *Loader) ListAllWorkflows() (flows, modules []string, err error) {
	files, err := l.findYAMLFiles(l.workflowsDir, true) // recursive=true
	if err != nil {
		return nil, nil, err
	}

	for _, file := range files {
		wf, err := l.parser.Parse(file)
		if err != nil {
			continue // skip invalid files
		}
		name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		switch wf.Kind {
		case core.KindFlow:
			flows = append(flows, name)
		default:
			modules = append(modules, name)
		}
	}
	return flows, modules, nil
}

// ListFlows returns names of all available flows
func (l *Loader) ListFlows() ([]string, error) {
	flows, _, err := l.ListAllWorkflows()
	return flows, err
}

// ListModules returns names of all available modules
func (l *Loader) ListModules() ([]string, error) {
	_, modules, err := l.ListAllWorkflows()
	return modules, err
}

// ClearCache clears the workflow cache
func (l *Loader) ClearCache() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache = make(map[string]*CacheEntry)
}
