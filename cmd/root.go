/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
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
	Short: "Tools for managing stateful applications in kubernetes",
	Long: `Useful tools for managing stateful applications in kubernetes.
- Backup PVs/PVCs for a Qserv deployment which is based on CSI storage.
- TODO Can create PVCs/PVs for a Qserv deployment which use localStorageClass.`,
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
