/*
MIT License

Copyright (c) 2020 Sanjay Rawat - https://rawsanj.dev

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
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
	DynamoDb         string = "dynamodb"
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
	4. DynamoDb

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
	configurationString, err := marshalConfiguration(&configuration)
	if err != nil {
		return err
	}

	viper.Set("application", "setup is a cli tool written in Go to download and install development services.")
	viper.Set("version", "1.0")
	viper.Set(ConfigurationKey, configurationString)
	viper.AddConfigPath(home)
	viper.SetConfigName(".setup")
	viper.SetConfigType("yml")

	err = viper.SafeWriteConfig()
	if err != nil {
		return errors.New("Error Writing Config File. Error: " + err.Error())
	}
	return nil
}

func marshalConfiguration(configuration *Configuration) (string, error) {
	marshal, err := yaml.Marshal(&configuration)
	if err != nil {
		return "", errors.New("Error Marshalling Configuration. Error: " + err.Error())
	}
	return string(marshal), nil
}

func unMarshalConfiguration(configurationStr string, applicationConfiguration *Configuration) error  {
	err := yaml.Unmarshal([]byte(configurationStr), &applicationConfiguration)
	if err != nil {
		return errors.New("Error UnMarshalling Configuration. Error: " + err.Error())
	} else {
		return nil
	}
}

// Create Service Configuration which is to be downloaded and installed
// Add more services here in future releases
func createApplicationConfig(home string) Configuration {

	kafkaVersions := []VersionMap{
		{"Name": "kafka-2.11-1.0.0", "Scala": "2.11", "Version": "1.0.0"}, {"Name": "kafka-2.12-1.0.0", "Scala": "2.12", "Version": "1.0.0"},
		{"Name": "kafka-2.11-1.0.1", "Scala": "2.11", "Version": "1.0.1"}, {"Name": "kafka-2.12-1.0.1", "Scala": "2.12", "Version": "1.0.1"},
		{"Name": "kafka-2.11-1.0.2", "Scala": "2.11", "Version": "1.0.2"}, {"Name": "kafka-2.12-1.0.2", "Scala": "2.12", "Version": "1.0.2"},
		{"Name": "kafka-2.11-1.1.0", "Scala": "2.11", "Version": "1.1.0"}, {"Name": "kafka-2.12-1.1.0", "Scala": "2.12", "Version": "1.1.0"},
		{"Name": "kafka-2.11-1.1.1", "Scala": "2.11", "Version": "1.1.1"}, {"Name": "kafka-2.12-1.1.1", "Scala": "2.12", "Version": "1.1.1"},
		{"Name": "kafka-2.11-2.0.0", "Scala": "2.11", "Version": "2.0.0"}, {"Name": "kafka-2.12-2.0.0", "Scala": "2.12", "Version": "2.0.0"},
		{"Name": "kafka-2.11-2.0.1", "Scala": "2.11", "Version": "2.0.1"}, {"Name": "kafka-2.12-2.0.1", "Scala": "2.12", "Version": "2.0.1"},
		{"Name": "kafka-2.11-2.1.0", "Scala": "2.11", "Version": "2.1.0"}, {"Name": "kafka-2.12-2.1.0", "Scala": "2.12", "Version": "2.1.0"},
		{"Name": "kafka-2.11-2.1.1", "Scala": "2.11", "Version": "2.1.1"}, {"Name": "kafka-2.12-2.1.1", "Scala": "2.12", "Version": "2.1.1"},
		{"Name": "kafka-2.11-2.2.0", "Scala": "2.11", "Version": "2.2.0"}, {"Name": "kafka-2.12-2.2.0", "Scala": "2.12", "Version": "2.2.0"},
		{"Name": "kafka-2.11-2.2.1", "Scala": "2.11", "Version": "2.2.1"}, {"Name": "kafka-2.12-2.2.1", "Scala": "2.12", "Version": "2.2.1"},
		{"Name": "kafka-2.11-2.2.2", "Scala": "2.11", "Version": "2.2.2"}, {"Name": "kafka-2.12-2.2.2", "Scala": "2.12", "Version": "2.2.2"},
		{"Name": "kafka-2.11-2.3.0", "Scala": "2.11", "Version": "2.3.0"}, {"Name": "kafka-2.12-2.3.0", "Scala": "2.12", "Version": "2.3.0"},
		{"Name": "kafka-2.12-2.3.1", "Scala": "2.12", "Version": "2.3.1"}, {"Name": "kafka-2.11-2.3.1", "Scala": "2.11", "Version": "2.3.1"}, {"Name": "kafka-2.12-2.3.1", "Scala": "2.12", "Version": "2.3.1"},
		{"Name": "kafka-2.11-2.4.0", "Scala": "2.11", "Version": "2.4.0"}, {"Name": "kafka-2.12-2.4.0", "Scala": "2.12", "Version": "2.4.0"}, {"Name": "kafka-2.13-2.4.0", "Scala": "2.13", "Version": "2.4.0"},
		{"Name": "kafka-2.11-2.4.1", "Scala": "2.11", "Version": "2.4.1"}, {"Name": "kafka-2.12-2.4.1", "Scala": "2.12", "Version": "2.4.1"}, {"Name": "kafka-2.13-2.4.1", "Scala": "2.13", "Version": "2.4.1"},
		{"Name": "kafka-2.12-2.5.0", "Scala": "2.12", "Version": "2.5.0"}, {"Name": "kafka-2.13-2.5.0", "Scala": "2.13", "Version": "2.5.0"},
	}
	kafkaService := Service{
		Name:             Kafka,
		UrlTemplate:      "https://archive.apache.org/dist/kafka/{{.Version}}/kafka_{{.Scala}}-{{.Version}}.tgz",
		Versions:         createVersionMap(&kafkaVersions),
		SelectedVersion:  "kafka-2.13-2.5.0",
		InstallationPath: filepath.FromSlash(home + "/.bin" + "/" + Kafka),
		IsEnabled:        true,
	}

	cassandraVersions := []VersionMap{{"Name": "v2.1.21", "Version": "2.1.21"}, {"Name": "v2.2.17", "Version": "2.2.17"}, {"Name": "v3.0.20", "Version": "3.0.20"}, {"Name": "v3.11.7", "Version": "3.11.7"}, {"Name": "v4.0-beta1", "Version": "4.0-beta1"}}
	cassandraService := Service{
		Name:             Cassandra,
		UrlTemplate:      "https://downloads.apache.org/cassandra/{{.Version}}/apache-cassandra-{{.Version}}-bin.tar.gz",
		Versions:         createVersionMap(&cassandraVersions),
		SelectedVersion:  "v3.11.7",
		InstallationPath: filepath.FromSlash(home + "/.bin" + "/" + Cassandra),
		IsEnabled:        true,
	}

	dynamoDbVersions := []VersionMap{{"Name": "Asia_Pacific_(Mumbai)_Region", "Region": "ap-south-1", "RegionName": "-mumbai"}, {"Name": "Asia_Pacific_(Singapore)_Region", "Region": "ap-southeast-1", "RegionName": "-singapore"},
		{"Name": "Asia_Pacific_(Tokyo)_Region", "Region": "ap-northeast-1", "RegionName": "-tokyo"}, {"Name": "Europe_(Frankfurt)_Region", "Region": "eu-central-1", "RegionName": "-frankfurt"},
		{"Name": "South_America_(São_Paulo)_Region", "Region": "sa-east-1", "RegionName": "-sao-paulo"}, {"Name": "US_West_(Oregon)_Region", "Region": "us-west-2", "RegionName": ""}}
	dynamoDbService := Service{
		Name:             DynamoDb,
		UrlTemplate:      "https://s3.{{.Region}}.amazonaws.com/dynamodb-local{{.RegionName}}/dynamodb_local_latest.tar.gz",
		Versions:         createVersionMap(&dynamoDbVersions),
		SelectedVersion:  "US_West_(Oregon)_Region",
		InstallationPath: filepath.FromSlash(home + "/.bin" + "/" + DynamoDb),
		IsEnabled:        true,
	}

	services := make(map[string]Service)
	services[Kafka] = kafkaService
	services[Cassandra] = cassandraService
	services[DynamoDb] = dynamoDbService

	configuration := Configuration{
		Info:     "Customize below Configuration to point to an internal url, versions or disable any service. When using internal URL with version, make sure url template is valid",
		Services: services,
	}

	return configuration
}

// TODO: Remove the use of Name and use the key directly
func createVersionMap(versions *[]VersionMap) map[string]VersionMap {
	versionMap := make(map[string]VersionMap)
	for _, version := range *versions {
		versionKey := ""
		for key, val := range version {
			if key == "Name" {
				versionKey = val
			}
		}
		versionMap[versionKey] = version
	}
	return versionMap
}
