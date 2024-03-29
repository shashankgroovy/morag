package cmd

import (
	"fmt"
	"os"

	"github.com/shashankgroovy/morag/utils"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "morag",
	Short: "Morag is a command-line tool that lets you download the entire catalog/library of an artist from Spotify",
	Long: `Morag lets your download the entire catalog/library of an artist
from Spotify and saves it to a csv file. It uses OAuth2 for authentication.

Issue the login command to start fetching data from Spotify`,
	Run: root,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.morag.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".morag" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".morag")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Calls help if a user is not logged in else shows app banner and tries to
// refresh access token if not expired.
func root(cmd *cobra.Command, args []string) {

	if _, err := utils.TestAndSetToken(); err == nil {
		fmt.Println("Use `morag help fetch` to learn more about how to get tracks")
	} else {
		cmd.Help()
	}
}
