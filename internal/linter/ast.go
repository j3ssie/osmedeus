package linter

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// ParseWorkflowAST parses a workflow YAML file and builds the AST with node mapping
func ParseWorkflowAST(path string) (*WorkflowAST, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	return ParseWorkflowASTFromContent(data, path)
}

// ParseWorkflowASTFromContent parses workflow YAML content and builds the AST
func ParseWorkflowASTFromContent(content []byte, path string) (*WorkflowAST, error) {
	// Parse the YAML into the workflow struct
	var workflow core.Workflow
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		formatted := yaml.FormatError(err, false, true)
		return nil, fmt.Errorf("YAML parse error:\n%s", formatted)
	}

	// Parse the YAML AST for line/column tracking
	file, err := parser.ParseBytes(content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML AST: %w", err)
	}

	// Build the node map for path-to-position lookups
	nodeMap := make(map[string]ast.Node)
	for _, doc := range file.Docs {
		buildNodeMap(doc.Body, "", nodeMap)
	}

	// Get the root node (first document body)
	var root ast.Node
	if len(file.Docs) > 0 {
		root = file.Docs[0].Body
	}

	return &WorkflowAST{
		Workflow: &workflow,
		FilePath: path,
		Source:   content,
		Root:     root,
		NodeMap:  nodeMap,
	}, nil
}

// GetNodePosition returns the line and column for a YAML path
// Returns 0, 0 if the path is not found
func (w *WorkflowAST) GetNodePosition(path string) (line, column int) {
	node, ok := w.NodeMap[path]
	if !ok {
		return 0, 0
	}
	return getNodePosition(node)
}

// GetNodeByPath returns the AST node for a given YAML path
func (w *WorkflowAST) GetNodeByPath(path string) ast.Node {
	return w.NodeMap[path]
}

// GetLine returns the source line at the given line number (1-based)
func (w *WorkflowAST) GetLine(lineNum int) string {
	lines := splitLines(w.Source)
	if lineNum < 1 || lineNum > len(lines) {
		return ""
	}
	return lines[lineNum-1]
}

// splitLines splits content into lines preserving line content
func splitLines(content []byte) []string {
	var lines []string
	var current []byte
	for _, b := range content {
		if b == '\n' {
			lines = append(lines, string(current))
			current = nil
		} else {
			current = append(current, b)
		}
	}
	if len(current) > 0 {
		lines = append(lines, string(current))
	}
	return lines
}

// buildNodeMap recursively builds a mapping from YAML paths to AST nodes
func buildNodeMap(node ast.Node, path string, nodeMap map[string]ast.Node) {
	if node == nil {
		return
	}

	// Store current node at path
	if path != "" {
		nodeMap[path] = node
	}

	switch n := node.(type) {
	case *ast.MappingNode:
		for _, value := range n.Values {
			buildNodeMap(value, path, nodeMap)
		}
	case *ast.MappingValueNode:
		keyStr := getKeyString(n.Key)
		var newPath string
		if path == "" {
			newPath = keyStr
		} else {
			newPath = path + "." + keyStr
		}
		nodeMap[newPath] = n
		// Also store the value node
		if n.Value != nil {
			nodeMap[newPath+".value"] = n.Value
			buildNodeMap(n.Value, newPath, nodeMap)
		}
	case *ast.SequenceNode:
		for i, value := range n.Values {
			indexPath := fmt.Sprintf("%s[%d]", path, i)
			nodeMap[indexPath] = value
			buildNodeMap(value, indexPath, nodeMap)
		}
	case *ast.DocumentNode:
		buildNodeMap(n.Body, path, nodeMap)
	case *ast.AnchorNode:
		buildNodeMap(n.Value, path, nodeMap)
	case *ast.AliasNode:
		// Aliases reference other nodes, no need to recurse
	}
}

// getKeyString extracts the string value from a key node
func getKeyString(node ast.Node) string {
	switch n := node.(type) {
	case *ast.StringNode:
		return n.Value
	case *ast.LiteralNode:
		return n.Value.Value
	default:
		return fmt.Sprintf("%v", node)
	}
}

// getNodePosition extracts line and column from an AST node
func getNodePosition(node ast.Node) (line, column int) {
	if node == nil {
		return 0, 0
	}

	token := node.GetToken()
	if token == nil {
		return 0, 0
	}

	pos := token.Position
	if pos == nil {
		return 0, 0
	}

	return pos.Line, pos.Column
}

// FindStepNode finds the AST node for a step by name
func (w *WorkflowAST) FindStepNode(stepName string) ast.Node {
	for i, step := range w.Workflow.Steps {
		if step.Name == stepName {
			path := fmt.Sprintf("steps[%d]", i)
			return w.NodeMap[path]
		}
	}
	return nil
}

// FindStepPosition finds the line/column for a step by name
func (w *WorkflowAST) FindStepPosition(stepName string) (line, column int) {
	node := w.FindStepNode(stepName)
	return getNodePosition(node)
}

// FindStepFieldPosition finds the line/column for a specific field within a step
func (w *WorkflowAST) FindStepFieldPosition(stepIndex int, field string) (line, column int) {
	path := fmt.Sprintf("steps[%d].%s", stepIndex, field)
	return w.GetNodePosition(path)
}

// FindExportPosition finds the line/column for an export variable in a step
func (w *WorkflowAST) FindExportPosition(stepIndex int, exportName string) (line, column int) {
	path := fmt.Sprintf("steps[%d].exports.%s", stepIndex, exportName)
	return w.GetNodePosition(path)
}

// GetAllPaths returns all paths in the node map (useful for debugging)
func (w *WorkflowAST) GetAllPaths() []string {
	paths := make([]string, 0, len(w.NodeMap))
	for path := range w.NodeMap {
		paths = append(paths, path)
	}
	return paths
}
