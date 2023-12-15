package utilitymodule

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"io"
	"path/filepath"
	"archive/zip"
	"softwareupdator/packages/util/remotemodule"
	"strings"
)

// Walkthrough a directory
func WalkDirectory(dirpath string) []string{
	var dirfiles []string

	fileSystem := os.DirFS(dirpath)
	
	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		
		// Check if the path is a file or a directory
		filePath := filepath.Join(dirpath,path)
		isFileCheck := CheckFileOrDirectory(filePath)

		if(isFileCheck){
			dirfiles = append(dirfiles, path)
		}

		return nil
	})

	return dirfiles
}

// Get Directory Name
func GetDirName(path string) string{
	dirPathSplit := strings.Split(path, "\\")
	dirName := dirPathSplit[len(dirPathSplit) - 1]

	return dirName
}

// Check whether the path is a directory or a file
func CheckFileOrDirectory(path string) bool{
	fileInfo,err := os.Stat(path)

	if(err != nil){
		log.Fatal(err)
	}

	if(fileInfo.IsDir()){
		return false
	}else{
		return true
	}
}

func ClearTempDirectory(){
	allFiles, err := os.ReadDir("temp")
	if err != nil{
		log.Fatal(err)
	}

	for _, file := range allFiles{	
		if file.Name() != "README"{

			pathToFile := filepath.Join("temp",file.Name())

			if CheckFileOrDirectory(pathToFile){
				err := os.Remove(pathToFile)
				if err != nil{
					log.Fatal(err)
				}
			}else{
				err := os.RemoveAll(pathToFile)
				if err != nil{
					log.Fatal(err)
				}
			}
		}
	}
}

// https://stackoverflow.com/questions/20357223/easy-way-to-unzip-file
func Unzip(src, dest string) error {
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer func() {
        if err := r.Close(); err != nil {
            panic(err)
        }
    }()

    os.MkdirAll(dest, 0755)

    // Closure to address file descriptors issue with all the deferred .Close() methods
    extractAndWriteFile := func(f *zip.File) error {
        rc, err := f.Open()
        if err != nil {
            return err
        }
        defer func() {
            if err := rc.Close(); err != nil {
                panic(err)
            }
        }()

        path := filepath.Join(dest, f.Name)

        // Check for ZipSlip (Directory traversal)
        if !strings.HasPrefix(path, filepath.Clean(dest) + string(os.PathSeparator)) {
            return fmt.Errorf("illegal file path: %s", path)
        }

        if f.FileInfo().IsDir() {
            os.MkdirAll(path, f.Mode())
        } else {
            os.MkdirAll(filepath.Dir(path), f.Mode())
            f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return err
            }
            defer func() {
                if err := f.Close(); err != nil {
                    panic(err)
                }
            }()

            _, err = io.Copy(f, rc)
            if err != nil {
                return err
            }
        }
        return nil
    }

    for _, f := range r.File {
        err := extractAndWriteFile(f)
        if err != nil {
            return err
        }
    }

    return nil
}

func getChangedValue(contentValue string) string{
	changedContent := strings.Split(contentValue, " : ")
	return changedContent[0]
}

func performSoftwareUpdate(changedContent []string, applicationDir string){
	// Create backup file
	backupFolderPath := filepath.Join("temp", "bckp")
	err := os.Mkdir(backupFolderPath, 666)
	if err != nil{
		log.Fatal(err)
	}

	// Move existing changed files to the backup folder
	for _, value := range changedContent {
		changedItem := getChangedValue(value)
		oldFilePath := filepath.Join(applicationDir, changedItem)
		newFilePath := filepath.Join(backupFolderPath, changedItem)

		if _, err := os.Stat(oldFilePath); err == nil {
			// Ensure the directory structure exists in the destination
			if err := os.MkdirAll(filepath.Dir(newFilePath), 0755); err != nil {
				log.Fatal(err)
			}

			// Move the file to the destination
			err := os.Rename(oldFilePath, newFilePath)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Unzip zip file
	zipFileName := fmt.Sprintf("%v_update.zip", GetDirName(applicationDir))
	zipFilePath := filepath.Join("temp", zipFileName)
	zipExtractPath := filepath.Join("temp", "TestApplication")
	err2 := os.Mkdir(zipExtractPath, 666)
	if err2 != nil{
		log.Fatal(err2)
	}

	err3 := Unzip(zipFilePath,zipExtractPath)
	if err3 != nil{
		log.Fatal(err3)
	}

	// Replace Changed or add New Files
	for _, value := range changedContent{
		changedItem := getChangedValue(value)
		newOrChangedPath := filepath.Join(zipExtractPath, changedItem)
		applicationPath := filepath.Join(applicationDir, changedItem)

		err := os.Rename(newOrChangedPath, applicationPath)
		if err != nil{
			performRestore(changedContent, applicationDir)
			log.Fatal(err)
		}
	}
}

func performRestore(changedContent []string, applicationDir string){

	for _, value := range changedContent{
		changedItem := getChangedValue(value)

		backupFilePath := filepath.Join("temp", "bckp", changedItem)
		applicationFilePath := filepath.Join(applicationDir, changedItem)

		// If New File : Delete
		// If Updated File : Replace

		if _, err := os.Stat(backupFilePath); err == nil {
			err1 := os.Remove(applicationFilePath)
			if err1 != nil{
				log.Fatal(err1)
			}

			err2 := os.Rename(backupFilePath, applicationFilePath)
			if err2 != nil{
				log.Fatal(err2)
			}

		}else{
			err := os.Remove(applicationFilePath)
			if err != nil{
				log.Fatal(err)
			}
		}

	}
}

func PerformUpdate(projectName string, changedContent []string, applicationDir string){
	var updateContinue string
	var restoreCheck string

	// Get Both zip and hash file
	fmt.Println("\n# Downloading Update...")
	remotemodule.GetRemoteFiles(projectName, true)

	// Perform Update
	fmt.Println("\n# ATTENTION :- Before continuing confirm all the related applications are closed !")

	for {
		fmt.Print("Enter (Y/y) to continue :- ")

		_, err := fmt.Scanf("%s\n",&updateContinue)
		if err != nil{
			log.Fatal(err)
		}

		if strings.ToLower(updateContinue) == "y"{
			fmt.Println("# Performing the update...!,  Please don't close this window !")
			break
		}else{
			fmt.Println("# Invalid Input Try Again !!!\n")
		}
	}

	performSoftwareUpdate(changedContent, applicationDir)
	fmt.Println("# Update Complete...\n")

	fmt.Println("\n# ATTENTION :- Check if the application is working correctly !")
	fmt.Print("\n\nIf working as expected enter (Y/y)\nTo revert back enter (n/N)\nChoice :- ")

	_, err := fmt.Scanf("%s\n", &restoreCheck)
	if err != nil{
		log.Fatal(err)
	}

	if strings.ToLower(restoreCheck) == "n"{
		performRestore(changedContent, applicationDir)
		fmt.Println("\n\n# Restore Complete ...")
		
	}else{
		fmt.Println("\n\n# Exiting Application ...")
	}
}

// Set Log flags
func init(){
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}