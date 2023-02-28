package generate

import (
	"math/rand"
	"time"

	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	rand.NewSource(time.Now().UnixNano())
	return &cli.Command{
		Name: "generate",
		Subcommands: []*cli.Command{
			subCommandKey,
			subCommandUUID,
			subCommandPassword,
		},
	}
}
