package encode

import "github.com/urfave/cli/v2"

func Command() *cli.Command {
	return &cli.Command{
		Name: "encode",
		Subcommands: []*cli.Command{
			subCommandWebP,
		},
	}
}
