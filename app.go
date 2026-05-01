package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type FileOperation struct {
	OriginalPath string `json:"original_path"`
	NewPath      string `json:"new_path"`
	FileName     string `json:"file_name"`
}

type HistoryRecord struct {
	ID           string          `json:"id"`
	FolderPath   string          `json:"folder_path"`
	OrganizeBy   string          `json:"organize_by"`
	Timestamp    int64           `json:"timestamp"`
	Operations   []FileOperation `json:"operations"`
	IsRolledBack bool            `json:"is_rolled_back"`
}

type CustomRule struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Condition     string `json:"condition"`
	ConditionType string `json:"condition_type"`
	TargetFolder  string `json:"target_folder"`
}

type App struct {
	ctx             context.Context
	historyFile     string
	customRulesFile string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}
	appDataDir := filepath.Join(homeDir, ".file_organizer")
	os.MkdirAll(appDataDir, 0755)
	a.historyFile = filepath.Join(appDataDir, "history.json")
	a.customRulesFile = filepath.Join(appDataDir, "custom_rules.json")
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) OrganizeFolder(path, organizeBy string) error {
	var operations []FileOperation
	var err error

	switch organizeBy {
	case "File Type":
		operations, err = a.OrganizeByFile(path)
	case "Year":
		operations, err = a.OrganizeByYear(path)
	case "Month":
		operations, err = a.OrganizeByMonth(path)
	default:
		customRules, loadErr := a.loadCustomRules()
		if loadErr == nil {
			for _, rule := range customRules {
				if rule.ID == organizeBy {
					operations, err = a.OrganizeByCustomRule(path, rule)
					break
				}
			}
		}
	}

	if err != nil {
		return err
	}

	if len(operations) > 0 {
		record := HistoryRecord{
			ID:           generateID(),
			FolderPath:   path,
			OrganizeBy:   organizeBy,
			Timestamp:    time.Now().Unix(),
			Operations:   operations,
			IsRolledBack: false,
		}
		a.saveHistoryRecord(record)
	}

	return nil
}

func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func (a *App) saveHistoryRecord(record HistoryRecord) error {
	history, err := a.loadHistory()
	if err != nil {
		history = []HistoryRecord{}
	}
	history = append([]HistoryRecord{record}, history...)

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.historyFile, data, 0644)
}

func (a *App) loadHistory() ([]HistoryRecord, error) {
	data, err := os.ReadFile(a.historyFile)
	if err != nil {
		return nil, err
	}
	var history []HistoryRecord
	err = json.Unmarshal(data, &history)
	return history, err
}

func (a *App) GetHistory() ([]HistoryRecord, error) {
	return a.loadHistory()
}

func (a *App) Rollback(historyID string) error {
	history, err := a.loadHistory()
	if err != nil {
		return err
	}

	var targetRecord *HistoryRecord
	for i := range history {
		if history[i].ID == historyID && !history[i].IsRolledBack {
			targetRecord = &history[i]
			break
		}
	}

	if targetRecord == nil {
		return fmt.Errorf("history record not found or already rolled back")
	}

	for i := len(targetRecord.Operations) - 1; i >= 0; i-- {
		op := targetRecord.Operations[i]
		_, err := os.Stat(op.NewPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("File not found, skipping: %s\n", op.NewPath)
				continue
			}
			return err
		}

		finalDest := a.getUniquePath(op.OriginalPath)

		err = os.MkdirAll(filepath.Dir(finalDest), 0755)
		if err != nil {
			return err
		}

		err = os.Rename(op.NewPath, finalDest)
		if err != nil {
			fmt.Printf("Error rolling back file %s: %v\n", op.NewPath, err)
			continue
		}

		fmt.Printf("Rolled back: %s -> %s\n", op.NewPath, finalDest)
	}

	targetRecord.IsRolledBack = true

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	os.WriteFile(a.historyFile, data, 0644)

	for _, op := range targetRecord.Operations {
		dir := filepath.Dir(op.NewPath)
		if dir != targetRecord.FolderPath {
			isEmpty, _ := isDirEmpty(dir)
			if isEmpty {
				os.Remove(dir)
			}
		}
	}

	return nil
}

func isDirEmpty(dir string) (bool, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(files) == 0, nil
}

func (a *App) getUniquePath(path string) string {
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	counter := 1

	for {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
		if ext == "" {
			path = filepath.Join(dir, fmt.Sprintf("%s_%d", base, counter))
		} else {
			path = filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, counter, ext))
		}
		counter++
	}
}

func (a *App) OrganizeByMonth(path string) ([]FileOperation, error) {
	orgf := make(map[string]int)
	var operations []FileOperation

	files, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return operations, err
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err)
		return operations, err
	}

	for _, f := range fileinfo {
		if f.IsDir() {
			continue
		}

		d := f.Sys().(*syscall.Win32FileAttributeData)
		cTime := time.Unix(0, d.CreationTime.Nanoseconds())
		des := cTime.Month().String()
		fmt.Printf("File: %s, Month Created: %s\n", f.Name(), des)

		destDir := filepath.Join(path, des)
		if _, ok := orgf[des]; !ok {
			orgf[des] = 0
			os.Mkdir(destDir, 0755)
		}

		oldPath := filepath.Join(path, f.Name())
		newPath := filepath.Join(destDir, f.Name())

		newPath = a.getUniquePath(newPath)

		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Println("Error moving file:", err)
			continue
		}

		operations = append(operations, FileOperation{
			OriginalPath: oldPath,
			NewPath:      newPath,
			FileName:     f.Name(),
		})

		fmt.Printf("Moved %s to %s\n", oldPath, newPath)
	}
	return operations, nil
}

