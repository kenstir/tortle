/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var verbosity int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tt",
	Short: "tt (torrent tool or tortle) is a tool for querying Deluge or qBittorrent",
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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./tt.toml)")

	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase verbosity (may be specified multiple times)")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Use ./tt.toml or $HOME/tt.toml
		viper.AddConfigPath(".")
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
