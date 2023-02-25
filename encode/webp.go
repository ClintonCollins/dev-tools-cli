package encode

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pterm/pterm"
	"github.com/rs/zerolog/log"

	"golang.org/x/sync/errgroup"

	"DevToolsCLI/file"
)

type WebPHandler struct {
	InputDirectoryInfo  file.DirectoryInfo
	OutputDirectoryInfo file.DirectoryInfo
	AbsoluteInputPath   string
	AbsoluteOutputPath  string
	JpegsEnabled        bool
	PngsEnabled         bool
	GifsEnabled         bool
	Lossless            bool
	Quality             int
}

func WebP(inputDirectory, outputDirectory string, quality int, lossless, jpegs, pngs, gifs bool) error {
	absoluteInputPath, err := filepath.Abs(inputDirectory)
	if err != nil {
		log.Error().Err(err).Msg("Error getting absolute path of input directory")
		return err
	}
	absoluteOutputPath, err := filepath.Abs(outputDirectory)
	if err != nil {
		log.Error().Err(err).Msg("Error getting absolute path of output directory")
		return err
	}

	wpHandler := &WebPHandler{
		AbsoluteInputPath:  absoluteInputPath,
		AbsoluteOutputPath: absoluteOutputPath,
		JpegsEnabled:       jpegs,
		PngsEnabled:        pngs,
		GifsEnabled:        gifs,
		Lossless:           lossless,
		Quality:            quality,
	}

	inputDirectoryInfo, err := file.GetDirectoryInfoIO(absoluteInputPath, absoluteOutputPath, absoluteInputPath)
	if err != nil {
		log.Error().Err(err).Msg("Error getting input directory info")
		return err
	}
	outputDirectoryInfo, err := file.GetDirectoryInfoIO(absoluteOutputPath, absoluteOutputPath, absoluteOutputPath)
	if err != nil {
		log.Error().Err(err).Msg("Error getting output directory info")
		return err
	}
	wpHandler.InputDirectoryInfo = inputDirectoryInfo
	wpHandler.OutputDirectoryInfo = outputDirectoryInfo

	errOutput := wpHandler.createOutputDirectoriesFromInputSubDirectories()
	if errOutput != nil {
		log.Error().Err(errOutput).Msg("Error creating output directories")
		return errOutput
	}

	return wpHandler.Run()
}

func (w *WebPHandler) createOutputDirectoriesFromInputSubDirectories() error {
	for _, subDir := range w.InputDirectoryInfo.SubDirectories {
		outputSubDir := file.GetTrunkedOutputPath(w.AbsoluteInputPath, w.AbsoluteOutputPath, subDir, true)
		err := os.MkdirAll(outputSubDir, 755)
		if err != nil {
			log.Error().Err(err).Msg("Error creating output sub directory")
			return err
		}
	}
	return nil
}

