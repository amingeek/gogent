package main

import (
	"gogent/core"
	_ "gogent/plugins/builtin"
	_ "gogent/plugins/read"
)

func main() {
	core.AgentLoop()
}
