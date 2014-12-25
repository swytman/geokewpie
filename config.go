package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Db
}

type Db struct {
	Host     string
	Port     string
	User     string
	Password string
	Dbname   string
}

func load_config(filename string) *Config {
	var config Config
	fmt.Printf("Config loading..  \r\n")
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}

	return &config
}
