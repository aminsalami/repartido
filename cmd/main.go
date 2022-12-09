package main

import (
	"github.com/aminsalami/discache/internal/agent"
)

func main() {
	xAgent := agent.NewDefaultAgent()
	xAgent.Start()
}
