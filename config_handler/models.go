package config_handler

import "golang.org/x/crypto/ssh"

type ConnectionConfig struct {
	Username         string
	Password         string
	KeyFile          string
	Host             string
	File             string
	RemoteCommand    string
	SSHConfig        *ssh.ClientConfig
	TrampDir         string
	Port             string
	TempDirPath      string
	TempMetaDirPath  string
	TempHashFilePath string
	TempFilePath     string
	Editor			 string
}