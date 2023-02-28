package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/urfave/cli/v2"

	"DevToolsCLI/edit"
	"DevToolsCLI/encode"
	"DevToolsCLI/generate"
	"DevToolsCLI/logging"
)

func main() {
	logging.SetGlobal()
	handleCLI()
}

func handleCLI() {
	tool := &cli.App{
		Name: "tools",
		Commands: []*cli.Command{
			edit.Command(),
			generate.Command(),
			encode.Command(),
		},
	}
	err := tool.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("Error running CLI tool")
	}
}
