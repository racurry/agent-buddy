package main

import (
	"os"

	"github.com/agenthubdev/agent-buddy/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
