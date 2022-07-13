package config

import "gitlab.ozon.dev/MShulgin/homework-3/common/pkg/config"

type Config struct {
	Kafka   config.KafkaConfig `yaml:"kafka"`
	Storage struct {
		PG config.PGConfig `yaml:"pg"`
	} `yaml:"storage"`
}
