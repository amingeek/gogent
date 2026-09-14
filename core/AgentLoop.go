package core

import (
	"fmt"
	"gogent/plugins"
	"gogent/utils"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

func AgentLoop() {
	fmt.Println("Welcome to gogent! : ")
	fmt.Println("Type /exit to quit.")
	fmt.Println()

	infos := plugins.GetGlobalRegistry().GetPluginInfos()
	if len(infos) > 0 {
		fmt.Printf("Loaded %d plugins:\n", len(infos))
		for _, info := range infos {
			fmt.Printf("  [%s v%s] %s (tools: %s)\n", info.Name, info.Version, info.Description, strings.Join(info.ToolNames, ", "))
		}
		fmt.Println()
	}

	for {
		fmt.Print("=> ")
		userPrompt := strings.TrimSpace(utils.ReadUserPrompt())
		if userPrompt == "/exit" {
			break
		}
		if userPrompt == "" {
			continue
		}

		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Suffix = " Thinking..."
		s.Start()

		response := CallAI(userPrompt)

		s.Stop()

		fmt.Println("Gogent:", response)
	}
	utils.Exit()
}