func (w *WebPHandler) Run() error {
	// infoLogger := logging.GetInfoLogger()

	pterm.DefaultSection.Println("Currently configured encoding settings.")
	errRender := pterm.DefaultTable.WithData(pterm.TableData{
		{"Input Directory", w.InputDirectoryInfo.Path},
		{"Output Directory", w.OutputDirectoryInfo.Path},
		{"Number of Files", strconv.FormatInt(w.InputDirectoryInfo.NumberOfFiles, 10)},
		{"Number of Directories", strconv.FormatInt(w.InputDirectoryInfo.NumberOfDirectories, 10)},
		{"Total Size Before Encoding", humanize.Bytes(uint64(w.InputDirectoryInfo.TotalSize))},
		{"Jpegs Enabled", fmt.Sprintf("%t", w.JpegsEnabled)},
		{"Jpegs Found", strconv.FormatInt(w.InputDirectoryInfo.JpegCount, 10)},
		{"Pngs Enabled", fmt.Sprintf("%t", w.PngsEnabled)},
		{"Pngs Found", strconv.FormatInt(w.InputDirectoryInfo.PngCount, 10)},
		{"Gifs Enabled", fmt.Sprintf("%t", w.GifsEnabled)},
		{"Gifs Found", strconv.FormatInt(w.InputDirectoryInfo.GifCount, 10)},
		{"Lossless Enabled", fmt.Sprintf("%t", w.Lossless)},
		{"Quality", strconv.Itoa(w.Quality)},
	}).Render()
	if errRender != nil {
		log.Error().Err(errRender).Msg("Error rendering table")
		return errRender
	}

	confirmed, errConfirm := pterm.DefaultInteractiveConfirm.WithDefaultValue(true).Show("Are you sure you want to continue?")
	if errConfirm != nil {
		log.Error().Err(errConfirm).Msg("Error confirming")
		return errConfirm
	}
	if !confirmed {
		return nil
	}

	numCoresUsed := runtime.NumCPU()
	startTime := time.Now()
	wg := new(errgroup.Group)
	wg.SetLimit(numCoresUsed)

	var totalFilesToProcess int64 = 0
	if w.JpegsEnabled {
		totalFilesToProcess += w.InputDirectoryInfo.JpegCount
	}
	if w.PngsEnabled {
		totalFilesToProcess += w.InputDirectoryInfo.PngCount
	}
	if w.GifsEnabled {
		totalFilesToProcess += w.InputDirectoryInfo.GifCount
	}

	progressBar, errProgress := pterm.DefaultProgressbar.WithTotal(int(totalFilesToProcess)).Start("Encoding files to WebP")
	if errProgress != nil {
		log.Error().Err(errProgress).Msg("Error creating progress bar")
		return errProgress
	}
	for _, f := range w.InputDirectoryInfo.KnownIOFiles {
		f := f
		if w.JpegsEnabled && f.Type == file.TypeJpeg {
			wg.Go(func() error {
				errEncode := encodeFile(f.InputPath, f.OutputPath, w.Quality, w.Lossless)
				progressBar.Increment()
				return errEncode
			})
		}
		if w.PngsEnabled && f.Type == file.TypePng {
			wg.Go(func() error {
				errEncode := encodeFile(f.InputPath, f.OutputPath, w.Quality, w.Lossless)
				progressBar.Increment()
				return errEncode
			})
		}
		if w.GifsEnabled && f.Type == file.TypeGif {
			wg.Go(func() error {
				errEncode := encodeFile(f.InputPath, f.OutputPath, w.Quality, w.Lossless)
				progressBar.Increment()
				return errEncode
			})
		}
	}
	errWait := wg.Wait()
	if errWait != nil {
		log.Error().Err(errWait).Msg("Error converting files")
		return errWait
	}

	updatedOutputDirInfo, errOutputDirInfo := file.GetDirectoryInfo(w.AbsoluteOutputPath)
	if errOutputDirInfo != nil {
		log.Error().Err(errOutputDirInfo).Msg("Error getting output directory info")
		return errOutputDirInfo
	}

	outputSizeDifference := updatedOutputDirInfo.TotalSize
	if w.OutputDirectoryInfo.TotalSize > outputSizeDifference {
		outputSizeDifference = w.OutputDirectoryInfo.TotalSize - updatedOutputDirInfo.TotalSize
	}

	spaceSaved := w.InputDirectoryInfo.TotalSize - updatedOutputDirInfo.TotalSize
	if spaceSaved < 0 {
		spaceSaved = 0
	}

	totalTimeTaken := time.Since(startTime)

	pterm.Println()
	pterm.DefaultSection.Println("Encoding Summary")
	errRender = pterm.DefaultTable.WithData(pterm.TableData{
		{"Total Input File Size Before Encoding", humanize.Bytes(uint64(w.InputDirectoryInfo.TotalSize))},
		{"Total Output File Size After Encoding", humanize.Bytes(uint64(outputSizeDifference))},
		{"Total Space Saved", humanize.Bytes(uint64(spaceSaved))},
		{"Total Time Taken to Encode All Files", totalTimeTaken.String()},
	}).Render()
	return errRender
}

func encodeFile(inputPath, outputPath string, quality int, lossless bool) error {
	var args []string
	if lossless {
		args = append(args, "-lossless")
	}

	args = append(args, "-q", fmt.Sprintf("%d", quality), "-mt", inputPath, "-o",
		outputPath, "-quiet")
	cmd := exec.Command("cwebp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
