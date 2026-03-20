package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var prettyPrint bool
var decode bool
var aiDecoder string

var rootCmd = &cobra.Command{
	Use:   "packetcats",
	Short: "A network packet generator and capture tool",
	Long: `packetcats is a CLI tool for generating and capturing network packets.
It supports various protocols including TCP, UDP, DNS, and more.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.packetcats.yaml)")
	rootCmd.PersistentFlags().BoolVar(&prettyPrint, "pretty-print", false, "Pretty print the output")
	rootCmd.PersistentFlags().BoolVar(&decode, "decode", false, "Decode the packet")
	rootCmd.PersistentFlags().StringVar(&aiDecoder, "ai-decoder", "", "AI decoder to use (chatgpt|claude|gemini|opencode)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".packetcats")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
