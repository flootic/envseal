package main

import (
	"github.com/xfrr/envseal/internal/commands"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	_ = commands.Execute()
}
