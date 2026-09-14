package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gogent/config"
	"gogent/models"
	"gogent/plugins"
	"gogent/utils"
)

const MaxToolIterations = 10

func CallAI(prompt string) string {
	BASE_URL := config.GetApiBaseUrl()
	API_KEY := config.GetApiKey()
	registry := plugins.GetGlobalRegistry()

	messages := []models.Message{
		{Role: "user", Content: prompt},
	}

	toolDefs := registry.GetToolsForPrompt()
	var apiTools []models.Tool
	for _, td := range toolDefs {
		functionData, _ := json.Marshal(td["function"])
		var fn models.Function
		json.Unmarshal(functionData, &fn)
		apiTools = append(apiTools, models.Tool{
			Type:     "function",
			Function: fn,
		})
	}

	utils.Infof("AI: sending prompt (%d tools advertised)", len(apiTools))

	for iteration := 0; iteration < MaxToolIterations; iteration++ {
		reqBody := models.ChatRequest{
			Model:    "Qwen/Qwen2.5-Coder-7B-Instruct",
			Messages: messages,
			Tools:    apiTools,
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Sprintf("Error building JSON: %v", err)
		}

		url := BASE_URL + "/chat/completions"

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Sprintf("Error building request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+API_KEY)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Sprintf("Error sending request: %v", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Sprintf("Error reading response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			utils.Errorf("AI returned HTTP %s", resp.Status)
			return fmt.Sprintf("Error: status %s\nMessage: %s", resp.Status, string(body))
		}

		var chatResp models.ChatResponse
		if err := json.Unmarshal(body, &chatResp); err != nil {
			return fmt.Sprintf("Error parsing JSON: %v", err)
		}

		if len(chatResp.Choices) == 0 {
			return "No response received."
		}

		assistantMsg := chatResp.Choices[0].Message
		messages = append(messages, assistantMsg)

		// 1) Native OpenAI-style tool_calls
		toolCalls := assistantMsg.ToolCalls

		// 2) Fallback: model wrote the call as a <functions> block in text
		if len(toolCalls) == 0 {
			toolCalls = ParseFunctionCalls(assistantMsg.Content)
			if len(toolCalls) > 0 {
				utils.Infof("AI: found %d function call(s) in response text", len(toolCalls))
			}
		}

		if len(toolCalls) == 0 {
			return CleanToolContent(assistantMsg.Content)
		}

		for _, toolCall := range toolCalls {
			toolName := toolCall.Function.Name
			toolArgsJSON := toolCall.Function.Arguments

			utils.Infof("tool %s: args=%s", toolName, toolArgsJSON)

			result, err := registry.ExecuteTool(toolName, toolArgsJSON)
			if err != nil {
				result = fmt.Sprintf("Error executing tool %s: %v", toolName, err)
				utils.Errorf("tool %s failed: %v", toolName, err)
			} else {
				utils.Infof("tool %s: result=%.200s", toolName, result)
			}

			toolMsg := models.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: toolCall.ID,
			}
			messages = append(messages, toolMsg)
		}
	}

	return "Error: maximum tool iterations reached."
}
