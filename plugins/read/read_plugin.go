package read

import (
	"fmt"
	"strings"

	"gogent/plugins"
)

type ReadPlugin struct {
	initialized bool
	config      map[string]interface{}
}

func NewReadPlugin() *ReadPlugin {
	return &ReadPlugin{}
}

func (p *ReadPlugin) Name() string        { return "read" }
func (p *ReadPlugin) Version() string     { return "1.0.0" }
func (p *ReadPlugin) Description() string { return "Advanced file reading plugin with support for line ranges, multiple files, and search" }

func (p *ReadPlugin) Initialize(config map[string]interface{}) error {
	p.config = config
	p.initialized = true
	return nil
}

func (p *ReadPlugin) GetTools() []plugins.Tool {
	return []plugins.Tool{
		&ReadFileTool{},
		&ReadLinesTool{},
		&ReadMultipleFilesTool{},
		&SearchFileTool{},
	}
}

type ReadFileTool struct{}

func (t *ReadFileTool) Name() string { return "read_file" }
func (t *ReadFileTool) Description() string {
	return "Read entire content of a single file. Use when you need to see the full contents of a file."
}
func (t *ReadFileTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Relative path to the file from workspace root",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}
	if !plugins.FileExists(path) {
		return "", fmt.Errorf("file does not exist: %s", path)
	}
	return plugins.ReadFileContent(path)
}

type ReadLinesTool struct{}

func (t *ReadLinesTool) Name() string { return "read_lines" }
func (t *ReadLinesTool) Description() string {
	return "Read a specific range of lines from a file. Use when you only need to see a portion of a large file."
}
func (t *ReadLinesTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Relative path to the file from workspace root",
			},
			"start_line": map[string]interface{}{
				"type":        "integer",
				"description": "Starting line number (1-indexed)",
				"minimum":     1,
			},
			"end_line": map[string]interface{}{
				"type":        "integer",
				"description": "Ending line number (inclusive, 1-indexed). Use -1 for end of file",
				"minimum":     -1,
			},
		},
		"required": []string{"path", "start_line"},
	}
}

func toInt(v interface{}) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case float64:
		return int(val), nil
	case int64:
		return int(val), nil
	default:
		return 0, fmt.Errorf("value must be an integer, got %T", v)
	}
}

func (t *ReadLinesTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}
	if !plugins.FileExists(path) {
		return "", fmt.Errorf("file does not exist: %s", path)
	}

	startLine, err := toInt(args["start_line"])
	if err != nil {
		return "", fmt.Errorf("start_line: %v", err)
	}
	if startLine < 1 {
		return "", fmt.Errorf("start_line must be >= 1")
	}

	endLine := -1
	if val, ok := args["end_line"]; ok {
		endLine, err = toInt(val)
		if err != nil {
			return "", fmt.Errorf("end_line: %v", err)
		}
	}

	return plugins.ReadFileLines(path, startLine, endLine)
}

type ReadMultipleFilesTool struct{}

func (t *ReadMultipleFilesTool) Name() string { return "read_multiple_files" }
func (t *ReadMultipleFilesTool) Description() string {
	return "Read multiple files at once. Use when you need to compare or view several files together."
}
func (t *ReadMultipleFilesTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"paths": map[string]interface{}{
				"type":        "array",
				"description": "Array of relative file paths from workspace root",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"paths"},
	}
}

func (t *ReadMultipleFilesTool) Execute(args map[string]interface{}) (string, error) {
	pathsInterface, ok := args["paths"].([]interface{})
	if !ok {
		return "", fmt.Errorf("paths must be an array")
	}

	var paths []string
	for _, p := range pathsInterface {
		if pathStr, ok := p.(string); ok {
			paths = append(paths, pathStr)
		}
	}

	if len(paths) == 0 {
		return "", fmt.Errorf("no valid paths provided")
	}

	var results []string
	for _, path := range paths {
		content, err := plugins.ReadFileContent(path)
		if err != nil {
			results = append(results, fmt.Sprintf("=== %s ===\nError: %v", path, err))
			continue
		}
		results = append(results, fmt.Sprintf("=== %s ===\n%s", path, content))
	}

	return strings.Join(results, "\n\n"), nil
}

type SearchFileTool struct{}

func (t *SearchFileTool) Name() string { return "search_file" }
func (t *SearchFileTool) Description() string {
	return "Search for a pattern within a file and return matching lines with context. Use when you need to find specific content in a file."
}
func (t *SearchFileTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Relative path to the file from workspace root",
			},
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Search pattern (supports basic string matching)",
			},
			"context_lines": map[string]interface{}{
				"type":        "integer",
				"description": "Number of context lines before and after each match",
				"minimum":     0,
				"default":     2,
			},
			"case_sensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether search should be case sensitive",
				"default":     false,
			},
		},
		"required": []string{"path", "pattern"},
	}
}

func (t *SearchFileTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}
	if !plugins.FileExists(path) {
		return "", fmt.Errorf("file does not exist: %s", path)
	}

	pattern, ok := args["pattern"].(string)
	if !ok {
		return "", fmt.Errorf("pattern must be a string")
	}

	contextLines := 2
	if val, ok := args["context_lines"]; ok {
		ctx, err := toInt(val)
		if err != nil {
			return "", fmt.Errorf("context_lines: %v", err)
		}
		contextLines = ctx
	}

	caseSensitive := false
	if val, ok := args["case_sensitive"].(bool); ok {
		caseSensitive = val
	}

	return plugins.SearchFileContent(path, pattern, contextLines, caseSensitive)
}

func init() {
	plugins.RegisterPlugin(NewReadPlugin())
}

var _ plugins.Plugin = (*ReadPlugin)(nil)
