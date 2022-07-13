package main

import (
	"context"
	"expvar"
	_ "expvar"
	"flag"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/stdlib"
	confReader "gitlab.ozon.dev/MShulgin/homework-3/common/pkg/config"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/db"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/kafka"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/internal/config"
	"gitlab.ozon.dev/MShulgin/homework-3/shipment/internal/shipment"
	"net/http"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "config/shipment.yml", "conf file path")
	flag.Parse()

	var conf config.Config
	if err := confReader.ReadConfig(&conf, configPath, confReader.NewYamlReader()); err != nil {
		logger.Error("failed to read config file: " + err.Error())
		panic(err)
	}

	ctx := context.Background()
	pgCluster, err := db.NewPgCluster(ctx, conf.Etcd.Servers, conf.Storage.PGCluster.Name)
	if err != nil {
		panic(err)
	}
	defer pgCluster.Close()

	kafkaProducer, err := kafka.NewSyncProducer(conf.Kafka.Brokers)
	if err != nil {
		logger.Error("failed to create kafka producer: " + err.Error())
		panic(err)
	}

	shipmentService := shipment.NewService(&shipment.PgStorage{Cluster: pgCluster}, kafkaProducer)

	if err := shipment.StartKafkaSubscription(shipmentService, conf.Kafka, ctx); err != nil {
		logger.Error("failed to start kafka subscription: " + err.Error())
		panic(err)
	}

	handler := shipment.NewHandler(shipmentService)

	router := mux.NewRouter()
	router.Handle("/stats", expvar.Handler())
	router.HandleFunc("/api/shipment", handler.SaveShipment).Methods(http.MethodPost)
	router.HandleFunc("/api/shipment", handler.FilterShipments).Methods(http.MethodGet).Queries("orderId", "{orderId}")
	router.HandleFunc("/api/shipment/{shipmentId}", handler.GetShipment).Methods(http.MethodGet)

	if err := http.ListenAndServe(conf.Server.Http.Addr, router); err != nil {
		logger.Error(err.Error())
	}
}
