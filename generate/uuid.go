package generate

import (
	"github.com/google/uuid"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var subCommandUUID = &cli.Command{
	Name:   "uuid",
	Action: UUID,
}

func UUID(c *cli.Context) error {
	newUUID := uuid.New()
	pterm.Success.Println("Generated UUID: " + pterm.LightBlue(newUUID.String()))
	return nil
}
