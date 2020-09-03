package config_handler

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)


func getCliFlags() (ConnectionConfig){
	userName := flag.String("u", "", "Username")
	password:= flag.String("p", "", "Password")
	sshKey := flag.String("i", "", "SSH Key file")
	hostip := flag.String("h", "", "Host IP")
	filen := flag.String("f", "", "Path to remote file")
	port := flag.String("port", "22", "SSH Port.")
	trampDir := flag.String("trampdir", "", "Root of directory to store temp files")
	editor := flag.String("editor", os.Getenv("EDITOR"), "Editor to use, defaults to $EDITOR")

	flag.Parse()

	return ConnectionConfig{
		Username:         *userName,
		Password:         *password,
		KeyFile:          *sshKey,
		Host:             *hostip,
		File:             *filen,
		Port:             *port,
		TrampDir:         *trampDir,
		Editor:			  *editor,
		TempDirPath:      fmt.Sprintf("%s/%s", *trampDir, *hostip),
		TempMetaDirPath:  fmt.Sprintf("%s/%s/.meta", *trampDir, *hostip),
		TempHashFilePath: fmt.Sprintf("%s/%s/.meta/%s", *trampDir, *hostip, filepath.Base(*filen)),
		TempFilePath:     fmt.Sprintf("%s/%s/%s", *trampDir, *hostip, filepath.Base(*filen)),
	}
}

func DumpStrings(cfg ConnectionConfig){
	fmt.Println("Username: %s", cfg.Username)
	fmt.Println("Password: %s", cfg.Password)
	fmt.Println("KeyFile: %s", cfg.KeyFile)
	fmt.Println("Host: %s", cfg.Host)
	fmt.Println("File: %s", cfg.File)
	fmt.Println("Port: %s", cfg.Port)
	fmt.Println("TrampDir: %s", cfg.TrampDir)
	fmt.Println("Editor: %s", cfg.Editor)
	fmt.Println("TempDirPath: %s", cfg.TempDirPath)
	fmt.Println("TempMetaDirPath: %s", cfg.TempMetaDirPath)
	fmt.Println("TempHashFile: %s", cfg.TempHashFilePath)
	fmt.Println("TempFilePath: %s", cfg.TempFilePath)
}
func VerifyArgs()(ConnectionConfig, error) {
	cliArgs := getCliFlags()

	if len(cliArgs.Host) <= 0 {
		fmt.Println("Please specify a host!")
		flag.Usage()
		os.Exit(1)
	}

	if len(cliArgs.Editor) <= 0 {
		fmt.Println("Couldnt find editor. Either set $EDITOR or use the --editor flag")
		os.Exit(1)
	}

	if len(cliArgs.File) <= 0 {
		fmt.Println("Please specify a file!")
		flag.Usage()
		os.Exit(1)
	}

	if len(cliArgs.TrampDir) <= 0 {
		//user, err := user.Current()
		//if err != nil {
		//	fmt.Printf("Error fetching user")
		//	os.Exit(1)
		//}
		//cliArgs.TrampDir = fmt.Sprintf("%s/.supertramp", user.HomeDir)
		cliArgs.TrampDir = ".supertramp"
		cliArgs.TempDirPath = fmt.Sprintf("%s/%s", cliArgs.TrampDir, cliArgs.Host)
		cliArgs.TempMetaDirPath = fmt.Sprintf("%s/%s/.meta", cliArgs.TrampDir, cliArgs.Host)
		cliArgs.TempHashFilePath = fmt.Sprintf("%s/%s/.meta/%s", cliArgs.TrampDir, cliArgs.Host, filepath.Base(cliArgs.File))
		cliArgs.TempFilePath = fmt.Sprintf("%s/%s/%s", cliArgs.TrampDir, cliArgs.Host, filepath.Base(cliArgs.File))
	}

	if len(cliArgs.Username) <= 0 {
		usr, _ := user.Current()
		cliArgs.Username = usr.Username
		}

	if len(cliArgs.KeyFile) <= 0 {
		usr, _ := user.Current()
		cliArgs.KeyFile = fmt.Sprintf("%s/.ssh/id_rsa", usr.HomeDir)
		}

	cliArgs.SSHConfig = generateConfig(cliArgs)
	return cliArgs, nil

}

func generateConfig (cliargs ConnectionConfig) (*ssh.ClientConfig) {
	var authMethod ssh.AuthMethod
	var hostKey ssh.PublicKey

	if len(cliargs.Password) > 0 {
		authMethod = ssh.Password(cliargs.Password)
	} else if len(cliargs.KeyFile) > 0 {
		key, err := ioutil.ReadFile(cliargs.KeyFile)
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Fatalf("Can't read key!")
			return nil
		} else {
			authMethod = ssh.PublicKeys(signer)
		}
	}
	sshConfig := &ssh.ClientConfig{
		User: cliargs.Username,
		Auth: []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey() //todo
	return sshConfig
}


