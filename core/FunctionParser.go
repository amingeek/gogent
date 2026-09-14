package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gogent/models"
)

var (
	functionsBlockRe  = regexp.MustCompile(`(?s)<functions\b[^>]*>(.*?)</functions>`)
	functionTagRe     = regexp.MustCompile(`(?s)<function\b[^>]*name\s*=\s*["']([^"']+)["'][^>]*>(.*?)</function>`)
	parametersBlockRe = regexp.MustCompile(`(?s)<parameters\b[^>]*>(.*?)</parameters>`)
)

// ParseFunctionCalls extracts tool calls from model content that is written as
// XML-like blocks, e.g.:
//
//	<functions>
//	  <function name="write_file">
//	    <parameters>
//	      <path>index.html</path>
//	      <content>...</content>
//	    </parameters>
//	  </function>
//	</functions>
//
// It converts those blocks into standard models.ToolCall entries so the agent
// loop can execute them exactly like native tool_calls.
func ParseFunctionCalls(content string) []models.ToolCall {
	block := functionsBlockRe.FindStringSubmatch(content)
	if len(block) < 2 {
		return nil
	}
	inner := block[1]

	var calls []models.ToolCall
	for _, m := range functionTagRe.FindAllStringSubmatch(inner, -1) {
		if len(m) < 3 {
			continue
		}
		name := strings.TrimSpace(m[1])
		args := parseParams(m[2])
		argJSON, err := json.Marshal(args)
		if err != nil {
			argJSON = []byte("{}")
		}

		call := models.ToolCall{
			ID:   fmt.Sprintf("call_%d", len(calls)+1),
			Type: "function",
		}
		call.Function.Name = name
		call.Function.Arguments = string(argJSON)
		calls = append(calls, call)
	}
	return calls
}

// parseParams reads <key>value</key> pairs from the text body of a function
// block. Values are kept as strings; tools must tolerate string numbers/booleans.
func parseParams(body string) map[string]interface{} {
	args := map[string]interface{}{}

	if m := parametersBlockRe.FindStringSubmatch(body); len(m) == 2 {
		body = m[1]
	}

	region := body
	for {
		region = strings.TrimLeft(region, " \t\r\n")
		if region == "" {
			break
		}
		start := strings.IndexByte(region, '<')
		if start < 0 {
			break
		}
		afterOpen := region[start+1:]
		gt := strings.IndexByte(afterOpen, '>')
		if gt < 0 {
			break
		}
		rawTag := afterOpen[:gt]
		if rawTag == "" || strings.HasPrefix(rawTag, "/") {
			break
		}
		key := rawTag
		if sp := strings.IndexAny(key, " \t\r\n"); sp >= 0 {
			key = key[:sp]
		}
		valueStart := start + 1 + gt + 1
		closeTag := "</" + key + ">"
		rest := region[valueStart:]
		end := strings.Index(rest, closeTag)
		if end < 0 {
			args[key] = strings.TrimSpace(rest)
			break
		}
		args[key] = strings.TrimSpace(rest[:end])
		region = rest[end+len(closeTag):]
	}
	return args
}

// CleanToolContent strips any <functions>...</functions> block from model
// output and returns the remaining visible text.
func CleanToolContent(content string) string {
	idx := strings.Index(content, "<functions")
	if idx == -1 {
		return strings.TrimSpace(content)
	}
	end := idx
	if closeIdx := strings.LastIndex(content, "</functions>"); closeIdx != -1 {
		end = closeIdx + len("</functions>")
	}
	clean := strings.TrimSpace(content[:idx] + content[end:])
	clean = strings.Trim(clean, "`\r\n")
	return strings.TrimSpace(clean)
}
