package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	DeployDuration int `yaml:"deploy_duration"`
	Projects       map[string]struct {
		Token string
		Hosts []int
	}
	Zabbix struct {
		Endpoint string
		Username string
		Password string
	}
}

var config Config

func parseConfig() {
	configPath := "config.yml"

	response, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Printf("failed to read %s: %s", configPath, err)
		panic(err)
	}

	yaml.Unmarshal(response, &config)
}
