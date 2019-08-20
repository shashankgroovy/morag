package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	tokenFile := os.Getenv("TOKEN_FILE")

	if _, err := os.Stat(tokenFile); err == nil {
		utils.Banner()

		// Token file exists
		// Validate if the access token is valid
		authJson, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			log.Println("Bad json in token file", err.Error())
			log.Println("Use the login command to authenticate again")
		}

		// we initialize our Users array
		var authToken utils.OAuthToken

		// we unmarshal our byteArray which contains our
		// jsonFile's content into 'users' which we defined above
		json.Unmarshal(authJson, &authToken)

		// Check if access token has expired
		err = authToken.ValidateAccessToken()
		if err != nil {
			// Get a new access token
			_ = authToken.GetNewAccessToken()
			log.Println("Access token expired. Fetching a new access token")
			fmt.Println("You are authenticated now!")
			fmt.Println("Use `morag help fetch` to learn more about how to get tracks")
		} else {
			fmt.Println("You are authenticated!")
			fmt.Println("Use `morag help fetch` to learn more about how to get tracks")
		}

	} else {
		cmd.Help()
	}
}
