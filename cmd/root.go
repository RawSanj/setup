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
	"com.github/RawSanj/setup/util"
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

const (
	Kafka            string = "kafka"
	Cassandra        string = "cassandra"
	ConfigurationKey string = "configuration"
)

type Configuration struct {
	Info     string             `yaml:"info"`
	Services map[string]Service `yaml:"services"`
}

type VersionMap map[string]string

type Service struct {
	Name             string                `yaml:"name"`
	UrlTemplate      string                `yaml:"urlTemplate"`
	InstallationPath string                `yaml:"path"`
	IsEnabled        bool                  `yaml:"enabled"`
	Versions         map[string]VersionMap `yaml:"availableVersions"`
	SelectedVersion  string                `yaml:"defaultVersion"`
	ActiveVersion    string                `yaml:"activeVersion"`
	InstalledVersion []string              `yaml:"installedVersion"`
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
		return cmd.Help()
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
	_ = viper.ReadInConfig()
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
	if util.FileExists(filepath.FromSlash(home + "/.setup.yml")) {
		return nil
	}

	configuration := createApplicationConfig(home)
	marshal, err := yaml.Marshal(&configuration)
	if err != nil {
		return errors.New("Error Marshalling Configuration. Error: " + err.Error())
	}

	viper.Set("application", "setup is a cli tool written in Go to download and install development services.")
	viper.Set("version", "1.0")
	viper.Set(ConfigurationKey, string(marshal))
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
// Add more services here in future releases
func createApplicationConfig(home string) Configuration {

	kafkaVersions := []VersionMap{
		{"Scala": "2.11", "Version": "1.0.0"}, {"Scala": "2.12", "Version": "1.0.0"}, {"Scala": "2.11", "Version": "1.0.1"}, {"Scala": "2.12", "Version": "1.0.1"}, {"Scala": "2.11", "Version": "1.0.2"}, {"Scala": "2.12", "Version": "1.0.2"},
		{"Scala": "2.11", "Version": "1.1.0"}, {"Scala": "2.12", "Version": "1.1.0"}, {"Scala": "2.11", "Version": "1.1.1"}, {"Scala": "2.12", "Version": "1.1.1"},
		{"Scala": "2.11", "Version": "2.0.0"}, {"Scala": "2.12", "Version": "2.0.0"}, {"Scala": "2.11", "Version": "2.0.1"}, {"Scala": "2.12", "Version": "2.0.1"},
		{"Scala": "2.11", "Version": "2.1.0"}, {"Scala": "2.12", "Version": "2.1.0"}, {"Scala": "2.11", "Version": "2.1.1"}, {"Scala": "2.12", "Version": "2.1.1"},
		{"Scala": "2.11", "Version": "2.2.0"}, {"Scala": "2.12", "Version": "2.2.0"}, {"Scala": "2.11", "Version": "2.2.1"}, {"Scala": "2.12", "Version": "2.2.1"}, {"Scala": "2.11", "Version": "2.2.2"}, {"Scala": "2.12", "Version": "2.2.2"},
		{"Scala": "2.11", "Version": "2.3.0"}, {"Scala": "2.12", "Version": "2.3.1"}, {"Scala": "2.11", "Version": "2.3.1"}, {"Scala": "2.12", "Version": "2.3.1"},
		{"Scala": "2.11", "Version": "2.4.0"}, {"Scala": "2.12", "Version": "2.4.0"}, {"Scala": "2.13", "Version": "2.4.0"}, {"Scala": "2.11", "Version": "2.4.1"}, {"Scala": "2.12", "Version": "2.4.1"}, {"Scala": "2.13", "Version": "2.4.1"},
		{"Scala": "2.12", "Version": "2.5.0"}, {"Scala": "2.13", "Version": "2.5.0"},
	}
	kafka := Service{
		Name:             Kafka,
		UrlTemplate:      "https://archive.apache.org/dist/kafka/{{.Version}}/kafka_{{.Scala}}-{{.Version}}.tgz",
		Versions:         createVersionMap(&kafkaVersions),
		SelectedVersion:  "Version_2.5.0-Scala_2.13",
		InstallationPath: filepath.FromSlash(home + "/.bin" + "/" + Kafka),
		IsEnabled:        true,
	}

	cassandraVersions := []VersionMap{{"Version": "2.1.21"}, {"Version": "2.2.16"}, {"Version": "3.0.20"}, {"Version": "3.11.6"}, {"Version": "4.0-alpha4"}}
	cassandra := Service{
		Name:             Cassandra,
		UrlTemplate:      "https://downloads.apache.org/cassandra/{{.Version}}/apache-cassandra-{{.Version}}-bin.tar.gz",
		Versions:         createVersionMap(&cassandraVersions),
		SelectedVersion:  "Version_3.11.6",
		InstallationPath: filepath.FromSlash(home + "/.bin" + "/" + Cassandra),
		IsEnabled:        true,
	}

	services := make(map[string]Service)
	services[Kafka] = kafka
	services[Cassandra] = cassandra

	configuration := Configuration{
		Info:     "Customize below Configuration to point to an internal url, versions or disable any service. When using internal URL with version, make sure url template is proper",
		Services: services,
	}

	return configuration
}

func createVersionMap(versions *[]VersionMap) map[string]VersionMap {
	versionMap := make(map[string]VersionMap)
	for _, version := range *versions {
		versionKey := ""
		index := 1
		for key, val := range version {
			if index < len(version) {
				versionKey = versionKey + fmt.Sprintf("%s_%s-", key, val)
			} else {
				versionKey = versionKey + fmt.Sprintf("%s_%s", key, val)
			}
			index++
		}
		versionMap[versionKey] = version
	}
	return versionMap
}
