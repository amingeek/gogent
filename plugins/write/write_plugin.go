package write

import (
	"fmt"
	"gogent/plugins"
	"os"
	"path/filepath"
	"strings"
)

type WritePlugin struct {
	initialized bool
	config      map[string]interface{}
}

func NewWritePlugin() *WritePlugin {
	return &WritePlugin{}
}

func (p *WritePlugin) Name() string    { return "write" }
func (p *WritePlugin) Version() string { return "1.0.0" }
func (p *WritePlugin) Description() string {
	return "Create and modify files by writing or appending content"
}

func (p *WritePlugin) Initialize(config map[string]interface{}) error {
	p.config = config
	p.initialized = true
	return nil
}

func (p *WritePlugin) GetTools() []plugins.Tool {
	return []plugins.Tool{
		&WriteFileTool{},
		&AppendFileTool{},
		&WriteMultipleFilesTool{},
	}
}

// Write File plugin:
type WriteFileTool struct{}

func (t *WriteFileTool) Name() string { return "write_file" }
func (t *WriteFileTool) Description() string {
	return "Write or overwrite the entire content of a file. Use when you need to create a new file or replace a file's content."
}

func (t *WriteFileTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to write (relative to workspace root)",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The full content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}
	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content must be a string")
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("cannot create directory %s: %v", dir, err)
		}
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("cannot write file %s: %v", path, err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len([]byte(content)), path), nil
}

// AppendFile :

type AppendFileTool struct{}

func (t *AppendFileTool) Name() string { return "append_file" }
func (t *AppendFileTool) Description() string {
	return "Append text to the end of an existing or new file. Use when you need to add content without overwriting."
}
func (t *AppendFileTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to append to",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to append",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *AppendFileTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}
	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content must be a string")
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("cannot open file %s: %v", path, err)
	}
	defer file.Close()

	n, err := file.WriteString(content)
	if err != nil {
		return "", fmt.Errorf("cannot append to file %s: %v", path, err)
	}

	return fmt.Sprintf("Successfully appended %d bytes to %s", n, path), nil
}

// write more files in 1 msg :

type WriteMultipleFilesTool struct{}

func (t *WriteMultipleFilesTool) Name() string { return "write_multiple_files" }
func (t *WriteMultipleFilesTool) Description() string {
	return "Write several files at once. Use when creating or updating multiple files in one step."
}
func (t *WriteMultipleFilesTool) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"files": map[string]interface{}{
				"type":        "array",
				"description": "List of files to write",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path":    map[string]interface{}{"type": "string", "description": "File path"},
						"content": map[string]interface{}{"type": "string", "description": "File content"},
					},
					"required": []string{"path", "content"},
				},
			},
		},
		"required": []string{"files"},
	}
}

func (t *WriteMultipleFilesTool) Execute(args map[string]interface{}) (string, error) {
	filesInterface, ok := args["files"].([]interface{})
	if !ok {
		return "", fmt.Errorf("files must be an array")
	}

	var results []string
	for i, fileItem := range filesInterface {
		fileMap, ok := fileItem.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("files[%d] must be an object", i)
		}
		path, _ := fileMap["path"].(string)
		content, _ := fileMap["content"].(string)
		if path == "" {
			return "", fmt.Errorf("files[%d].path is required", i)
		}

		dir := filepath.Dir(path)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				results = append(results, fmt.Sprintf("%s: ERROR %v", path, err))
				continue
			}
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			results = append(results, fmt.Sprintf("%s: ERROR %v", path, err))
			continue
		}
		results = append(results, fmt.Sprintf("Wrote %s (%d bytes)", path, len([]byte(content))))
	}

	return strings.Join(results, "\n"), nil
}

func init() {
	if err := plugins.RegisterPlugin(NewWritePlugin()); err != nil {
		fmt.Println("ERROR: failed to register write plugin:", err)
	}
}

var _ plugins.Plugin = (*WritePlugin)(nil)
