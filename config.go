package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	LnAddr    string `json:"ln_addr"`
	HandleSer string `json:"handle_ser_addr"`
	StaticSer string `json:"static_ser_addr"`
}

func ParseConfig(path string) (cfg *Config, err error) {
	file, err := os.Open(path)

	if err != nil {
		fmt.Println("open config.json err: " + err.Error())
		return nil, err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("read config.json err: " + err.Error())
		return nil, err
	}

	cfg = &Config{}

	if err = json.Unmarshal(data, cfg); err != nil {
		fmt.Println("unmarshal json err :" + err.Error())
		return nil, err
	}

	return cfg, err
}
