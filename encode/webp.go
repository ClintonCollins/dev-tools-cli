package encode

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/h2non/filetype/types"
	"github.com/pterm/pterm"
	"github.com/rs/zerolog/log"

	"github.com/h2non/filetype"
	"golang.org/x/sync/errgroup"
)

type fileType string

const (
	FileTypeJpeg    fileType = "jpeg"
	FileTypePng     fileType = "png"
	FileTypeGif     fileType = "gif"
	FileTypeUnknown fileType = "unknown"
)

type WebPHandler struct {
	InputDirectoryInfo  directoryInfo
	OutputDirectoryInfo directoryInfo
	AbsoluteInputPath   string
	AbsoluteOutputPath  string
	JpegsEnabled        bool
	PngsEnabled         bool
	GifsEnabled         bool
	Lossless            bool
	Quality             int
}

type directoryInfo struct {
	Path                string
	NumberOfDirectories int64
	NumberOfFiles       int64
	SubDirectories      []string
	TotalSize           int64
	JpegCount           int64
	PngCount            int64
	GifCount            int64
	KnownFiles          []fileInfo
	UnknownFiles        []fileInfo
}

type fileInfo struct {
	inputPath  string
	outputPath string
	fileType   fileType
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

	inputDirectoryInfo, err := wpHandler.getDirectoryInfo(absoluteInputPath)
	if err != nil {
		log.Error().Err(err).Msg("Error getting input directory info")
		return err
	}
	outputDirectoryInfo, err := wpHandler.getDirectoryInfo(absoluteOutputPath)
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

func getFileTypeFromFilePath(path string) (types.Type, error) {
	file, errOpen := os.Open(path)
	if errOpen != nil {
		log.Error().Err(errOpen).Msg("Error opening file")
		return types.Unknown, errOpen
	}
	defer func() {
		_ = file.Close()
	}()
	fileHeader := make([]byte, 261)
	_, errRead := file.Read(fileHeader)
	if errRead != nil {
		log.Error().Err(errRead).Msg("Error reading file header")
		return types.Unknown, errRead
	}
	fType, errType := filetype.Match(fileHeader)
	if errType != nil {
		log.Error().Err(errType).Msg("Error matching file type")
		return types.Unknown, errType
	}
	return fType, nil
}

func (w *WebPHandler) getTrunkedOutputPath(filePath string, isDir bool) string {
	if isDir {
		trimmedDir := strings.TrimPrefix(filePath, w.AbsoluteInputPath)
		return filepath.Join(w.AbsoluteOutputPath, trimmedDir)
	}
	dir := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)
	trunkedPath := strings.TrimPrefix(dir, w.AbsoluteInputPath)
	fileNameNoExtension := trimFileExtension(fileName)
	return filepath.Join(w.AbsoluteOutputPath, trunkedPath, fileNameNoExtension+".webp")
}

func (w *WebPHandler) createOutputDirectoriesFromInputSubDirectories() error {
	for _, subDir := range w.InputDirectoryInfo.SubDirectories {
		outputSubDir := w.getTrunkedOutputPath(subDir, true)
		err := os.MkdirAll(outputSubDir, 755)
		if err != nil {
			log.Error().Err(err).Msg("Error creating output sub directory")
			return err
		}
	}
	return nil
}

func (w *WebPHandler) getDirectoryInfo(directory string) (directoryInfo, error) {
	dInfo := directoryInfo{
		Path: directory,
	}
	err := filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return nil
		}
		if f.IsDir() {
			dInfo.NumberOfDirectories++
			dInfo.SubDirectories = append(dInfo.SubDirectories, path)
			return nil
		}
		fInfo := fileInfo{
			inputPath:  path,
			outputPath: w.getTrunkedOutputPath(path, false),
		}
		fileType, errFileType := getFileTypeFromFilePath(path)
		if errFileType != nil {
			log.Error().Err(errFileType).Msg("Error getting file type")
			return nil
		}
		switch fileType.MIME.Value {
		case "image/jpeg":
			dInfo.JpegCount++
			fInfo.fileType = FileTypeJpeg
			dInfo.KnownFiles = append(dInfo.KnownFiles, fInfo)
		case "image/gif":
			dInfo.GifCount++
			fInfo.fileType = FileTypeGif
			dInfo.KnownFiles = append(dInfo.KnownFiles, fInfo)
		case "image/png":
			dInfo.PngCount++
			fInfo.fileType = FileTypePng
			dInfo.KnownFiles = append(dInfo.KnownFiles, fInfo)
		default:
			fInfo.fileType = FileTypeUnknown
			dInfo.UnknownFiles = append(dInfo.UnknownFiles, fInfo)
		}
		dInfo.TotalSize += f.Size()
		dInfo.NumberOfFiles++
		return nil
	})
	return dInfo, err
}

func trimFileExtension(name string) string {
	fileNameNoExtension := name
	extensionLength := len(name) - len(filepath.Ext(name))
	if extensionLength < len(name) && extensionLength > 0 {
		fileNameNoExtension = name[:extensionLength]
	}
	return fileNameNoExtension
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
	for _, file := range w.InputDirectoryInfo.KnownFiles {
		file := file
		if w.JpegsEnabled && file.fileType == FileTypeJpeg {
			wg.Go(func() error {
				errEncode := encodeFile(file.inputPath, file.outputPath, w.Quality, w.Lossless)
				progressBar.Increment()
				return errEncode
			})
		}
		if w.PngsEnabled && file.fileType == FileTypePng {
			wg.Go(func() error {
				errEncode := encodeFile(file.inputPath, file.outputPath, w.Quality, w.Lossless)
				progressBar.Increment()
				return errEncode
			})
		}
		if w.GifsEnabled && file.fileType == FileTypeGif {
			wg.Go(func() error {
				errEncode := encodeFile(file.inputPath, file.outputPath, w.Quality, w.Lossless)
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

	updatedOutputDirInfo, errOutputDirInfo := w.getDirectoryInfo(w.AbsoluteOutputPath)
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