func (a *App) OrganizeByYear(path string) ([]FileOperation, error) {
	orgf := make(map[int]int)
	var operations []FileOperation

	files, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return operations, err
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err)
		return operations, err
	}

	for _, f := range fileinfo {
		if f.IsDir() {
			continue
		}

		d := f.Sys().(*syscall.Win32FileAttributeData)
		cTime := time.Unix(0, d.CreationTime.Nanoseconds())
		fmt.Printf("File: %s, Year Created: %d\n", f.Name(), cTime.Year())

		des := strconv.Itoa(cTime.Year())
		destDir := filepath.Join(path, des)
		if _, ok := orgf[cTime.Year()]; !ok {
			orgf[cTime.Year()] = 0
			os.Mkdir(destDir, 0755)
		}

		oldPath := filepath.Join(path, f.Name())
		newPath := filepath.Join(destDir, f.Name())

		newPath = a.getUniquePath(newPath)

		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Println("Error moving file:", err)
			continue
		}

		operations = append(operations, FileOperation{
			OriginalPath: oldPath,
			NewPath:      newPath,
			FileName:     f.Name(),
		})

		fmt.Printf("Moved %s to %s\n", oldPath, newPath)
	}
	return operations, nil
}

func (a *App) OrganizeByFile(path string) ([]FileOperation, error) {
	file_ext := map[string]string{
		".txt":  "text file",
		".pdf":  "pdf",
		".jpg":  "image",
		".png":  "image",
		".jpeg": "image",
		".mp3":  "audio",
		".ppt":  "powerpoint",
		".mkv":  "video",
		".mp4":  "video",
		".zip":  "zip files",
		".csv":  "csv files",
		".xlsx": "spreadsheets",
		".msi":  "software",
		".apk":  "software",
		".exe":  "software",
	}

	var operations []FileOperation

	files, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return operations, err
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err)
		return operations, err
	}
	for _, f := range fileinfo {

		if !f.IsDir() {
			ext := filepath.Ext(f.Name())
			fmt.Println(f.Name())
			if des, ok := file_ext[ext]; ok {
				destDir := filepath.Join(path, des)
				os.Mkdir(destDir, 0755)

				oldPath := filepath.Join(path, f.Name())
				newPath := filepath.Join(destDir, f.Name())

				newPath = a.getUniquePath(newPath)

				err = os.Rename(oldPath, newPath)
				if err != nil {
					fmt.Println("Error moving file:", err)
					continue
				}

				operations = append(operations, FileOperation{
					OriginalPath: oldPath,
					NewPath:      newPath,
					FileName:     f.Name(),
				})

				fmt.Printf("Moved %s to %s\n", oldPath, newPath)
			}

		}
	}
	return operations, nil
}

func (a *App) SelectDirectory() (string, error) {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Folder",
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) OrganizeByCustomRule(path string, rule CustomRule) ([]FileOperation, error) {
	var operations []FileOperation

	files, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return operations, err
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err)
		return operations, err
	}

	for _, f := range fileinfo {
		if f.IsDir() {
			continue
		}

		if a.matchesRule(f, rule) {
			destDir := filepath.Join(path, rule.TargetFolder)
			os.Mkdir(destDir, 0755)

			oldPath := filepath.Join(path, f.Name())
			newPath := filepath.Join(destDir, f.Name())

			newPath = a.getUniquePath(newPath)

			err = os.Rename(oldPath, newPath)
			if err != nil {
				fmt.Println("Error moving file:", err)
				continue
			}

			operations = append(operations, FileOperation{
				OriginalPath: oldPath,
				NewPath:      newPath,
				FileName:     f.Name(),
			})

			fmt.Printf("Moved %s to %s\n", oldPath, newPath)
		}
	}

	return operations, nil
}

func (a *App) matchesRule(file os.FileInfo, rule CustomRule) bool {
	switch rule.ConditionType {
	case "extension":
		ext := filepath.Ext(file.Name())
		return strings.EqualFold(strings.TrimPrefix(ext, "."), rule.Condition)
	case "filename_contains":
		return strings.Contains(strings.ToLower(file.Name()), strings.ToLower(rule.Condition))
	case "filename_starts_with":
		return strings.HasPrefix(strings.ToLower(file.Name()), strings.ToLower(rule.Condition))
	case "size_larger":
		size, err := strconv.ParseInt(rule.Condition, 10, 64)
		if err != nil {
			return false
		}
		return file.Size() > size*1024
	case "size_smaller":
		size, err := strconv.ParseInt(rule.Condition, 10, 64)
		if err != nil {
			return false
		}
		return file.Size() < size*1024
	default:
		return false
	}
}

func (a *App) loadCustomRules() ([]CustomRule, error) {
	data, err := os.ReadFile(a.customRulesFile)
	if err != nil {
		return nil, err
	}
	var rules []CustomRule
	err = json.Unmarshal(data, &rules)
	return rules, err
}

func (a *App) GetCustomRules() ([]CustomRule, error) {
	return a.loadCustomRules()
}

func (a *App) SaveCustomRule(rule CustomRule) error {
	rules, err := a.loadCustomRules()
	if err != nil {
		rules = []CustomRule{}
	}

	if rule.ID == "" {
		rule.ID = generateID()
		rules = append(rules, rule)
	} else {
		for i := range rules {
			if rules[i].ID == rule.ID {
				rules[i] = rule
				break
			}
		}
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.customRulesFile, data, 0644)
}

func (a *App) DeleteCustomRule(ruleID string) error {
	rules, err := a.loadCustomRules()
	if err != nil {
		return err
	}

	newRules := []CustomRule{}
	for _, rule := range rules {
		if rule.ID != ruleID {
			newRules = append(newRules, rule)
		}
	}

	data, err := json.MarshalIndent(newRules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.customRulesFile, data, 0644)
}
