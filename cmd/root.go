/*
Copyright © 2024 MahdiNajazadeh
*/
package cmd

import (
	"fmt"
	"os"
	"startier/config"
	"startier/internal/node"

	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	appConfig *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "startier",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if appConfig == nil {
			fmt.Println("Config is not loaded.")
			return
		}
		// fmt.Printf("Config loaded: %+v\n", appConfig)
		n, err := node.New(appConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error In Initialisation Node : %+v", err)
			os.Exit(1)
		}
		if err = n.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error In Running Node : %+v", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().
		StringVar(&cfgFile, "config", "", "config file (default is /etc/startier/config.json)")
}

func initConfig() {
	var err error
	configPath := "/etc/startier"
	if cfgFile != "" {
		configPath = cfgFile
	}
	appConfig, err = config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	// fmt.Fprintf(os.Stderr, "Using config: %+v\n", appConfig)
}
