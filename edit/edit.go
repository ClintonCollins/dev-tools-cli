package edit

import "github.com/urfave/cli/v2"

func Command() *cli.Command {
	return &cli.Command{
		Name: "edit",
		Subcommands: []*cli.Command{
			subCommandRename,
		},
	}
}
