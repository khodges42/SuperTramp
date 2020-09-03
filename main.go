package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"superTramp/config_handler"
	"time"
)



func main(){
	connConfig, err := config_handler.VerifyArgs()
	if err != nil {
		log.Fatalf("Failed while generating connection config: %s",err)
		os.Exit(1)
	}

	err = createDirsIfNotExist(connConfig)
	if err != nil {
		log.Fatalf("Can't create temp dirs: %s", err)
		os.Exit(1)
	}
	//defer os.Remove(connConfig.TempMetaDirPath)
	//defer os.Remove(connConfig.TempDirPath)
	defer os.RemoveAll(connConfig.TrampDir)

	fmt.Println("Getting Remote File")
	err = scpRemoteToLocal(connConfig, connConfig.File, connConfig.TempFilePath)
	if err != nil {
		fmt.Println("Error while getting file ", err)
		os.Exit(1)
	}
	defer os.Remove(connConfig.TempFilePath)

	err = writeMD5ToMetaFile(connConfig.TempFilePath, connConfig.TempHashFilePath)
	if err != nil {
		//config_handler.DumpStrings(connConfig)
		fmt.Println("Error while writing metafile ", err)
		os.Exit(1)
	}
	defer os.Remove(connConfig.TempHashFilePath)

	fmt.Println("Launching Editor")
	cmnd := exec.Command(connConfig.Editor, connConfig.TempFilePath) // Sometimes it wont be the first arg, maybe?
	//cmnd.Run() // and wait
	cmnd.Start()

	fmt.Println("Waiting for save")
	waitForFileChange(connConfig.TempFilePath)

	if compareRemoteMD5(connConfig) != true {
		fmt.Println("Remote file has changed, continue? y/n")
		userconfirm := yn()
		if userconfirm != true {
			fmt.Println("Aborting")
			os.Exit(1)
		}
	}

	fmt.Println("Syncing remote with local")
	scpLocalToRemote(connConfig, connConfig.File, connConfig.TempFilePath)

}

func writeMD5ToMetaFile(filepath string, hashfilepath string) error{
	filehash, err := hashFileMD5(filepath)
	if err != nil {
		fmt.Println("Hashing failed")
		fmt.Println(filehash)
		fmt.Println(err)
		return err
	}

	tempMetaFile, err := os.Create(hashfilepath)

	if err != nil {
		fmt.Println(fmt.Sprintf("Create %s failed", hashfilepath))
		return err
	}

	defer tempMetaFile.Close()

	_, err = tempMetaFile.WriteString(filehash)

	if err != nil {
		fmt.Println("Write to metafile failed")
		return err
	}
	return nil
	}

func waitForFileChange(filepath string){
	doneChan := make(chan bool)

	go func(doneChan chan bool) {
		defer func() {
			doneChan <- true
		}()

		err := watchFile(filepath)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("File has been changed")
	}(doneChan)

	<-doneChan
}

func createDirsIfNotExist(connConfig config_handler.ConnectionConfig) error {

	if _, err := os.Stat(connConfig.TempDirPath); os.IsNotExist(err) {
		os.MkdirAll(connConfig.TempDirPath, 0777)
		fmt.Println("created",connConfig.TempDirPath)
	} else {
		return err
	}

	if _, err := os.Stat(connConfig.TempMetaDirPath); os.IsNotExist(err) {
		os.MkdirAll(connConfig.TempMetaDirPath, 0777)
		fmt.Println("created",connConfig.TempMetaDirPath)
	} else {
		return err
	}

	return nil
}

func compareRemoteMD5(connConfig config_handler.ConnectionConfig) bool {
	createDirsIfNotExist(connConfig)
	file, err := ioutil.TempFile(connConfig.TempMetaDirPath, "comparison")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	scpRemoteToLocal(connConfig, connConfig.File, file.Name())

	remoteFileHash, err := hashFileMD5(file.Name())
	if err != nil{
		fmt.Println("Error while fetching remote md5: %s", err)
		return false // I know I know
	}

	localFileHash, err := ioutil.ReadFile(connConfig.TempHashFilePath)
	if err != nil {
		fmt.Println( "Error while reading local hash file: %s", err)
		return false // I know I know
	}

	return remoteFileHash == string(localFileHash)



}

func watchFile(filePath string) error {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	for {
		stat, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func scpLocalToRemote(connConfig config_handler.ConnectionConfig, remotepath string, localpath string) error{
	clientConfig, _ := auth.PrivateKey(connConfig.Username, connConfig.KeyFile, ssh.InsecureIgnoreHostKey())

	client := scp.NewClient(fmt.Sprintf("%s:%s", connConfig.Host, connConfig.Port), &clientConfig)

	err := client.Connect()
	if err != nil {
		return err
	}

	// Opens local file
	f, _ := os.Open(localpath)

	defer client.Close()
	defer f.Close()

	err = client.CopyFromFile(*f, remotepath, "0777")
	if err != nil {
		return err
	}

	return nil
}

func scpRemoteToLocal(connConfig config_handler.ConnectionConfig, remotepath string, localpath string) error{

	// Go-SCP Doesnt implement a way to copy remote to local. Its not a very good library in general
	// why doesnt go have support for scp that isnt written by some CS first year
	// and why didnt I just write this whole thing in bash to begin with
	cmnd := exec.Command("scp", fmt.Sprintf("%s@%s:%s", connConfig.Username, connConfig.Host, remotepath), localpath)
	cmnd.Run() // and wait
	//cmnd.Start()

	//clientConfig, _ := auth.PrivateKey(connConfig.Username, connConfig.KeyFile, ssh.InsecureIgnoreHostKey())

	//client := scp.NewClient(fmt.Sprintf("%s:%s", connConfig.Host, connConfig.Port), &clientConfig)

	//err := client.Connect()
	//if err != nil {
	//	return err
	//}

	// Opens local file
	//f, err := os.OpenFile(localpath, os.O_CREATE, 0777)
	//f, err := os.Create(localpath)

	//f, err := os.Open(localpath)
	//if err != nil {
	//	fmt.Println(err)
	//}

	//defer client.Close()
	//defer f.Close()

	// Copies remote file
	//err = client.CopyFile(f, remotepath, "0777")
	//if err != nil {
	//	return err
	//}

	return nil
}

func hashFileMD5(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil

}

func yn() bool {
	var s string

	fmt.Printf("(y/N): ")
	_, err := fmt.Scan(&s)
	if err != nil {
		panic(err)
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true
	}
	return false
}