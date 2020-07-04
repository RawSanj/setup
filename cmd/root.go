/*
Copyright © 2020 Sanjay Rawat <rawsanj.dev>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

var cfgFile string
var applicationConfiguration Configuration

const (
	Kafka     string = "kafka"
	Cassandra string = "cassandra"
)

type Configuration struct {
	Info     string             `yaml:"info"`
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Name             string `yaml:"name"`
	Url              string `yaml:"url"`
	InstallationPath string `yaml:"path"`
	IsEnabled        bool   `yaml:"enabled"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "setup",
	Short: "setup is a cli tool written in Go to download and install development services",
	Long: `setup is a cli tool written in Go to download and install development services.
It also helps in setting shortcut or aliases to run them.

Currently it supports the following services:
	1. Kafka
	2. Zookeeper
	3. Cassandra

Author: Sanjay Rawat - https://rawsanj.dev`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfigFile()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return initializeApplicationConfiguration()
	},
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.setup.yaml)")

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

		// Search config in home directory with name ".setup" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".setup")
		viper.SetConfigType("yml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	//if err := viper.ReadInConfig(); err == nil {
	//	fmt.Println("Using config file:", viper.ConfigFileUsed())
	//} else {
	//	fmt.Println("Error Reading Config", err)
	//}
}

// Initialize and create Service Configuration yaml file
// Ignore file creation if $HOME/.setup.yml already exists
func initializeConfigFile() error {

	home, err := homedir.Dir()
	if err != nil {
		fmt.Println("Error getting User HOME directory. Please set the HOME environment pointing to your Home directory.", "Error is: ", err)
		os.Exit(1)
		return err
	}

	// Skip Config and File creation if file already exists
	if FileExists(filepath.FromSlash(home + "/.setup.yml")) {
		return nil
	}

	configuration := createApplicationConfig(home)
	marshal, err := yaml.Marshal(&configuration)
	if err != nil {
		return errors.New("Error Marshalling Configuration. Error: " + err.Error())
	}

	viper.Set("application", "setup is a cli tool written in Go to download and install development services.")
	viper.Set("version", "1.0")
	viper.Set("configuration", string(marshal))
	viper.AddConfigPath(home)
	viper.SetConfigName(".setup")
	viper.SetConfigType("yml")

	err = viper.SafeWriteConfig()
	if err != nil {
		return errors.New("Error Writing Config File. Error: " + err.Error())
	}
	return nil
}

// Create Service Configuration which is to be downloaded and installed
// Add more services here
func createApplicationConfig(home string) Configuration {
	kafka := Service{
		Name:             Kafka,
		Url:              "https://downloads.apache.org/kafka/2.5.0/kafka_2.12-2.5.0.tgz",
		InstallationPath: filepath.FromSlash(home + "/.bin" + "/" + Kafka),
		IsEnabled:        true,
	}
	cassandra := Service{
		Name:             Cassandra,
		Url:              "https://downloads.apache.org/cassandra/3.11.6/apache-cassandra-3.11.6-bin.tar.gz",
		InstallationPath: filepath.FromSlash(home + "/.bin" + "/" + Cassandra),
		IsEnabled:        true,
	}

	services := make(map[string]Service)
	services[Kafka] = kafka
	services[Cassandra] = cassandra

	configuration := Configuration{
		Info:     "you can customize below Configuration to point to an internal url or disable any service",
		Services: services,
	}

	return configuration
}

// Read Configuration from $HOME/.setup.yml and marshall & set into applicationConfiguration
func initializeApplicationConfiguration() error {
	services := viper.GetString("configuration")
	if services == "" {
		return errors.New("configuration initialization failed")
	}
	applicationConfiguration = Configuration{}

	err := yaml.Unmarshal([]byte(services), &applicationConfiguration)
	if err != nil {
		return errors.New("Error UnMarshalling Configuration. Error: " + err.Error())
	}
	return nil
}

// Check if file exists and is not a directory
func FileExists(fileName string) bool {
	stat, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !stat.IsDir()
}
