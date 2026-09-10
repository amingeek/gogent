package core

import (
	"gogent/plugins"
)

func GetAvailableTools() []plugins.Tool {
	return plugins.GetGlobalRegistry().GetTools()
}

func DetectRelevantTools(prompt string) []string {
	return plugins.GetGlobalRegistry().DetectTools(prompt)
}

func GetPluginInfos() []plugins.PluginInfo {
	return plugins.GetGlobalRegistry().GetPluginInfos()
}
