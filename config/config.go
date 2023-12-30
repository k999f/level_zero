package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	DbUser string `json:"db_user"`
	DbPassword string `json:"db_password"`
	DbName string `json:"db_name"`
	ServerPort string `json:"server_port"`
	NatsUrl string `json:"nats_url"`
	NatsSubject string `json:"nats_subject"`
}

func InitialzeConfig(path string) Config {
	config := Config{}
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Reading config file error: ", err)
	}
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal("Unmarshalsing config file error: ", err)
	}
	return config
}