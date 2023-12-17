package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Database DatabaseConfig `json:"database"`
	AMI AMIConfig           `json:"ami"`
	OAMI AMIConfig          `json:"oami"`
	Global GlobalConfig	`json:"global"`
}

type DatabaseConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Name     string `json:"name"`
}

type AMIConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Enabled  string `json:"enabled"`
}

type GlobalConfig struct {
	Maxcalls  string `json:"maxcalls"`
	MaxCPS string `json:"maxcps"`
	QueueCPS string `json:"queuecps"`
	HttpPort string `json:"httpport"`
	HttpsPort string `json:"httpsport"`
	HttpsPriKey string `json:"ssl_privatekey"`
	HttpsPubKey string `json:"ssl_fullchain"`
	AllowedIP string `json:"allowedips"`
}


func LoadConfig(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	config := Config{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
