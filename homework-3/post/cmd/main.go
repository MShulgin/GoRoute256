package main

import (
	"context"
	"flag"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	confReader "gitlab.ozon.dev/MShulgin/homework-3/common/pkg/config"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/kafka"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
	"gitlab.ozon.dev/MShulgin/homework-3/post/internal/config"
	"gitlab.ozon.dev/MShulgin/homework-3/post/internal/post"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "config/post.yml", "conf file path")
	flag.Parse()

	var conf config.Config
	if err := confReader.ReadConfig(&conf, configPath, confReader.NewYamlReader()); err != nil {
		logger.Error("failed to read config file: " + err.Error())
		panic(err)
	}

	ctx := context.Background()
	db, err := sqlx.Connect("pgx", conf.Storage.PG.ConnString())
	if err != nil {
		logger.Error("failed to connect to database: " + err.Error())
		panic(err)
	}
	kafkaProducer, err := kafka.NewSyncProducer(conf.Kafka.Brokers)
	if err != nil {
		logger.Error("failed to create kafka producer: " + err.Error())
		panic(err)
	}
	postService := post.NewService(&post.PgStorage{Db: db}, kafkaProducer)

	if err := post.StartKafkaSubscription(postService, conf.Kafka, ctx); err != nil {
		logger.Error("failed to start kafka subscription: " + err.Error())
		panic(err)
	}

	select {}
}
