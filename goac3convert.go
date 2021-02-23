package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func heading(message string) {
	fmt.Printf("\n")
	fmt.Printf("+" + strings.Repeat("-", len(message)+2) + "+\n")
	fmt.Printf("| %v |\n", message)
	fmt.Printf("+" + strings.Repeat("-", len(message)+2) + "+")
	fmt.Printf("\n")
}

func taskError(message string) {
	fmt.Printf("\033[31m%v\033[0m\n", message)
}

func taskSuccess(message string) {
	fmt.Printf("\033[92m%v\033[0m\n", message)
}

func taskInfo(message string) {
	fmt.Printf("\033[36m%v\033[0m\n", message)
}

func taskWarning(message string) {
	fmt.Printf("\033[33m%v\033[0m\n", message)
}

func checkExec(cmd string) bool {
	_, err := exec.LookPath(cmd)
	if err != nil {
		taskError("Could not find executable: " + cmd)
		taskError(fmt.Sprint(err))
		return false
	} else {
		// taskSuccess("Found command executable: " + cmd)
		return true
	}
}

func ac3AudioCheck(fileName string) bool {
	cmdName := "ffprobe"
	if checkExec(cmdName) == true {
		cmdExec, _ := exec.LookPath(cmdName)
		cmdAudioType := exec.Cmd{
			Path: cmdExec,
			Args: []string{cmdExec, "-v", "error", "-select_streams", "a:0", "-show_entries", "stream=codec_name", "-of", "default=nokey=1:noprint_wrappers=1", fileName},
		}
		// fmt.Println("Command:", cmdAudioType.String())
		out, err := cmdAudioType.CombinedOutput()
		if err != nil {
			taskError("Unable to get audio stream type from file")
			taskError(fmt.Sprint(err))
		} else {
			outStr := strings.TrimSpace(string(out))
			if strings.Contains(outStr, "ac3") {
				taskSuccess("The audio stream is ac3")
				return true
			} else {
				taskWarning("The audio stream is " + outStr)
				return false
			}
		}
	}
	return false
}

func ac3Convert(inputFile string) {
	fileExt := filepath.Ext(inputFile)
	fileName := inputFile[:len(inputFile)-len(fileExt)]
	workingFile := fileName + "_conv" + fileExt
	outputFile := fileName + "_ac3" + fileExt
	if fileExists(outputFile) == false && !strings.Contains(inputFile, "_ac3") {
		if ac3AudioCheck(inputFile) == false {
			taskInfo("Running ac3 audio conversion...")
			taskInfo("Working file: " + workingFile)
			cmdName := "ffmpeg"
			if checkExec(cmdName) == true {
				cmdExec, _ := exec.LookPath(cmdName)
				cmdAudioConvert := exec.Cmd{
					Path:   cmdExec,
					Args:   []string{cmdExec, "-y", "-v", "quiet", "-stats", "-i", inputFile, "-c:v", "copy", "-c:a", "ac3", workingFile},
					Stdout: os.Stdout,
					Stderr: os.Stdout,
				}
				// fmt.Println("Command:", cmdAudioConvert.String())
				if err := cmdAudioConvert.Run(); err != nil {
					taskError("Unable to convert audio")
					taskError(fmt.Sprint(err))
				} else {
					taskSuccess("Conversion completed.")
					renameFile(workingFile, outputFile)
					taskSuccess("File audio converted to ac3 and saved as " + outputFile)
					delOriginal := false
					for _, arg := range os.Args {
						if strings.Contains(arg, "-d") {
							delOriginal = true
						}
					}
					if delOriginal {
						removeFile(inputFile)
					}
				}
			}
		} else {
			taskWarning("File already contains ac3 audio stream")
		}
	} else {
		taskWarning("An ac3 filename already exists")
	}
}

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func removeFile(fileName string) {
	if fileExists(fileName) == true {
		taskInfo("Deleting file: " + fileName)
		err := os.Remove(fileName)
		if err != nil {
			taskError(fmt.Sprint(err))
		}
	}
}

func renameFile(origFileName, newFileName string) {
	taskInfo("Renaming file: " + origFileName + " >>> " + newFileName)
	err := os.Rename(origFileName, newFileName)
	if err != nil {
		taskError(fmt.Sprint(err))
	}
}

func usage() {
	fmt.Println(`
USAGE:
Note: Files will be saved in original path with _ac3 added to filename

Example 1: Convert files leaving original source file			 
goac3convert /pathtofiles

Example 2: Convert files deleting original source after conversion
goac3convert /pathtofiles -d
	`)
}

func getFileList(dirPath string) []string {
	var fileList []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return fileList
}

func main() {
	heading("GO FFMPEG AC3 VIDEOFILE AUDIO CONVERSION")
	if len(os.Args) == 1 {
		taskWarning("\nNo file path specified.\n")
		usage()
		os.Exit(1)
	}
	fileTypes := []string{".mkv", ".mp4", ".avi", ".mov"}
	dirPath := os.Args[1]
	fileList := getFileList(dirPath)
	fileCount := 0
	for _, file := range fileList {
		for _, fileExt := range fileTypes {
			if strings.ToLower(filepath.Ext(file)) == fileExt {
				fileCount += 1
				taskInfo("\nAnalysing file: " + fmt.Sprint(fileCount) + "\n" + file)
				ac3Convert(file)
			}
		}
	}
	if fileCount == 0 {
		taskWarning("\nCheck directory/filepath. No files found.\n")
		usage()
	}
}
