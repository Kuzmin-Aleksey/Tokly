package app

import (
	"FairLAP/internal/config"
	"FairLAP/internal/domain/service/detector"
	"FairLAP/internal/domain/service/groups"
	"FairLAP/internal/domain/service/lapconfig"
	"FairLAP/internal/domain/service/mask"
	"FairLAP/internal/domain/service/metrics"
	"FairLAP/internal/infrastructure/persistence/images"
	"FairLAP/internal/infrastructure/persistence/mysql"
	"FairLAP/internal/server"
	"FairLAP/pkg/contextx"
	"FairLAP/pkg/logx"
	"FairLAP/pkg/middlewarex"
	"FairLAP/pkg/yolo_model"
	"context"
	"github.com/gorilla/mux"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/lmittmann/tint"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(cfg *config.Config) {
	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	l := initLogger(cfg.Debug)

	db, err := mysql.Connect(cfg.MySQL)
	if err != nil {
		log.Fatal("connect to mysql fail: ", err)
	}
	defer db.Close()

	detectionsRepo := mysql.NewDetectionsRepo(db)
	groupsRepo := mysql.NewGroupsRepo(db)
	lapConfigRepo := mysql.NewLapConfigRepo(db)

	imagesRepo := images.New(cfg.ImagesPath)

	yoloConfig, err := yolo_model.ReadConfig(cfg.YoloModel.ModelConfig)
	if err != nil {
		log.Fatal("read yolo model config error: ", err)
	}

	yoloModel := yolo_model.NewModel(cfg.YoloModel.Model, yoloConfig)
	defer yoloModel.Close()

	detectorService := detector.NewService(yoloModel, detectionsRepo, imagesRepo)
	groupsService := groups.NewService(groupsRepo, imagesRepo)
	lapConfigService := lapconfig.NewService(lapConfigRepo, cfg.DefaultLapConfig)
	metricsService := metrics.NewService(groupsRepo, detectionsRepo, lapConfigService)
	maskService := mask.NewService(detectionsRepo, nil)

	httpServer := newHttpServer(l, detectorService, groupsService, metricsService, lapConfigService, maskService, imagesRepo, cfg.Http)

	go func() {
		if cfg.Http.SSLCertPath != "" && cfg.Http.SSLKeyPath != "" {
			log.Println("Starting https server on", cfg.Http.Host)
			if err := httpServer.ListenAndServeTLS(cfg.Http.SSLCertPath, cfg.Http.SSLKeyPath); err != nil {
				log.Fatal("Listen http: ", err)
			}
			return
		}
		log.Println("Starting http server on", cfg.Http.Host)
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal("Listen http: ", err)
		}
	}()

	sig := <-shutdown
	log.Println("exit by signal: ", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Println("http server shutdown error: ", err)
	}
	cancel()

	os.Exit(0)
}

func newHttpServer(
	l *slog.Logger,
	detector *detector.Service,
	groups *groups.Service,
	metrics *metrics.Service,
	lapConfig *lapconfig.Service,
	mask *mask.Service,
	images *images.Images,
	cfg *config.HttpConfig,
) *http.Server {
	analyzerServer := server.NewDetectorServer(detector)
	groupsServer := server.NewGroupsServer(groups)
	metricsServer := server.NewMetricServer(metrics)
	imagesServer := server.NewImagesServer(images)
	lapConfigServer := server.NewLapConfigServer(lapConfig)
	maskServer := server.NewMaskService(mask)

	s := server.NewServer(
		analyzerServer,
		groupsServer,
		metricsServer,
		lapConfigServer,
		maskServer,
		imagesServer,
	)

	rtr := mux.NewRouter()
	s.InitRoutes(rtr)

	rtr.Use(
		middlewarex.TraceId,
		middlewarex.Logger,
		middlewarex.RequestLogging(logx.NewSensitiveDataMasker(), 1000),
		middlewarex.ResponseLogging(logx.NewSensitiveDataMasker(), 1000),
		middlewarex.NoCache,
		middlewarex.Recovery,
	)
	if cfg.HandleTimeoutSec > 0 {
		rtr.Use(middlewarex.WithTimeout(time.Duration(cfg.HandleTimeoutSec) * time.Second))
	}

	return &http.Server{
		Addr:         cfg.Host,
		Handler:      rtr,
		ReadTimeout:  time.Duration(cfg.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeoutSec) * time.Second,
		ErrorLog:     slog.NewLogLogger(l.Handler(), slog.LevelError),
		BaseContext: func(net.Listener) context.Context {
			return contextx.WithLogger(context.Background(), l)
		},
	}
}

func initLogger(debug bool) *slog.Logger {
	if debug {
		return slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:   false,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
			TimeFormat:  time.StampMilli,
			NoColor:     false,
		}))
	}

	rt, err := rotatelogs.New("logs/%Y-%m-%d.log",
		rotatelogs.WithRotationTime(time.Hour*24),
		rotatelogs.WithMaxAge(time.Hour*24*15),
	)
	if err != nil {
		log.Fatal(err)
	}

	return slog.New(slog.NewJSONHandler(io.MultiWriter(os.Stdout, rt), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
