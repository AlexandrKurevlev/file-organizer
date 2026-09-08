package main

import (
	"fmt"
	"log"
	"os"
)

type FileOrganizer struct {
	sourceDir      string
	rulesMap       map[string]string
	processedFiles int
	logFile        *os.File
}

func NewFileOrganizer(sourceDir string) (*FileOrganizer, error) {
	if len(sourceDir) == 0 {
		return nil, fmt.Errorf("строка пути пустая")
	}

	info, err := os.Stat(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("директория не найдена: %q", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("указан путь до файла вместо директории: %q", sourceDir)
	}

	return &FileOrganizer{
		sourceDir:      sourceDir,
		rulesMap:       DefaultRules,
		processedFiles: 0,
	}, nil
}

func (fo *FileOrganizer) initLog() error {
	file, err := os.OpenFile("organizer.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	fo.logFile = file
	log.SetOutput(file)
	return nil
}

func (fo *FileOrganizer) logSuccess(message string) {
	log.Println("[SUCCESS]", message)
}

func (fo *FileOrganizer) logError(message string) {
	log.Println("[ERROR]", message)
}

func (fo *FileOrganizer) Close() error {
	if fo.logFile == nil {
		return nil
	}

	return fo.logFile.Close()
}

var DefaultRules = map[string]string{
	".jpg":  "Images",
	".jpeg": "Images",
	".png":  "Images",
	".pdf":  "Documents",
	".doc":  "Documents",
	".docx": "Documents",
	".txt":  "Documents",
	".mp3":  "Music",
	".wav":  "Music",
	".mp4":  "Video",
	".avi":  "Video",
	".zip":  "Archives",
	".rar":  "Archives",
}

func main() {
	pathToDir := "./go.mod"
	_, err := NewFileOrganizer(pathToDir)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	fmt.Println("FileOrganizer: создан для директории:", pathToDir)
}
