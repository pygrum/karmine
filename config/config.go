package config

import (
	"encoding/json"
	"os"
)

type configData struct {
	BinPath      string    `json:"bin_path"`
	SrcPath      string    `json:"src_path"`
	CertFilePath string    `json:"cert_pem"`
	KeyFilePath  string    `json:"key_pem"`
	Endpoint     string    `json:"endpoint"`
	DBPath       string    `json:"db"`
	SslDomain    string    `json:"ssl_domain"`
	Hosts        []Profile `json:"hosts"`
	Stages       Stages    `json:"stages"`
}

type Stages struct {
	File string `json:"filestage"`
	Cmd  string `json:"cmdstage"`
}

type Profile struct {
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Strain string `json:"strain"`
}

func ConfigPath() string {
	path, _ := os.UserHomeDir()
	path += "/.konfig"
	return path
}

func GetBinPath() (string, error) {
	bytes, err := os.ReadFile(ConfigPath())
	if err != nil {
		return "", err
	}
	config := &configData{}
	if err = json.Unmarshal(bytes, config); err != nil {
		return "", err
	}
	return config.BinPath, nil
}

func GetFullConfig() (*configData, error) {
	bytes, err := os.ReadFile(ConfigPath())
	if err != nil {
		return nil, err
	}
	config := &configData{}
	if err = json.Unmarshal(bytes, config); err != nil {
		return nil, err
	}
	return config, nil
}

// Returns the absolute paths of the ssl key pair files, which are stored in the configuration file.
func GetSSLPair() (string, string, error) {
	bytes, err := os.ReadFile(ConfigPath())
	if err != nil {
		return "", "", err
	}
	config := &configData{}
	if err = json.Unmarshal(bytes, config); err != nil {
		return "", "", err
	}
	return config.CertFilePath, config.KeyFilePath, nil
}
