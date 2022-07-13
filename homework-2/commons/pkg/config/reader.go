package config

import "gopkg.in/yaml.v2"

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
