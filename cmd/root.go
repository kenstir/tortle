/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var logFile io.WriteCloser
var verbosity int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tt",
	Short: "tt (torrent tool or tortle) is a multi-tool for Deluge and qBittorrent",
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
	cobra.OnInitialize(initConfig, initLogging)
	cobra.OnFinalize(finalizeLogging)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./tt.toml)")

	rootCmd.PersistentFlags().StringP("log-file", "l", "", "Log file (default is stdout)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase verbosity (may be specified multiple times)")
	viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	// fmt.Println("root init")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// fmt.Println("root initConfig")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Use tt.toml from ., exe_dir, or $HOME
		viper.AddConfigPath(".")
		if exePath, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exePath)
			viper.AddConfigPath(exeDir)
		}
		viper.AddConfigPath("$HOME")
		viper.SetConfigName("tt")
		viper.SetConfigType("toml")
	}

	// Read in config file and print error only if it's not a "not found" error
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config file: %s\n", err)
			os.Exit(1)
		}
	}

	// Read in environment variables that match flags
	// On second thought, don't; it causes $USERNAME to override the config file
	// viper.AutomaticEnv()
}

func initLogging() {
	//fmt.Println("root initLogging")

	// handle quiet first
	if viper.GetBool("quiet") {
		stdoutLogger.SetOutput(io.Discard)
		return
	}

	// open the log file
	logFileName := viper.GetString("log-file")
	if logFileName != "" {
		logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", logFileName, err)
			os.Exit(1)
		}
		stdoutLogger.SetOutput(logFile)
	}
}

func finalizeLogging() {
	//fmt.Println("root finalizeLogging")

	if logFile != nil {
		if logFile != nil {
			logFile.Close()
		}
	}
}

// exitWithFinalizers closes the log file and exits with the given code
func exitWithFinalizers(code int) {
	finalizeLogging()
	os.Exit(code)
}

// fatalError closes the log file, prints the error message to stderr, and exits 1
func fatalError(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	exitWithFinalizers(1)
}
