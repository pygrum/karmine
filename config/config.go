package config

import (
	"encoding/json"
	"os"
)

type configData struct {
	BinPath      string    `json:"bin_path"`
	SrcPath      string    `json:"src_path"`
	SqlUser      string    `json:"sql_user"`
	SqlPass      string    `json:"sql_pass"`
	CertFilePath string    `json:"cert_pem"`
	KeyFilePath  string    `json:"key_pem"`
	Endpoint     string    `json:"endpoint"`
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

func GetSqlCreds() (string, string, error) {
	bytes, err := os.ReadFile(ConfigPath())
	if err != nil {
		return "", "", err
	}
	config := &configData{}
	if err = json.Unmarshal(bytes, config); err != nil {
		return "", "", err
	}
	return config.SqlUser, config.SqlPass, nil
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
