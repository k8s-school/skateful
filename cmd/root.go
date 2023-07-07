/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	logger   *zap.SugaredLogger
	logLevel int
	storage  string
	outFile  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "skateful",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	cobra.OnInitialize(initLogger)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.skateful.yaml)")

	flag.Parse()

	rootCmd.PersistentFlags().StringVar(&storage, "storage", csi, "Storage type: 'csi' or 'local'")
	rootCmd.PersistentFlags().StringVar(&outFile, "out", "pvc-pv.yaml", "Path to output file")

	viper.BindPFlag("storage", rootCmd.PersistentFlags().Lookup("storage"))
	viper.BindPFlag("out", rootCmd.PersistentFlags().Lookup("out"))
}

// setUpLogs set the log output ans the log level
func initLogger() {

	rawJSON := []byte(`{
		"level": "debug",
		"encoding": "console",
		"outputPaths": ["stdout", "/tmp/logs"],
		"errorOutputPaths": ["stderr"],
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "levelEncoder": "lowercase"
		}
	  }`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	_logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer _logger.Sync()
	logger = _logger.Sugar()

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
