package config

import "fmt"

type HttpServerConfig struct {
	Addr string `yaml:"addr"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	GroupId string   `yaml:"groupId"`
}

type PGConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func (c *PGConfig) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, c.Database)
}

type CacheConfig struct {
	Addr string `yaml:"addr"`
	Type string `yaml:"type"`
}

type EtcdConfig struct {
	Servers []string `yaml:"servers"`
}

type PGClusterConfig struct {
	Name string `yaml:"name"`
}
