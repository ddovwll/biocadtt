package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ddovwll/biocadtt/src/internal/application/services"
	"github.com/ddovwll/biocadtt/src/internal/infrastructure/data"
	"github.com/ddovwll/biocadtt/src/internal/infrastructure/data/repositories"
	"github.com/ddovwll/biocadtt/src/internal/infrastructure/reportgen"
	"github.com/ddovwll/biocadtt/src/internal/infrastructure/tsv"
	"github.com/ddovwll/biocadtt/src/internal/presentation/http_api/controllers"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := initLogger(cfg.AppEnv)
	pool, err := initDB(cfg.PostgresConfig())
	if err != nil {
		logger.Error("failed to initialize postgres pool", "error", err)
		os.Exit(1)
	}

	deviceDataRepository, processedFileRepository := initRepositories(pool)
	tsvReader := initTsvReader(logger)
	txManager := initTxManager(pool)
	pdfGenerator, err := initPdfGenerator(cfg.ReportsDir)
	if err != nil {
		logger.Error("failed to initialize pdf generator", "error", err)
		os.Exit(1)
	}

	deviceDataService := initServices(
		deviceDataRepository,
		processedFileRepository,
		tsvReader,
		txManager,
		pdfGenerator,
	)
	deviceDataController := initController(deviceDataService, logger)
	server := initServer(deviceDataController, cfg.HTTPAddr)
	tsvWatcher := initTsvWatcher(deviceDataService, logger, cfg.WatcherDir)
	watcherCtx, cancelWatcherCtx := context.WithCancel(context.Background())

	go func() {
		tsvWatcher.Start(watcherCtx, cfg.WatcherHeartbeat)
	}()
	logger.Info("tsv watcher started")

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error starting http server",
				"error", err,
			)
		}
	}()
	logger.Info("starting server")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cancelWatcherCtx()
	logger.Info("tsv watcher shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("error shutting down http server",
			"error", err,
		)
	}
	logger.Info("server shut down")
	pool.Close()
}

func initLogger(env string) *slog.Logger {
	levels := map[string]slog.Level{
		"dev":  slog.LevelDebug,
		"prod": slog.LevelInfo,
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: env == "dev",
		Level:     levels[env],
	})

	return slog.New(handler)
}

func initDB(pgCfg data.PostgresConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return data.NewPgxPool(ctx, pgCfg)
}

func initRepositories(pool *pgxpool.Pool) (*repositories.DeviceDataRepo, *repositories.ProcessedFileRepo) {
	deviceDataRepository := repositories.NewDeviceDataRepository(pool)
	processedFileRepository := repositories.NewProcessedFileRepo(pool)

	return deviceDataRepository, processedFileRepository
}

func initTxManager(pool *pgxpool.Pool) *data.TxManager {
	return data.NewTxManager(pool)
}

func initServices(
	deviceDataRepo *repositories.DeviceDataRepo,
	processedFileRepo *repositories.ProcessedFileRepo,
	tsvReader *tsv.TsvReader,
	txManager *data.TxManager,
	reportGenerator *reportgen.PdfGenerator,
) *services.DeviceDataService {
	return services.NewDeviceDataService(deviceDataRepo, processedFileRepo, tsvReader, txManager, reportGenerator)
}

func initController(
	deviceDataService *services.DeviceDataService,
	logger *slog.Logger,
) *controllers.DeviceDataController {
	return controllers.NewDeviceDataController(deviceDataService, logger)
}

func initServer(deviceDataController *controllers.DeviceDataController, port string) *http.Server {
	mux := http.NewServeMux()
	deviceDataController.UseController(mux)

	return &http.Server{
		Addr:    port,
		Handler: mux,
	}
}

func initTsvWatcher(
	deviceDataService *services.DeviceDataService,
	logger *slog.Logger,
	directory string,
) *tsv.TsvWatcher {
	return tsv.NewTsvWatcher(deviceDataService, logger, directory)
}

func initTsvReader(logger *slog.Logger) *tsv.TsvReader {
	return tsv.NewTsvReader(logger)
}

func initPdfGenerator(outputDir string) (*reportgen.PdfGenerator, error) {
	return reportgen.NewService(outputDir)
}
