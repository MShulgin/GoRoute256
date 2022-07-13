package config

import "time"

type Config struct {
	Server struct {
		Grpc struct {
			Addr string
		}
		Http struct {
			Addr string
		}
	}
	Storage struct {
		PG struct {
			Host     string
			Port     int
			User     string
			Password string
			Database string
		}
		Schema struct {
			Path string
		}
	}
	Service struct {
		ValueUpdater struct {
			Tick time.Duration
		} `yaml:"value_updater"`
	}
}
