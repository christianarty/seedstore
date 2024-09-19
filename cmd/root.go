package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var cfgDir string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "Seedstore",
	Short: "Automatically download file from remote server using MQTT and LFTP",
	Long: `
Seedstore is a CLI binary that empowers allows for users to easily subscribe and
easily download from a remote location, initiating the connection from the client.`,
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.seedstore/config.json)")
	rootCmd.PersistentFlags().StringVar(&cfgDir, "configDir", "", "custom config directory (default is $HOME/.seedstore)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		defaultSeedstoreConfigDir := filepath.Join(home, ".seedstore")

		if cfgDir != "" {
			viper.AddConfigPath(cfgDir)
		}
		// Search config in home directory with name ".seedstore" (without extension).
		viper.AddConfigPath(defaultSeedstoreConfigDir)
		viper.SetConfigType("json")
		viper.SetConfigName("config")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("SEEDSTORE")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

}
