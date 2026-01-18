package functions

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dop251/goja"
	"github.com/itchyny/gojq"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// jq executes a jq query on JSON data
// Usage: jq(jsonData, query)
// Examples:
//   - jq('{"url": "example.com"}', '.url') -> "example.com"
//   - jq('{"items": [1, 2, 3]}', '.items[]') -> [1, 2, 3]
//   - jq('{"data": {"status": "ok"}}', '.data.status') -> "ok"
func (vf *vmFunc) jq(call goja.FunctionCall) goja.Value {
	jsonData := call.Argument(0).String()
	query := call.Argument(1).String()
	logger.Get().Debug("Calling jq", zap.String("query", query), zap.Int("dataLength", len(jsonData)))

	if jsonData == "undefined" || jsonData == "" {
		logger.Get().Warn("jq: empty JSON data provided")
		return vf.vm.ToValue("")
	}

	if query == "undefined" || query == "" {
		logger.Get().Warn("jq: empty query provided")
		return vf.vm.ToValue("")
	}

	// Parse JSON data
	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		logger.Get().Debug("jq: input is not valid JSON, treating as raw string", zap.Error(err))
		// If not valid JSON, treat as raw string
		data = jsonData
	}

	// Parse jq query
	q, err := gojq.Parse(query)
	if err != nil {
		logger.Get().Warn("jq: invalid query syntax", zap.String("query", query), zap.Error(err))
		return vf.vm.ToValue("")
	}

	// Execute query
	iter := q.Run(data)
	v, ok := iter.Next()
	if !ok {
		logger.Get().Debug("jq: query returned no results", zap.String("query", query))
		return vf.vm.ToValue("")
	}

	// Check for error
	if err, isErr := v.(error); isErr {
		logger.Get().Warn("jq: query execution error", zap.String("query", query), zap.Error(err))
		return vf.vm.ToValue(fmt.Sprintf("error: %s", err.Error()))
	}

	logger.Get().Debug("jq result", zap.String("query", query), zap.Any("result", v))
	// Convert result to appropriate type
	return vf.vm.ToValue(v)
}

func (vf *vmFunc) jqFromFile(call goja.FunctionCall) goja.Value {
	path := call.Argument(0).String()
	query := call.Argument(1).String()
	logger.Get().Debug("Calling jqFromFile", zap.String("path", path), zap.String("query", query))

	if path == "undefined" || path == "" {
		logger.Get().Warn("jqFromFile: empty path provided")
		return vf.vm.ToValue("")
	}
	if query == "undefined" || query == "" {
		logger.Get().Warn("jqFromFile: empty query provided")
		return vf.vm.ToValue("")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		logger.Get().Warn("jqFromFile: failed to read file", zap.String("path", path), zap.Error(err))
		return vf.vm.ToValue("")
	}

	v, err := JQ(string(content), query)
	if err != nil {
		logger.Get().Warn("jqFromFile: query execution error", zap.String("query", query), zap.Error(err))
		return vf.vm.ToValue(fmt.Sprintf("error: %s", err.Error()))
	}
	if v == nil {
		return vf.vm.ToValue("")
	}

	return vf.vm.ToValue(v)
}

// JQ is a standalone function for jq queries (can be used outside Otto)
// Usage: JQ(jsonData, query) -> (interface{}, error)
func JQ(jsonData string, query string) (interface{}, error) {
	if jsonData == "" || query == "" {
		return nil, nil
	}

	// Parse JSON data
	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Parse jq query
	q, err := gojq.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("invalid jq query: %w", err)
	}

	// Execute query
	iter := q.Run(data)
	v, ok := iter.Next()
	if !ok {
		return nil, nil
	}

	// Check for error
	if err, isErr := v.(error); isErr {
		return nil, err
	}

	return v, nil
}

// JQString is a convenience function that returns the result as a string
func JQString(jsonData string, query string) string {
	result, err := JQ(jsonData, query)
	if err != nil || result == nil {
		return ""
	}

	switch v := result.(type) {
	case string:
		return v
	default:
		// Convert to JSON string for complex types
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}
