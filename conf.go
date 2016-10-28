package main

import (
	"encoding/json"
	"io/ioutil"
)

type Conf struct {
	Host    string `json:"host"`
	Project string `json:"project"`
	Token   string `json:"token"`
}

func loadConf(filename string) (*Conf, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cf Conf
	if err := json.Unmarshal(b, &cf); err != nil {
		return nil, err
	}

	return &cf, nil
}
