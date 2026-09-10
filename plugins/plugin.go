package plugins

type Tool interface {
	Name() string
	Description() string
	Parameters() interface{}
	Execute(args map[string]interface{}) (string, error)
}

type Plugin interface {
	Name() string
	Version() string
	Description() string
	GetTools() []Tool
	Initialize(config map[string]interface{}) error
}

type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type PluginInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	ToolNames   []string `json:"tool_names"`
}
