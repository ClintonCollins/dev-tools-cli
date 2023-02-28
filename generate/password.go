package generate

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
	mathrand "math/rand"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var subCommandPassword = &cli.Command{
	Name:   "password",
	Action: Password,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    "length",
			Aliases: []string{"l"},
			Value:   32,
		},
		&cli.BoolFlag{
			Name:    "numbers",
			Aliases: []string{"n"},
			Value:   true,
		},
		&cli.BoolFlag{
			Name:    "symbols",
			Aliases: []string{"s"},
			Value:   true,
		},
		&cli.IntFlag{
			Name:    "minNumbers",
			Aliases: []string{"minN"},
			Value:   0,
		},
		&cli.IntFlag{
			Name:    "minSymbols",
			Aliases: []string{"minS"},
			Value:   0,
		},
	},
}

func Password(c *cli.Context) error {
	var seed int64
	errBinaryRead := binary.Read(rand.Reader, binary.BigEndian, &seed)
	if errBinaryRead != nil {
		return errBinaryRead
	}
	length := c.Int("length")
	enableNumbers := c.IsSet("numbers")
	enableSymbols := c.IsSet("symbols")
	minNumbers := c.Int("minNumbers")
	minSymbols := c.Int("minSymbols")
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberSet := "0123456789"
	specialSet := "!@#$%^&*()_+"
	allSet := charSet + numberSet + specialSet

	if !c.IsSet("minNumbers") {
		maxRand := 1
		if length/4 > 0 {
			maxRand = length / 4
		}
		minN, err := rand.Int(rand.Reader, big.NewInt(int64(maxRand)))
		if err != nil {
			return err
		}
		minNumbers = int(minN.Int64())
	}
	if !c.IsSet("minSymbols") {
		maxRand := 1
		if length/4 > 0 {
			maxRand = length / 4
		}
		minS, err := rand.Int(rand.Reader, big.NewInt(int64(maxRand)))
		if err != nil {
			return err
		}
		minSymbols = int(minS.Int64())
	}

	buf := make([]byte, length)

	symbolsLeft := minSymbols
	numbersLeft := minNumbers
	for i := 0; i < length; i++ {
		if enableSymbols && symbolsLeft > 0 {
			symbolsLeft--
			randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(specialSet))))
			if err != nil {
				return err
			}
			buf[i] = specialSet[randInt.Int64()]
			continue
		}
		if enableNumbers && numbersLeft > 0 {
			numbersLeft--
			randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(numberSet))))
			if err != nil {
				return err
			}
			buf[i] = numberSet[randInt.Int64()]
			continue
		}

		randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(allSet))))
		if err != nil {
			return err
		}
		buf[i] = allSet[randInt.Int64()]
	}

	mathrand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})

	password := string(buf)

	pterm.Success.Println("Generated password: " + pterm.LightBlue(password))

	return nil
}
