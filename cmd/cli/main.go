package main

import (
	"github.com/xfrr/envseal-cli/internal/commands"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	_ = commands.Execute()
}
