package command

import (
	"log"
	"task-pool/config"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	envFile string
	Cfg     config.Config
	rootCmd = &cobra.Command{
		Use: "",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfigs()
		},
	}
)

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().StringVarP(&envFile, "env-file", "e", ".env", ".env file")

	rootCmd.AddCommand(runHTTPServerCMD())
}

func initializeConfigs() {
	err := godotenv.Load(envFile)
	if err != nil {
		log.Printf("warning: could not load .env file: %v\n", err)
	}

	c, err := config.Load()
	if err != nil {
		log.Fatalf("could not load configuration %s\n", err.Error())
	}

	Cfg = *c
}

func Execute() {
	rootCmd.Execute()
}
