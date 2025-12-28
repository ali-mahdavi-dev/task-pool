package command

import (
	"fmt"
	"log"
	"task-pool/config"
	postgresrepo "task-pool/internal/adapter/repository/postgres"
	service "task-pool/internal/application"
	"task-pool/internal/entrypoint"
	"task-pool/internal/entrypoint/handler"
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

	// Bootstrap the application
	bErr := bootstrap(app, conf)
	if bErr != nil {
		return bErr
	}

	logger.Info("Starting HTTP server on port").WithInt("port", conf.Server.Port).Log()

	aErr := app.Listen(fmt.Sprintf(":%d", conf.Server.Port))
	if aErr != nil {
		logger.Error("Failed to start HTTP server").WithError(aErr).Log()
		return fmt.Errorf("failed to start HTTP server: %w", aErr)
	}

	return nil
}

func bootstrap(app *fiber.App, conf config.Config) error {
	db, err := setupDB(conf)
	if err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

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

	return nil
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
