package generate

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var subCommandKey = &cli.Command{
	Name: "key",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "bytes",
			Value: 32,
		},
	},
	Action: Key,
}

func Key(c *cli.Context) error {
	length := c.Int("bytes")
	if !c.IsSet("bytes") {
		confirm, errConfirm := pterm.DefaultInteractiveConfirm.WithDefaultValue(true).Show("Use default length of 32 bytes?")
		if errConfirm != nil {
			return errConfirm
		}
		if !confirm {
			l, errLength := pterm.DefaultInteractiveTextInput.Show("Enter the length of the key in bytes")
			if errLength != nil {
				return errLength
			}
			length, errLength = strconv.Atoi(l)
			if errLength != nil {
				return errLength
			}
		}
	}

	encodingOptions := []string{"Hex", "Base64"}
	selectedEncoding, errEncoding := pterm.DefaultInteractiveSelect.WithOptions(encodingOptions).WithDefaultOption("Hex").Show("Select the encoding: ")
	if errEncoding != nil {
		return errEncoding
	}

	k := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, k)
	if err != nil {
		return err
	}
	keyString := ""
	switch selectedEncoding {
	case "Hex":
		keyString = hex.EncodeToString(k)
	case "Base64":
		keyString = base64.StdEncoding.EncodeToString(k)
	default:
		keyString = hex.EncodeToString(k)
	}
	pterm.Success.Println("Generated key: " + pterm.LightBlue(keyString))
	return nil
}
