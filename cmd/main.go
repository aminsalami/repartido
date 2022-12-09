package main

import (
	"github.com/aminsalami/repartido/internal/agent"
)

func main() {
	xAgent := agent.NewDefaultAgent()
	xAgent.Start()
}
