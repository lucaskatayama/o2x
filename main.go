package main

import (
	"github.com/lucaskatayama/oauth2-cli/cmd"
	_ "github.com/lucaskatayama/oauth2-cli/cmd/flows"
)

func main() {
	cmd.Execute()
}
