package generate

import (
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

func Key(c *cli.Context) error {
	length := c.Int("bytes")
	k := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, k)
	if err != nil {
		return err
	}
	keyString := hex.EncodeToString(k)
	pterm.Info.Println("Generated key: " + pterm.LightGreen(keyString))
	return nil
}
