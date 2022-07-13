package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	confReader "gitlab.ozon.dev/MShulgin/homework-2/commons/pkg/config"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/config"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/logging"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/market"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/pb"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/server"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/service"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "config/portfolio.yml", "conf file path")
	var autoMigrate bool
	flag.BoolVar(&autoMigrate, "migrate", false, "Auto apply db migration before start server")
	flag.Parse()

	conf, err := readConfig(configPath, confReader.NewYamlReader())
	if err != nil {
		panic(err)
	}

	if autoMigrate {
		runDbMigration(conf)
	}

	logging.Info("Starting portfolio server...")

	go startGrpcProxy(conf)
	startGrpcServer(conf)
}

func readConfig(configPath string, cfgReader confReader.Reader) (*config.Config, error) {
	confContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		logging.Error("fail to read conf file: " + err.Error())
		return nil, err
	}
	var conf config.Config
	err = cfgReader.ReadConfig(confContent, &conf)
	if err != nil {
		logging.Error("fail to parse conf file: " + err.Error())
		return nil, err
	}
	return &conf, nil
}

func startGrpcServer(conf *config.Config) {
	lis, err := net.Listen("tcp", conf.Server.Grpc.Addr)
	if err != nil {
		log.Fatalln("failed to bind address:", err)
	}
	pgConfig := conf.Storage.PG
	pgConn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		pgConfig.User, pgConfig.Password, pgConfig.Host, pgConfig.Port, pgConfig.Database)
	db, err := sqlx.Connect("pgx", pgConn)
	if err != nil {
		log.Fatalln(err)
	}

	accountService := service.NewAccountService(store.NewAccountsRepo(db))
	portfolioService := service.NewPortfolioService(store.NewPortfolioRepo(db),
		store.NewAssetRepo(db), store.NewPortfolioValueRepo(db), market.NewMoexStockClient())
	dashboardService := service.NewDefaultDashboardService(store.NewPortfolioRepo(db), store.NewPortfolioValueRepo(db))
	appServer := server.NewServer(accountService, portfolioService, dashboardService)

	updateValueFn := func() {
		err := portfolioService.UpdatePortfolioValues()
		if err != nil {
			logging.Error("Error while updating portfolio value: " + err.Error())
		}
	}
	model.NewScheduledExecutor(1*time.Second, conf.Service.ValueUpdater.Tick).Run(updateValueFn)

	grpcServer := grpc.NewServer()
	pb.RegisterPortfolioServiceServer(grpcServer, &appServer)

	go registerGrpcShutdown(grpcServer)

	logging.Info(fmt.Sprintf("Binding grpc server '%s'", conf.Server.Grpc.Addr))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalln("failed to server:", err)
	}
}

func registerGrpcShutdown(server *grpc.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan

	_, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	server.GracefulStop()

	logging.Info("Shutting down grpc server")
	os.Exit(0)
}

func startGrpcProxy(conf *config.Config) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterPortfolioServiceHandlerFromEndpoint(ctx, mux, conf.Server.Grpc.Addr, opts)
	if err != nil {
		panic(err)
	}

	logging.Info(fmt.Sprintf("Binding http server '%s'", conf.Server.Http.Addr))
	if err := http.ListenAndServe(conf.Server.Http.Addr, mux); err != nil {
		panic(err)
	}
}

func runDbMigration(conf *config.Config) {
	schemaDir := conf.Storage.Schema.Path
	pgConfig := conf.Storage.PG
	pgConnStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		pgConfig.User, pgConfig.Password, pgConfig.Host, pgConfig.Port, pgConfig.Database)
	db, err := goose.OpenDBWithDriver("pgx", pgConnStr)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v\n", err)
	}
	defer db.Close()

	if err := goose.Run("up", db, schemaDir); err != nil {
		log.Fatalf("error when migrating database: %v", err)
	}
}
