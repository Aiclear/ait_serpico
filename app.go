package main

import (
	"context"
	"encoding/json"
	"github.cfm/wailmapp/wails/v2/pkg/runtimet"
	"os"
	"os"
	"strconv"
	"syscall"
	"time"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// const organizeOptions = ["Year", "Month", "File Type"];
func (a *App) OrganizeFolder(path, Organizeby string) {
	switch Organizeby {
	case "File Type":
		OrganizeByFile(path)
	case "Year":
		OrganizebyYear(path)
	case "Month":
		OrganizebyMonth(path)
	}
}

func OrganizebyMonth(path string) {
	orgf := make(map[string]int)
	files, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err)
		//if directory is not read properly print error message
		return
	}

	for _, f := range fileinfo {
		d := f.Sys().(*syscall.Win32FileAttributeData)
		cTime := time.Unix(0, d.CreationTime.Nanoseconds())
		// t := cTime.Month()
		des := cTime.Month().String()
		fmt.Printf("File: %s, Year Created: %d\n", f.Name(), des)

		destDir := filepath.Join(path, des)
		if _, ok := orgf[des]; !ok {
			orgf[des] = 0
			os.Mkdir(destDir, 0755)
		}

		oldPath := filepath.Join(path, f.Name())
		newPath := filepath.Join(destDir, f.Name())

		// Move file to new directory
		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Println("Error moving file:", err)
			continue
		}

		fmt.Printf("Moved %s to %s\n", oldPath, newPath)
	}
}

func OrganizebyYear(path string) {
	orgf := make(map[int]int)
	files, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err) //if directory is not read properly print error message
		return
	}

	for _, f := range fileinfo {
		d := f.Sys().(*syscall.Win32FileAttributeData)
		cTime := time.Unix(0, d.CreationTime.Nanoseconds())
		// t := cTime.Year()
		fmt.Printf("File: %s, Year Created: %d\n", f.Name(), cTime.Year())

		des := strconv.Itoa(cTime.Year())
		destDir := filepath.Join(path, des)
		if _, ok := orgf[cTime.Year()]; !ok {
			orgf[cTime.Year()] = 0
			os.Mkdir(destDir, 0755)
		}

		oldPath := filepath.Join(path, f.Name())
		newPath := filepath.Join(destDir, f.Name())

		// Move file to new directory
		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Println("Error moving file:", err)
			continue
		}

		fmt.Printf("Moved %s to %s\n", oldPath, newPath)
	}
}

func OrganizeByFile(path string) {
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

	files, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return
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
		
	os.WriteFile(a.historyFile, data, 0644)

		r _, op := range targetRecord.Operations {
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

func (a *App) OrganizebyMonth(path string) ([]FileOperation, error) {
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

func (a *App) OrganizebyYear(path string) ([]FileOperation, error) {
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
