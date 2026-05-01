package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// FileMove represents a file move operation
type FileMove struct {
	SourcePath string `json:"sourcePath"`
	DestPath   string `json:"destPath"`
	FileName   string `json:"fileName"`
	IsConflict bool   `json:"isConflict"`
	Selected   bool   `json:"selected"`
}

// OrganizePreview represents the preview result for organizing
type OrganizePreview struct {
	SourcePath          string     `json:"sourcePath"`
	OrganizeBy          string     `json:"organizeBy"`
	DirectoriesToCreate []string   `json:"directoriesToCreate"`
	FilesToMove         []FileMove `json:"filesToMove"`
	ConflictCount       int        `json:"conflictCount"`
	TotalFiles          int        `json:"totalFiles"`
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

// PreviewOrganizeFolder returns a preview of the organization without actually moving files
func (a *App) PreviewOrganizeFolder(path, organizeBy string) (OrganizePreview, error) {
	var preview OrganizePreview
	preview.SourcePath = path
	preview.OrganizeBy = organizeBy
	preview.DirectoriesToCreate = []string{}
	preview.FilesToMove = []FileMove{}
	preview.ConflictCount = 0
	preview.TotalFiles = 0

	switch organizeBy {
	case "File Type":
		preview = a.previewOrganizeByFileType(path)
	case "Year":
		preview = a.previewOrganizeByYear(path)
	case "Month":
		preview = a.previewOrganizeByMonth(path)
	}

	return preview, nil
}

// ExecuteOrganizeFolder executes the organization based on the preview and user selection
func (a *App) ExecuteOrganizeFolder(preview OrganizePreview) error {
	// Create directories
	for _, dir := range preview.DirectoriesToCreate {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Println("Error creating directory:", err)
				return err
			}
		}
	}

	// Move selected files
	for _, fileMove := range preview.FilesToMove {
		if fileMove.Selected {
			err := os.Rename(fileMove.SourcePath, fileMove.DestPath)
			if err != nil {
				fmt.Println("Error moving file:", err)
				continue
			}
			fmt.Printf("Moved %s to %s\n", fileMove.SourcePath, fileMove.DestPath)
		}
	}

	return nil
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

// previewOrganizeByMonth previews organizing files by month
func (a *App) previewOrganizeByMonth(path string) OrganizePreview {
	var preview OrganizePreview
	preview.SourcePath = path
	preview.OrganizeBy = "Month"
	preview.DirectoriesToCreate = []string{}
	preview.FilesToMove = []FileMove{}
	preview.ConflictCount = 0
	preview.TotalFiles = 0

	orgf := make(map[string]bool)
	files, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return preview
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err)
		return preview
	}

	for _, f := range fileinfo {
		if !f.IsDir() {
			d := f.Sys().(*syscall.Win32FileAttributeData)
			cTime := time.Unix(0, d.CreationTime.Nanoseconds())
			des := cTime.Month().String()

			destDir := filepath.Join(path, des)
			if _, ok := orgf[des]; !ok {
				orgf[des] = true
				if _, err := os.Stat(destDir); os.IsNotExist(err) {
					preview.DirectoriesToCreate = append(preview.DirectoriesToCreate, destDir)
				}
			}

			oldPath := filepath.Join(path, f.Name())
			newPath := filepath.Join(destDir, f.Name())

			// Check if file already exists at destination (conflict)
			isConflict := false
			if _, err := os.Stat(newPath); err == nil {
				isConflict = true
				preview.ConflictCount++
			}

			fileMove := FileMove{
				SourcePath: oldPath,
				DestPath:   newPath,
				FileName:   f.Name(),
				IsConflict: isConflict,
				Selected:   !isConflict, // By default, don't select conflicting files
			}

			preview.FilesToMove = append(preview.FilesToMove, fileMove)
			preview.TotalFiles++
		}
	}

	return preview
}

// previewOrganizeByYear previews organizing files by year
func (a *App) previewOrganizeByYear(path string) OrganizePreview {
	var preview OrganizePreview
	preview.SourcePath = path
	preview.OrganizeBy = "Year"
	preview.DirectoriesToCreate = []string{}
	preview.FilesToMove = []FileMove{}
	preview.ConflictCount = 0
	preview.TotalFiles = 0

	orgf := make(map[int]bool)
	files, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return preview
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err)
		return preview
	}

	for _, f := range fileinfo {
		if !f.IsDir() {
			d := f.Sys().(*syscall.Win32FileAttributeData)
			cTime := time.Unix(0, d.CreationTime.Nanoseconds())
			des := strconv.Itoa(cTime.Year())

			destDir := filepath.Join(path, des)
			if _, ok := orgf[cTime.Year()]; !ok {
				orgf[cTime.Year()] = true
				if _, err := os.Stat(destDir); os.IsNotExist(err) {
					preview.DirectoriesToCreate = append(preview.DirectoriesToCreate, destDir)
				}
			}

			oldPath := filepath.Join(path, f.Name())
			newPath := filepath.Join(destDir, f.Name())

			// Check if file already exists at destination (conflict)
			isConflict := false
			if _, err := os.Stat(newPath); err == nil {
				isConflict = true
				preview.ConflictCount++
			}

			fileMove := FileMove{
				SourcePath: oldPath,
				DestPath:   newPath,
				FileName:   f.Name(),
				IsConflict: isConflict,
				Selected:   !isConflict, // By default, don't select conflicting files
			}

			preview.FilesToMove = append(preview.FilesToMove, fileMove)
			preview.TotalFiles++
		}
	}

	return preview
}

// previewOrganizeByFileType previews organizing files by file type
func (a *App) previewOrganizeByFileType(path string) OrganizePreview {
	var preview OrganizePreview
	preview.SourcePath = path
	preview.OrganizeBy = "File Type"
	preview.DirectoriesToCreate = []string{}
	preview.FilesToMove = []FileMove{}
	preview.ConflictCount = 0
	preview.TotalFiles = 0

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

	orgf := make(map[string]bool)
	files, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return preview
	}
	defer files.Close()

	fileinfo, err := files.Readdir(-1)
	if err != nil {
		fmt.Println("error reading directory:", err)
		return preview
	}

	for _, f := range fileinfo {
		if !f.IsDir() {
			ext := filepath.Ext(f.Name())
			if des, ok := file_ext[ext]; ok {
				destDir := filepath.Join(path, des)
				if _, ok := orgf[des]; !ok {
					orgf[des] = true
					if _, err := os.Stat(destDir); os.IsNotExist(err) {
						preview.DirectoriesToCreate = append(preview.DirectoriesToCreate, destDir)
					}
				}

				oldPath := filepath.Join(path, f.Name())
				newPath := filepath.Join(destDir, f.Name())

				// Check if file already exists at destination (conflict)
				isConflict := false
				if _, err := os.Stat(newPath); err == nil {
					isConflict = true
					preview.ConflictCount++
				}

				fileMove := FileMove{
					SourcePath: oldPath,
					DestPath:   newPath,
					FileName:   f.Name(),
					IsConflict: isConflict,
					Selected:   !isConflict, // By default, don't select conflicting files
				}

				preview.FilesToMove = append(preview.FilesToMove, fileMove)
				preview.TotalFiles++
			}
		}
	}

	return preview
}
