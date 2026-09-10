package plugins

import (
	"encoding/json"
	"fmt"
	"sync"
)

type PluginRegistry struct {
	plugins map[string]Plugin
	tools   map[string]Tool
	mu      sync.RWMutex
}

var globalRegistry *PluginRegistry

func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]Plugin),
		tools:   make(map[string]Tool),
	}
}

func GetGlobalRegistry() *PluginRegistry {
	if globalRegistry == nil {
		globalRegistry = NewPluginRegistry()
	}
	return globalRegistry
}

func RegisterPlugin(plugin Plugin) error {
	return GetGlobalRegistry().Register(plugin)
}

func (r *PluginRegistry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[plugin.Name()]; exists {
		return fmt.Errorf("plugin %s already registered", plugin.Name())
	}
	r.plugins[plugin.Name()] = plugin
	for _, tool := range plugin.GetTools() {
		if _, exists := r.tools[tool.Name()]; exists {
			return fmt.Errorf("tool %s already exists, cannot register plugin %s", tool.Name(), plugin.Name())
		}
		r.tools[tool.Name()] = tool
	}
	return nil
}

func (r *PluginRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}
	for _, tool := range plugin.GetTools() {
		delete(r.tools, tool.Name())
	}
	delete(r.plugins, name)
	return nil
}

func (r *PluginRegistry) GetTool(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, exists := r.tools[name]
	return tool, exists
}

func (r *PluginRegistry) GetTools() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var tools []Tool
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (r *PluginRegistry) GetToolsForPrompt() []map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var tools []map[string]interface{}
	for _, tool := range r.tools {
		toolDef := map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name(),
				"description": tool.Description(),
				"parameters":  tool.Parameters(),
			},
		}
		tools = append(tools, toolDef)
	}
	return tools
}

func (r *PluginRegistry) GetPlugins() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var plugins []Plugin
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

func (r *PluginRegistry) GetPluginInfos() []PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var infos []PluginInfo
	for _, plugin := range r.plugins {
		var toolNames []string
		for _, tool := range plugin.GetTools() {
			toolNames = append(toolNames, tool.Name())
		}
		infos = append(infos, PluginInfo{
			Name:        plugin.Name(),
			Version:     plugin.Version(),
			Description: plugin.Description(),
			ToolNames:   toolNames,
		})
	}
	return infos
}

func (r *PluginRegistry) DetectTools(prompt string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var detected []string
	for _, tool := range r.tools {
		if isToolRelevant(prompt, tool) {
			detected = append(detected, tool.Name())
		}
	}
	return detected
}

func (r *PluginRegistry) ExecuteTool(name string, argsJSON string) (string, error) {
	tool, exists := r.GetTool(name)
	if !exists {
		return "", fmt.Errorf("tool %s not found", name)
	}
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %v", err)
	}
	return tool.Execute(args)
}

func isToolRelevant(prompt string, tool Tool) bool {
	name := tool.Name()
	desc := tool.Description()
	keywords := extractKeywords(name + " " + desc)
	for _, keyword := range keywords {
		if contains(prompt, keyword) {
			return true
		}
	}
	return false
}

func extractKeywords(text string) []string {
	words := splitWords(text)
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true,
		"should": true, "may": true, "might": true, "must": true,
		"that": true, "this": true, "these": true, "those": true,
		"what": true, "which": true, "who": true, "whom": true,
		"whose": true, "where": true, "when": true, "why": true, "how": true,
		"all": true, "each": true, "every": true, "both": true, "few": true,
		"more": true, "most": true, "other": true, "some": true, "such": true,
		"no": true, "nor": true, "not": true, "only": true, "own": true,
		"same": true, "so": true, "than": true, "too": true, "very": true,
		"just": true, "can": true, "it": true, "its": true, "i": true,
		"me": true, "my": true, "we": true, "our": true, "you": true,
		"your": true, "he": true, "him": true, "his": true, "she": true,
		"her": true, "they": true, "them": true, "their": true,
	}
	var keywords []string
	for _, word := range words {
		word = cleanWord(word)
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	return keywords
}

func splitWords(text string) []string {
	var words []string
	var current []rune
	for _, r := range text {
		if isLetter(r) {
			current = append(current, r)
		} else if len(current) > 0 {
			words = append(words, string(current))
			current = nil
		}
	}
	if len(current) > 0 {
		words = append(words, string(current))
	}
	return words
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func cleanWord(word string) string {
	var result []rune
	for _, r := range word {
		if isLetter(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

func contains(text, substr string) bool {
	if len(text) < len(substr) {
		return false
	}
	lowerText := toLower(text)
	lowerSubstr := toLower(substr)
	for i := 0; i <= len(lowerText)-len(lowerSubstr); i++ {
		if lowerText[i:i+len(lowerSubstr)] == lowerSubstr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	var result []rune
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
