package main

import (
	"gogent/core"
	_ "gogent/plugins/builtin"
	_ "gogent/plugins/read"
	_ "gogent/plugins/write"
)

func main() {
	core.AgentLoop()
}
