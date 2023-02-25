package file

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"github.com/rs/zerolog/log"
)

type fileType string

const (
	TypeJpeg    fileType = "jpeg"
	TypePng     fileType = "png"
	TypeGif     fileType = "gif"
	TypeUnknown fileType = "unknown"
)

type DirectoryInfo struct {
	Path                string
	NumberOfDirectories int64
	NumberOfFiles       int64
	SubDirectories      []string
	TotalSize           int64
	JpegCount           int64
	PngCount            int64
	GifCount            int64
	KnownIOFiles        []InputOutputInfo
	UnknownIOFiles      []InputOutputInfo
	Files               []Info
}

type Info struct {
	Path string
	Type fileType
}

type InputOutputInfo struct {
	InputPath  string
	OutputPath string
	Type       fileType
}

func TrimFileExtension(name string) string {
	fileNameNoExtension := name
	extensionLength := len(name) - len(filepath.Ext(name))
	if extensionLength < len(name) && extensionLength > 0 {
		fileNameNoExtension = name[:extensionLength]
	}
	return fileNameNoExtension
}

func GetTrunkedOutputPath(absoluteInputPath, absoluteOutputPath, filePath string, isDir bool) string {
	if isDir {
		trimmedDir := strings.TrimPrefix(filePath, absoluteInputPath)
		return filepath.Join(absoluteOutputPath, trimmedDir)
	}
	dir := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)
	trunkedPath := strings.TrimPrefix(dir, absoluteInputPath)
	fileNameNoExtension := TrimFileExtension(fileName)
	return filepath.Join(absoluteOutputPath, trunkedPath, fileNameNoExtension+".webp")
}

func GetFileTypeFromFilePath(path string) (types.Type, error) {
	f, errOpen := os.Open(path)
	if errOpen != nil {
		log.Error().Err(errOpen).Msg("Error opening file")
		return types.Unknown, errOpen
	}
	defer func() {
		_ = f.Close()
	}()
	fileHeader := make([]byte, 261)
	_, errRead := f.Read(fileHeader)
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

func GetDirectoryInfo(directory string) (DirectoryInfo, error) {
	dInfo := DirectoryInfo{
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
		fInfo := Info{
			Path: path,
		}
		fileType, errFileType := GetFileTypeFromFilePath(path)
		if errFileType != nil {
			log.Error().Err(errFileType).Msg("Error getting file type")
			return nil
		}
		switch fileType.MIME.Value {
		case "image/jpeg":
			dInfo.JpegCount++
			fInfo.Type = TypeJpeg
			dInfo.Files = append(dInfo.Files, fInfo)
		case "image/gif":
			dInfo.GifCount++
			fInfo.Type = TypeGif
			dInfo.Files = append(dInfo.Files, fInfo)
		case "image/png":
			dInfo.PngCount++
			fInfo.Type = TypePng
			dInfo.Files = append(dInfo.Files, fInfo)
		default:
			fInfo.Type = TypeUnknown
			dInfo.Files = append(dInfo.Files, fInfo)
		}
		dInfo.TotalSize += f.Size()
		dInfo.NumberOfFiles++
		return nil
	})
	return dInfo, err
}

func GetDirectoryInfoIO(absoluteInputPath, absoluteOutputPath, directory string) (DirectoryInfo, error) {
	dInfo := DirectoryInfo{
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
		fInfo := InputOutputInfo{
			InputPath:  path,
			OutputPath: GetTrunkedOutputPath(absoluteInputPath, absoluteOutputPath, path, false),
		}
		fileType, errFileType := GetFileTypeFromFilePath(path)
		if errFileType != nil {
			log.Error().Err(errFileType).Msg("Error getting file type")
			return nil
		}
		switch fileType.MIME.Value {
		case "image/jpeg":
			dInfo.JpegCount++
			fInfo.Type = TypeJpeg
			dInfo.KnownIOFiles = append(dInfo.KnownIOFiles, fInfo)
		case "image/gif":
			dInfo.GifCount++
			fInfo.Type = TypeGif
			dInfo.KnownIOFiles = append(dInfo.KnownIOFiles, fInfo)
		case "image/png":
			dInfo.PngCount++
			fInfo.Type = TypePng
			dInfo.KnownIOFiles = append(dInfo.KnownIOFiles, fInfo)
		default:
			fInfo.Type = TypeUnknown
			dInfo.UnknownIOFiles = append(dInfo.UnknownIOFiles, fInfo)
		}
		dInfo.TotalSize += f.Size()
		dInfo.NumberOfFiles++
		return nil
	})
	return dInfo, err
}
