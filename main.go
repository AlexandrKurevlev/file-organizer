package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type FileOrganizer struct {
	sourceDir      string
	rulesMap       map[string]string
	logFile        *os.File
	statistic      map[string]*FileStat
	processedFiles int
	totalSize      int64
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
		statistic:      make(map[string]*FileStat),
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

func (fo *FileOrganizer) moveFile(sourcePath, targetDir string) error {
	targetPath := filepath.Join(fo.sourceDir, targetDir)
	err := os.MkdirAll(targetPath, 0750)
	if err != nil {
		fo.logError(fmt.Sprintf("ошибка создания новой директории: %q: %q", targetPath, err))
		return fmt.Errorf("ошибка создания новой директории: %q: %q", targetPath, err)
	}

	filename := filepath.Base(sourcePath)
	_, err = os.Stat(filepath.Join(targetPath, filename))
	if err == nil {
		filename = filename[:len(filename)-len(filepath.Ext(filename))] + "_" + time.Now().Format("2006-01-02_15-04-05") + filepath.Ext(filename)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		fo.logError(fmt.Sprintf("ошибка проверки существования файла: %q: %q", targetPath, err))
		return fmt.Errorf("ошибка проверки существования файла: %q: %q", targetPath, err)
	}

	targetPath = filepath.Join(targetPath, filename)
	err = os.Rename(sourcePath, targetPath)
	if err != nil {
		fo.logError(fmt.Sprintf("ошибка при перемещении файла из %q в %q: %q", sourcePath, targetPath, err))
		return fmt.Errorf("ошибка при перемещении файла из %q в %q: %q", sourcePath, targetPath, err)
	}
	fo.logSuccess(fmt.Sprintf("файл успешно перемешен из %q в %q", sourcePath, targetPath))
	return nil
}

func (fo *FileOrganizer) Organize() error {
	err := fo.initLog()
	if err != nil {
		return err
	}
	defer fo.Close()

	err = filepath.WalkDir(fo.sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Dir(path) != fo.sourceDir {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if folder, present := fo.rulesMap[ext]; present {
			err = fo.moveFile(path, folder)
			if err != nil {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}

			fo.totalSize += info.Size()
			fo.processedFiles++

			if _, present := fo.statistic[folder]; !present {
				fo.statistic[folder] = &FileStat{}
			}
			fo.statistic[folder].totalSize += info.Size()
			fo.statistic[folder].files++
		}
		return nil
	})

	return err
}

func (fo *FileOrganizer) generateReport() string {
	var res strings.Builder
	res.WriteString("=== Отчет о перемещении файлов ===")
	res.WriteString("\n\n")
	res.WriteString("Всего обработано файлов: " + strconv.Itoa(fo.processedFiles) + "\n")
	res.WriteString("Общий размер: " + convertBytes(fo.totalSize))
	res.WriteString("\n\n")
	res.WriteString("Статистика по категориям:")
	for folder := range fo.statistic {
		res.WriteString("\n\n" + folder + ":\n")
		res.WriteString(fo.statistic[folder].String())
	}
	return res.String()
}

type FileStat struct {
	files     int
	totalSize int64
}

func (fs *FileStat) String() string {
	res := "Файлов: " + strconv.Itoa(fs.files) + ", Размер: " + convertBytes(fs.totalSize)
	return res
}

func convertBytes(bytes int64) string {
	if bytes < 1000000 {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1000.0)
	}

	return fmt.Sprintf("%.2f MB", float64(bytes)/1000000.0)
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
	fmt.Println("=== Файловый органайзер ===")
	fmt.Print("Введите путь к директории для организации (Enter для текущей директории): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	sourcePath := strings.TrimSpace(input)

	if len(sourcePath) == 0 {
		sourcePath, _ = os.Getwd()
	}

	fo, err := NewFileOrganizer(sourcePath)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	fmt.Println("Начинаем организацию файлов...")
	err = fo.Organize()
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	fmt.Println(fo.generateReport())

	fmt.Println("Организация завершена! Подробности в файле organizer.log")
}
