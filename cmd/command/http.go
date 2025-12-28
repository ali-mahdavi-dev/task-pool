package command

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"task-pool/config"
	postgresrepo "task-pool/internal/adapter/repository/postgres"
	service "task-pool/internal/application"
	"task-pool/internal/domain/entity"
	"task-pool/internal/entrypoint"
	"task-pool/internal/entrypoint/handler"
	"task-pool/internal/worker"
	"task-pool/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func runHTTPServerCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "http",
		Short: "start http server",
		RunE: func(_ *cobra.Command, _ []string) error {
			initializeConfigs()

			log.Println("starting task-pool http server")

			return runHTTPServer(cfg)
		},
	}
}

func runHTTPServer(conf config.Config) error {
	app := fiber.New()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Bootstrap the application
	bootstrapResult, bErr := bootstrap(ctx, app, conf)
	if bErr != nil {
		return bErr
	}

	shutdown(ctx, app, bootstrapResult, conf)
	logger.Info("Starting HTTP server on port").WithInt("port", conf.Server.Port).Log()

	aErr := app.Listen(fmt.Sprintf(":%d", conf.Server.Port))
	if aErr != nil {
		logger.Error("Failed to start HTTP server").WithError(aErr).Log()
		return fmt.Errorf("failed to start HTTP server: %w", aErr)
	}
	logger.Info("Graceful shutdown completed").Log()

	return nil
}

type bootstrapResult struct {
	taskWorker worker.Worker[*entity.Task]
}

func bootstrap(ctx context.Context, app *fiber.App, conf config.Config) (*bootstrapResult, error) {
	db, err := setupDB(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	// Initialize channel
	taskChannel := make(chan *entity.Task, conf.TaskWorker.QueueSize)

	// Initialize repository
	taskRepository := postgresrepo.NewTaskRepository(db)

	// Initialize service
	taskService := service.NewTaskService(taskRepository)

	// Initialize handler
	taskHandler := handler.NewTaskHandler(taskService)

	// Register handlers
	entrypoint.RegisterHttpHandlers(app, entrypoint.HandlerOptions{
		TaskHandler: taskHandler,
	})

	// Initialize worker
	taskWorker := worker.NewTaskWorker(taskRepository, conf, taskChannel)

	// Start worker with context
	taskWorker.Run(context.Background())

	return &bootstrapResult{
		taskWorker: taskWorker,
	}, nil
}

func setupDB(conf config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		conf.Database.Host,
		conf.Database.Port,
		conf.Database.Username,
		conf.Database.Password,
		conf.Database.Name,
		conf.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to connect to database").WithError(err).Log()
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Failed to get underlying sql.DB").WithError(err).Log()
		return nil, err
	}

	sqlDB.SetMaxOpenConns(conf.Database.MaxOpenConnections)

	logger.Info("Database connection successfully").Log()
	return db, nil
}

func shutdown(ctx context.Context, app *fiber.App, bootstrapResult *bootstrapResult, conf config.Config) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig

		ctx, cancel := context.WithTimeout(context.Background(), conf.Server.ShutdownTimeout)
		defer cancel()

		bootstrapResult.taskWorker.Shutdown(ctx)
		logger.Info("Worker shutdown successfully").Log()

		if err := app.Shutdown(); err != nil {
			log.Printf("Error shutting down server: %v\n", err)
		}
		logger.Info("Server shutdown successfully").Log()

		<-ctx.Done()
	}()

}
