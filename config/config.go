package config

import (
	"encoding/json"
	"io/ioutil"
)

type Configuration struct {
	TemplateDir  string `json:"templatedir"`
	LayoutName   string `json:"layoutname"`
	BareName     string `json:"barename"`
	StaticDir    string `json:"staticdir"`
	RenderDir    string `json:"renderdir"`
	ChecksumName string `json:"checksumname"`
	BackendType  string `json:"backend"`
	Hostname     string `json:"hostname"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func Read(filename string) (Configuration, error) {
	config := Configuration{}

	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
