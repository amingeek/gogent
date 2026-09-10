package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gogent/config"
	"gogent/models"
	"gogent/plugins"
	"io"
	"net/http"
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

	for iteration := 0; iteration < MaxToolIterations; iteration++ {
		reqBody := models.ChatRequest{
			Model:    "aion-labs/aion-2.0",
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
			return fmt.Sprintf("Error: status %s\nMessage: %s", resp.Status, string(body))
		}

		var chatResp models.ChatResponse
		err = json.Unmarshal(body, &chatResp)
		if err != nil {
			return fmt.Sprintf("Error parsing JSON: %v", err)
		}

		if len(chatResp.Choices) == 0 {
			return "No response received."
		}

		assistantMsg := chatResp.Choices[0].Message
		messages = append(messages, assistantMsg)

		if len(assistantMsg.ToolCalls) == 0 {
			return assistantMsg.Content
		}

		for _, toolCall := range assistantMsg.ToolCalls {
			toolName := toolCall.Function.Name
			toolArgsJSON := toolCall.Function.Arguments

			result, err := registry.ExecuteTool(toolName, toolArgsJSON)
			if err != nil {
				result = fmt.Sprintf("Error executing tool %s: %v", toolName, err)
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
