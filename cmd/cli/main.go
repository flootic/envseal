package main

import (
	"github.com/xfrr/envseal/internal/cli/commands"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	_ = commands.Execute()
}
