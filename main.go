package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/urfave/cli/v2"

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
			{
				Name: "generate",
				Subcommands: []*cli.Command{
					{
						Name: "key",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:  "bytes",
								Value: 32,
							},
						},
						Action: generate.Key,
					},
				},
			},
			{
				Name: "encode",
				Subcommands: []*cli.Command{
					{
						Name: "webp",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "input",
								Required:    true,
								DefaultText: "the directory you want to encode the webp files from",
							},
							&cli.StringFlag{
								Name:     "output",
								Required: true,
								Usage:    "the directory you want to output the webp files to",
							},
							&cli.StringFlag{
								Name:     "lossless",
								Required: false,
								Usage:    "enable lossless encoding",
							},
							&cli.IntFlag{
								Name:     "quality",
								Required: false,
								Aliases:  []string{"q"},
								Usage:    "quality 0-100",
								Value:    80,
							},
							&cli.BoolFlag{
								Name:     "jpegs",
								Required: false,
								Usage:    "encode jpegs",
								Value:    false,
							},
							&cli.BoolFlag{
								Name:     "pngs",
								Required: false,
								Usage:    "encode pngs",
								Value:    false,
							},
							&cli.BoolFlag{
								Name:     "gifs",
								Required: false,
								Usage:    "encode gifs",
								Value:    false,
							},
						},
						Action: func(c *cli.Context) error {
							in := c.String("input")
							out := c.String("output")
							quality := c.Int("quality")
							lossless := c.IsSet("lossless")
							jpegs := c.Bool("jpegs")
							pngs := c.Bool("pngs")
							gifs := c.Bool("gifs")
							if !c.IsSet("jpegs") && !c.IsSet("pngs") && !c.IsSet("gifs") {
								jpegs = true
								pngs = true
								gifs = true
							}
							return encode.WebP(in, out, quality, lossless, jpegs, pngs, gifs)
						},
					},
				},
			},
		},
	}
	err := tool.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("Error running CLI tool")
	}
}
