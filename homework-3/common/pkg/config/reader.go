package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Reader interface {
	ReadConfig(in []byte, dst any) error
}

type yamlReader struct{}

func NewYamlReader() Reader {
	return yamlReader{}
}

func (y yamlReader) ReadConfig(in []byte, dst any) error {
	err := yaml.Unmarshal(in, dst)
	if err != nil {
		return err
	}
	return nil
}

func ReadConfig(dst any, configPath string, cfgReader Reader) error {
	confContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	err = cfgReader.ReadConfig(confContent, dst)
	if err != nil {
		return err
	}

	return nil
}
