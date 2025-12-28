package command

import (
	"log"
	"task-pool/config"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	envFile string
	cfg     config.Config
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
		panic(err)
	}

	c, err := config.Load()
	if err != nil {
		log.Fatalf("could not load configuration %s\n", err.Error())
	}

	cfg = *c
}

func Execute() {
	rootCmd.Execute()
}
