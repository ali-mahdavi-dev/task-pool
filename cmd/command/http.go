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
	"task-pool/internal/domain/entity"
	"task-pool/internal/entrypoint"
	"task-pool/internal/entrypoint/handler"
	"task-pool/internal/service"
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

			return runHTTPServer(Cfg)
		},
	}
}

func runHTTPServer(cfg config.Config) error {
	app := fiber.New()

	// Bootstrap the application
	bootstrapResult, bErr := bootstrap(app, cfg)
	if bErr != nil {
		return bErr
	}

	shutdown(app, bootstrapResult, cfg)
	logger.Info("Starting HTTP server on port").WithInt("port", cfg.Server.Port).Log()

	aErr := app.Listen(fmt.Sprintf(":%d", cfg.Server.Port))
	if aErr != nil {
		logger.Error("Failed to start HTTP server").WithError(aErr).Log()
		return fmt.Errorf("failed to start HTTP server: %w", aErr)
	}
	logger.Info("Graceful shutdown completed").Log()

	return nil
}

type bootstrapResult struct {
	taskWorker  worker.Worker[*entity.Task]
	taskChannel chan *entity.Task
}

func bootstrap(app *fiber.App, cfg config.Config) (*bootstrapResult, error) {
	db, err := setupDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	// Initialize channel
	taskChannel := make(chan *entity.Task, cfg.TaskWorker.QueueSize)

	// Initialize repository
	taskRepository := postgresrepo.NewTaskRepository(db)

	// Initialize service
	taskService := service.NewTaskService(taskRepository, taskChannel)

	// Initialize handler
	taskHandler := handler.NewTaskHandler(taskService)

	// Register handlers
	entrypoint.RegisterHttpHandlers(app, entrypoint.HandlerOptions{
		TaskHandler: taskHandler,
	})

	// Initialize worker
	taskWorker := worker.NewTaskWorker(taskRepository, cfg, taskChannel)

	// Start worker with context
	taskWorker.Run(context.Background())

	return &bootstrapResult{
		taskWorker:  taskWorker,
		taskChannel: taskChannel,
	}, nil
}

func setupDB(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode)

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

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConnections)

	// Auto migrate database tables
	err = db.AutoMigrate(&entity.Task{})
	if err != nil {
		logger.Error("Failed to auto migrate database").WithError(err).Log()
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}

	logger.Info("Database connection successfully").Log()
	return db, nil
}

func shutdown(app *fiber.App, bootstrapResult *bootstrapResult, conf config.Config) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig

		ctx, cancel := context.WithTimeout(context.Background(), conf.Server.ShutdownTimeout)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Printf("Error shutting down server: %v\n", err)
		}

		bootstrapResult.taskWorker.Shutdown()
		logger.Info("Worker shutdown successfully").Log()

		close(bootstrapResult.taskChannel)
		logger.Info("Task channel closed successfully").Log()

		logger.Info("Server shutdown successfully").Log()

		<-ctx.Done()
	}()

}
