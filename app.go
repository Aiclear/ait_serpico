package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/dop251/goja"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Extension   string    `json:"extension"`
	Size        int64     `json:"size"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	IsDirectory bool      `json:"isDirectory"`
	BaseName    string    `json:"baseName"`
}

type OrganizeResult struct {
	ShouldOrganize  bool   `json:"shouldOrganize"`
	TargetDirectory string `json:"targetDirectory"`
	NewFileName     string `json:"newFileName"`
}

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

type OrganizeError struct {
	Message string `json:"message"`
}

func (a *App) OrganizeFolder(path, organizeBy string) error {
	switch organizeBy {
	case "File Type":
		return OrganizeByFile(path)
	case "Year":
		return OrganizebyYear(path)
	case "Month":
		return OrganizebyMonth(path)
	default:
		return fmt.Errorf("未知的整理方式: %s", organizeBy)
	}
}

func (a *App) OrganizeFolderWithScript(path, organizeBy, customScript string) error {
	if organizeBy == "Custom Script" {
		return OrganizeByCustomScript(path, customScript)
	} else {
		return a.OrganizeFolder(path, organizeBy)
	}
}

func getFileInfo(path string, info os.FileInfo) *FileInfo {
	var createdTime, modifiedTime time.Time
	modifiedTime = info.ModTime()

	if d, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		createdTime = time.Unix(0, d.CreationTime.Nanoseconds())
		modifiedTime = time.Unix(0, d.LastWriteTime.Nanoseconds())
	}

	extension := filepath.Ext(info.Name())
	baseName := info.Name()
	if extension != "" {
		baseName = baseName[:len(baseName)-len(extension)]
	}

	return &FileInfo{
		Name:        info.Name(),
		Path:        path,
		Extension:   extension,
		Size:        info.Size(),
		Created:     createdTime,
		Modified:    modifiedTime,
		IsDirectory: info.IsDir(),
		BaseName:    baseName,
	}
}

func toScriptFileInfo(fileInfo *FileInfo) map[string]interface{} {
	return map[string]interface{}{
		"name":        fileInfo.Name,
		"path":        fileInfo.Path,
		"extension":   fileInfo.Extension,
		"size":        fileInfo.Size,
		"created":     fileInfo.Created.Unix(),
		"createdStr":  fileInfo.Created.Format("2006-01-02 15:04:05"),
		"modified":    fileInfo.Modified.Unix(),
		"modifiedStr": fileInfo.Modified.Format("2006-01-02 15:04:05"),
		"isDirectory": fileInfo.IsDirectory,
		"baseName":    fileInfo.BaseName,
	}
}

func ExecuteScript(script string, fileInfo *FileInfo) (*OrganizeResult, error) {
	vm := goja.New()

	scriptFileInfo := toScriptFileInfo(fileInfo)
	vm.Set("file", scriptFileInfo)

	fmt.Printf("执行脚本 - 文件: %s, 扩展名: %s\n", fileInfo.Name, fileInfo.Extension)

	_, err := vm.RunString(script)
	if err != nil {
		return nil, fmt.Errorf("脚本执行错误: %v", err)
	}

	resultValue := vm.Get("result")
	if resultValue == nil || goja.IsUndefined(resultValue) || goja.IsNull(resultValue) {
		fmt.Println("result 为 null 或 undefined，跳过此文件")
		return &OrganizeResult{
			ShouldOrganize: false,
		}, nil
	}

	resultMap, ok := resultValue.Export().(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("result 必须是一个对象")
	}

	result := &OrganizeResult{}

	if shouldOrganize, ok := resultMap["shouldOrganize"]; ok {
		fmt.Printf("找到 shouldOrganize: %v (类型: %T)\n", shouldOrganize, shouldOrganize)
		if b, ok := shouldOrganize.(bool); ok {
			result.ShouldOrganize = b
		}
	}

	if targetDirectory, ok := resultMap["targetDirectory"]; ok {
		fmt.Printf("找到 targetDirectory: %v (类型: %T)\n", targetDirectory, targetDirectory)
		if s, ok := targetDirectory.(string); ok {
			result.TargetDirectory = s
		}
	}

	if newFileName, ok := resultMap["newFileName"]; ok {
		if s, ok := newFileName.(string); ok {
			result.NewFileName = s
		}
	}

	fmt.Printf("解析结果: shouldOrganize=%v, targetDirectory=%s, newFileName=%s\n",
		result.ShouldOrganize, result.TargetDirectory, result.NewFileName)

	return result, nil
}

func OrganizeByCustomScript(path string, script string) error {
	files, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("无法打开目录 [%s]: %v", path, err)
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		return fmt.Errorf("无法读取目录 [%s]: %v", path, err)
	}

	var errorMessages []string

	for _, f := range fileinfo {
		if f.IsDir() {
			continue
		}

		oldPath := filepath.Join(path, f.Name())
		fileInfo := getFileInfo(oldPath, f)

		result, err := ExecuteScript(script, fileInfo)
		if err != nil {
			errorMsg := fmt.Sprintf("文件 [%s] 脚本执行错误: %v", f.Name(), err)
			fmt.Println(errorMsg)
			errorMessages = append(errorMessages, errorMsg)
			continue
		}

		if !result.ShouldOrganize {
			fmt.Printf("跳过文件 [%s]\n", f.Name())
			continue
		}

		if result.TargetDirectory == "" {
			errorMsg := fmt.Sprintf("文件 [%s] 目标目录为空", f.Name())
			fmt.Println(errorMsg)
			errorMessages = append(errorMessages, errorMsg)
			continue
		}

		targetDir := filepath.Join(path, result.TargetDirectory)
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			errorMsg := fmt.Sprintf("文件 [%s] 创建目标目录错误 [%s]: %v", f.Name(), targetDir, err)
			fmt.Println(errorMsg)
			errorMessages = append(errorMessages, errorMsg)
			continue
		}

		newFileName := f.Name()
		if result.NewFileName != "" {
			newFileName = result.NewFileName
		}

		newPath := filepath.Join(targetDir, newFileName)

		err = os.Rename(oldPath, newPath)
		if err != nil {
			errorMsg := fmt.Sprintf("文件 [%s -> %s] 移动错误: %v", oldPath, newPath, err)
			fmt.Println(errorMsg)
			errorMessages = append(errorMessages, errorMsg)
			continue
		}

		fmt.Printf("移动成功 [%s -> %s]\n", oldPath, newPath)
	}

	if len(errorMessages) > 0 {
		return fmt.Errorf("整理过程中发生 %d 个错误: %v", len(errorMessages), errorMessages)
	}

	return nil
}

func OrganizebyMonth(path string) error {
	orgf := make(map[string]int)
	files, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("无法打开目录 [%s]: %v", path, err)
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		return fmt.Errorf("无法读取目录 [%s]: %v", path, err)
	}

	for _, f := range fileinfo {
		d := f.Sys().(*syscall.Win32FileAttributeData)
		cTime := time.Unix(0, d.CreationTime.Nanoseconds())
		des := cTime.Month().String()
		fmt.Printf("File: %s, Year Created: %s\n", f.Name(), des)

		destDir := filepath.Join(path, des)
		if _, ok := orgf[des]; !ok {
			orgf[des] = 0
			os.Mkdir(destDir, 0755)
		}

		oldPath := filepath.Join(path, f.Name())
		newPath := filepath.Join(destDir, f.Name())

		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Printf("Error moving file %s: %v\n", f.Name(), err)
			continue
		}

		fmt.Printf("Moved %s to %s\n", oldPath, newPath)
	}

	return nil
}

func OrganizebyYear(path string) error {
	orgf := make(map[int]int)
	files, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("无法打开目录 [%s]: %v", path, err)
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		return fmt.Errorf("无法读取目录 [%s]: %v", path, err)
	}

	for _, f := range fileinfo {
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

		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Printf("Error moving file %s: %v\n", f.Name(), err)
			continue
		}

		fmt.Printf("Moved %s to %s\n", oldPath, newPath)
	}

	return nil
}

func OrganizeByFile(path string) error {
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
		return fmt.Errorf("无法打开目录 [%s]: %v", path, err)
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		return fmt.Errorf("无法读取目录 [%s]: %v", path, err)
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

				err = os.Rename(oldPath, newPath)
				if err != nil {
					fmt.Printf("Error moving file %s: %v\n", f.Name(), err)
					continue
				}

				fmt.Printf("Moved %s to %s\n", oldPath, newPath)
			}

		}
	}

	return nil
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
