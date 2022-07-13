package main

import (
	"context"
	"flag"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	confReader "gitlab.ozon.dev/MShulgin/homework-3/common/pkg/config"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/kafka"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/metrics"
	"gitlab.ozon.dev/MShulgin/homework-3/offer/internal/config"
	"gitlab.ozon.dev/MShulgin/homework-3/offer/internal/offer"
	"net/http"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "config/offer.yml", "conf file path")
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

	offerCache, err := offer.NewCacheFromConfig(conf.Storage.Cache)
	if err != nil {
		logger.Error("failed to connect to cache: " + err.Error())
		panic(err)
	}
	offerRepo := offer.NewPgCachedStorage(offerCache, db)

	offerService := offer.NewService(offerRepo, kafkaProducer)

	if err = offer.StartKafkaSubscription(offerService, conf.Kafka, ctx); err != nil {
		logger.Error("failed to start kafka subscription: " + err.Error())
		panic(err)
	}

	offerHandler := offer.Handler{OfferService: offerService}
	router := mux.NewRouter()

	router.Use(metrics.PrometheusMiddleware)

	router.HandleFunc("/api/offer/{offerId}/price", offerHandler.GetOfferPrice).Methods(http.MethodGet)
	router.HandleFunc("/api/offer", offerHandler.SaveOffer).Methods(http.MethodPost)

	reg := prometheus.NewRegistry()
	reg.MustRegister(metrics.HttpDuration)
	reg.MustRegister(collectors.NewBuildInfoCollector())
	reg.MustRegister(collectors.NewGoCollector(
		collectors.WithGoCollections(collectors.GoRuntimeMetricsCollection),
	))
	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true}))

	if err := http.ListenAndServe(conf.Server.Http.Addr, router); err != nil {
		logger.Error(err.Error())
	}
}
