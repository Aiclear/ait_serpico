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

func (a *App) OrganizeFolder(path, organizeBy string, customScript ...string) {
	switch organizeBy {
	case "File Type":
		OrganizeByFile(path)
	case "Year":
		OrganizebyYear(path)
	case "Month":
		OrganizebyMonth(path)
	case "Custom Script":
		if len(customScript) > 0 {
			OrganizeByCustomScript(path, customScript[0])
		}
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

func ExecuteScript(script string, fileInfo *FileInfo) (*OrganizeResult, error) {
	vm := goja.New()

	vm.Set("file", fileInfo)

	_, err := vm.RunString(script)
	if err != nil {
		return nil, fmt.Errorf("脚本执行错误: %v", err)
	}

	resultValue := vm.Get("result")
	if resultValue == nil || goja.IsUndefined(resultValue) || goja.IsNull(resultValue) {
		return &OrganizeResult{
			ShouldOrganize: false,
		}, nil
	}

	var result OrganizeResult
	err = vm.ExportTo(resultValue, &result)
	if err != nil {
		return nil, fmt.Errorf("结果解析错误: %v", err)
	}

	return &result, nil
}

func OrganizeByCustomScript(path string, script string) {
	files, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err)
		return
	}

	for _, f := range fileinfo {
		if f.IsDir() {
			continue
		}

		oldPath := filepath.Join(path, f.Name())
		fileInfo := getFileInfo(oldPath, f)

		result, err := ExecuteScript(script, fileInfo)
		if err != nil {
			fmt.Printf("执行脚本错误 [%s]: %v\n", f.Name(), err)
			continue
		}

		if !result.ShouldOrganize {
			fmt.Printf("跳过文件 [%s]\n", f.Name())
			continue
		}

		targetDir := filepath.Join(path, result.TargetDirectory)
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			fmt.Printf("创建目标目录错误 [%s]: %v\n", targetDir, err)
			continue
		}

		newFileName := f.Name()
		if result.NewFileName != "" {
			newFileName = result.NewFileName
		}

		newPath := filepath.Join(targetDir, newFileName)

		err = os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Printf("移动文件错误 [%s -> %s]: %v\n", oldPath, newPath, err)
			continue
		}

		fmt.Printf("移动成功 [%s -> %s]\n", oldPath, newPath)
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
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err) //if directory is not read properly print error message
		return
	}
	for _, f := range fileinfo {

		if !f.IsDir() {
			ext := filepath.Ext(f.Name())
			fmt.Println(f.Name())
			if des, ok := file_ext[ext]; ok {
				// p := path+
				destDir := filepath.Join(path, des)
				os.Mkdir(destDir, 0755)

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
	}

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
