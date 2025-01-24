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
	Use:   "torinfo",
	Short: "Query torrent info",
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./torinfo.toml)")

	rootCmd.PersistentFlags().StringP("server", "s", "localhost", "server address")
	rootCmd.PersistentFlags().IntP("port", "p", 9091, "server port")
	rootCmd.PersistentFlags().StringP("username", "U", "admin", "server username")
	rootCmd.PersistentFlags().StringP("password", "P", "password", "server password")
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase verbosity (may be specified multiple times)")
	viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Use ./torinfo.toml
		viper.AddConfigPath(".")
		viper.SetConfigName("torinfo")
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
