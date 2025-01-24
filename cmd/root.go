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
	viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
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

	// Read in environment variables that match flags
	//viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
