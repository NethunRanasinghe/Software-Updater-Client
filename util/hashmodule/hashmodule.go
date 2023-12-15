package hashmodule

import (
	"crypto/sha256"
	"io"
	"log"
	"os"
	"fmt"
	"bytes"
	"bufio"
)

func GenerateHashes(fpath string) []byte{
	file, err := os.Open(fpath)
	if err != nil{
		log.Fatal(err)
	}

	defer file.Close()

	hash := sha256.New()
	if _,err := io.Copy(hash, file); err != nil{
		log.Fatal(err)
	}

	return hash.Sum(nil)
}

func CheckHashes(localpath string, remotepath string) (bool, []string){
	var localContent []string
	var remoteContent []string
	var changedContent []string
	localMap := make(map[string]bool)

	// Read files and get there differences
	localFile,err := os.OpenFile(localpath, os.O_RDONLY, os.ModePerm)
	if(err != nil){
		log.Fatal(err)
	}
	defer localFile.Close()

	remoteFile,err := os.OpenFile(remotepath, os.O_RDONLY, os.ModePerm)
	if(err != nil){
		log.Fatal(err)
	}
	defer remoteFile.Close()

	localHash := GenerateHashes(localpath)
	remoteHash := GenerateHashes(remotepath)

	if(!bytes.Equal(localHash,remoteHash)){
		fmt.Println("# Updates Detected !, Verifying...")
	}else{
		fmt.Println("# No Updates !")
		return false, changedContent
	}

	localScanner := bufio.NewScanner(localFile)
	remoteScanner := bufio.NewScanner(remoteFile)

	// Fill local content slice
	for localScanner.Scan(){
		localContent = append(localContent, localScanner.Text())
	}

	if localError := localScanner.Err(); localError != nil{
		log.Fatal(localError)
	}

	// Fill remote content slice
	for remoteScanner.Scan(){
		remoteContent = append(remoteContent, remoteScanner.Text())
	}

	if remoteError := remoteScanner.Err(); remoteError != nil{
		log.Fatal(remoteError)
	}

	for _, value := range localContent{
		localMap[value] = true
	}

	for _, value := range remoteContent{
		if !localMap[value]{
			changedContent = append(changedContent, value)
		}
	}

	// Print files containing changes
	fmt.Print("# Changed/New Files...\n\n")
	for _,val := range changedContent{
		fmt.Println(val)
	}

	return true, changedContent
}