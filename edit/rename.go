package edit

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"DevToolsCLI/file"
)

func rename(path string) error {
	base := filepath.Base(path)
	dir := filepath.Dir(path)
	baseEscaped := strings.ReplaceAll(base, " ", "_")
	baseEscaped = strings.ToLower(baseEscaped)
	repl := regexp.MustCompile(`[^a-zA-Z0-9-._~]+`)
	baseEscaped = repl.ReplaceAllString(baseEscaped, "")
	errRename := os.Rename(path, filepath.Join(dir, baseEscaped))
	return errRename
}

func EscapeRenameFiles(c *cli.Context) error {
	targetDirectory := c.String("target")
	recursively := c.IsSet("recursive")
	targetDirectoryInfo, errDir := file.GetDirectoryInfo(targetDirectory)
	if errDir != nil {
		log.Error().Err(errDir).Msg("Failed to get directory info")
		return errDir
	}
	pterm.DefaultSection.Println("Renaming files in " + pterm.LightGreen(targetDirectory))
	errTable := pterm.DefaultTable.WithData(pterm.TableData{
		{"Target directory", targetDirectory},
		{"Recursively", fmt.Sprintf("%t", recursively)},
		{"Files", strconv.Itoa(len(targetDirectoryInfo.Files))},
		{"Directories", strconv.FormatInt(targetDirectoryInfo.NumberOfDirectories, 10)},
	}).Render()

	if errTable != nil {
		log.Error().Err(errTable).Msg("Failed to render table")
		return errTable
	}
	confirmed, errAsk := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).
		Show("Are you sure you want to rename all files in " + pterm.LightGreen(targetDirectory) + "?")
	if errAsk != nil {
		log.Error().Err(errAsk).Msg("Failed to get ask for confirmation")
		return errAsk
	}
	if !confirmed {
		return nil
	}

	total := len(targetDirectoryInfo.Files) + int(targetDirectoryInfo.NumberOfDirectories)

	progressBar, errProgress := pterm.DefaultProgressbar.WithTotal(total).Start()
	if errProgress != nil {
		log.Error().Err(errProgress).Msg("Failed to start progress bar")
		return errProgress
	}

	var dirs []string
	errWalk := filepath.WalkDir(targetDirectory, func(path string, d fs.DirEntry, err error) error {
		if path == targetDirectory {
			progressBar.Increment()
			return nil
		}
		if d.IsDir() {
			if !recursively {
				return filepath.SkipDir
			}
			dirs = append(dirs, path)
			return nil
		}
		errRename := rename(path)
		progressBar.Increment()
		return errRename
	})
	if errWalk != nil {
		log.Error().Err(errWalk).Msg("Failed to walk directory")
		return errWalk
	}
	for _, dir := range dirs {
		errRename := rename(dir)
		if errRename != nil {
			log.Error().Err(errRename).Msg("Failed to rename directory")
			return errRename
		}
		progressBar.Increment()
	}
	return nil
}
