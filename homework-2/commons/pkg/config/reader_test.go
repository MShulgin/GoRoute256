package config

import (
	"testing"
)

type yamlTestConfig struct {
	A struct {
		B string `yaml:"B"`
		C int    `yaml:"C"`
	} `yaml:"A"`
}

var yamlContent = "A:\n B: \"42\"\n C: 42\n"

func TestReadYamlConfig(t *testing.T) {
	in := []byte(yamlContent)

	var actual yamlTestConfig
	err := NewYamlReader().ReadConfig(in, &actual)
	if err != nil {
		t.Errorf("Unexpected error: " + err.Error())
		t.FailNow()
	}
	expected := yamlTestConfig{
		A: struct {
			B string `yaml:"B"`
			C int    `yaml:"C"`
		}{B: "42", C: 42},
	}
	if actual != expected {
		t.Errorf("Actual config %+v not equal to expected %+v", actual, expected)
	}
}
