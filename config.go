package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	Token    string
	Projects map[string]int
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
