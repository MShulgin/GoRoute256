package config

import "gitlab.ozon.dev/MShulgin/homework-3/common/pkg/config"

type Config struct {
	Server struct {
		Http config.HttpServerConfig `yaml:"http"`
	} `yaml:"server"`
	Kafka   config.KafkaConfig `yaml:"kafka"`
	Storage struct {
		PG    config.PGConfig    `yaml:"pg"`
		Cache config.CacheConfig `yaml:"cache"`
	} `yaml:"storage"`
}
