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
	"bytes"
	"com.github/RawSanj/setup/util"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

const AllKey = "all"

type UrlHolder struct {
	Version string
	Scala   string
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install all or select services which you want to be installed",
	Long: `install all or select services which you want to be installed
Currently supports various versions of Kafka, Cassandra and DynamoDb.
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		applicationConfiguration, err := initializeApplicationConfiguration()
		if err != nil {
			return err
		}

		acceptDefaultAndInstall, defaultPromptErr := acceptDefaultAndInstallAllPrompt(&applicationConfiguration)
		if defaultPromptErr != nil {
			return defaultPromptErr
		}

		var servicesToInstall []string

		if acceptDefaultAndInstall {
			servicesToInstall = defaultServicesToInstall(&applicationConfiguration)
		} else {
			servicesToInstall, err = chooseServicesToInstall(&applicationConfiguration)
			if err != nil {
				return err
			}
		}

		for _, selectedSvc := range servicesToInstall {
			err := downloadAndExtract(&applicationConfiguration, selectedSvc)
			if err != nil {
				fmt.Println("Error installing service", selectedSvc, "Error: ", err.Error())
			}
		}

		configurationString, err := marshalConfiguration(&applicationConfiguration)
		if err != nil {
			return err
		}
		viper.Set(ConfigurationKey, configurationString)
		err = viper.WriteConfig()
		if err != nil {
			return errors.New("Error Writing Config File. Error: " + err.Error())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

// Read Configuration from $HOME/.setup.yml and marshall & set into applicationConfiguration
func initializeApplicationConfiguration() (Configuration, error) {
	configurationString := viper.GetString(ConfigurationKey)
	applicationConfiguration := Configuration{}
	if configurationString == "" {
		return applicationConfiguration, errors.New("configuration initialization failed")
	}
	err := unMarshalConfiguration(configurationString, &applicationConfiguration)
	if err != nil {
		return applicationConfiguration, err
	}
	return applicationConfiguration, nil
}

func acceptDefaultAndInstallAllPrompt(applicationConfiguration *Configuration) (bool, error) {

	availableServices := make([]string, 0, len(applicationConfiguration.Services))

	for key, service := range applicationConfiguration.Services {
		if service.IsEnabled {
			availableServices = append(availableServices, key+":"+service.SelectedVersion)
		}
	}

	acceptDefault := false
	acceptDefaultPrompt := &survey.Confirm{
		Message: "Accept Default and install All Available Services?",
		Help:    fmt.Sprintf("Available Services to be installed: %s", availableServices),
	}
	acceptDefaultErr := survey.AskOne(acceptDefaultPrompt, &acceptDefault)

	if acceptDefaultErr != nil {
		return acceptDefault, acceptDefaultErr
	}
	return acceptDefault, nil
}

func defaultServicesToInstall(applicationConfiguration *Configuration) []string {

	servicesToInstall := make([]string, 0, len(applicationConfiguration.Services))
	for key, service := range applicationConfiguration.Services {
		if service.IsEnabled {
			servicesToInstall = append(servicesToInstall, key)
		}
	}

	return servicesToInstall
}

func chooseServicesToInstall(applicationConfiguration *Configuration) ([]string, error) {

	selectedServices, servicesPromptErr := selectServices(applicationConfiguration)
	if servicesPromptErr != nil {
		return nil, servicesPromptErr
	}

	if installAllKeyExists, _ := util.HasElement(selectedServices, AllKey); installAllKeyExists {
		selectedServices = defaultServicesToInstall(applicationConfiguration)
	}

	versionPromptErr := selectAndSetServiceVersion(applicationConfiguration, selectedServices)
	if versionPromptErr != nil {
		return nil, versionPromptErr
	}

	return selectedServices, nil
}

func selectServices(applicationConfiguration *Configuration) ([]string, error) {

	availableServices := make([]string, 0, len(applicationConfiguration.Services)+1)
	availableServices = append(availableServices, AllKey)

	for key, service := range applicationConfiguration.Services {
		if service.IsEnabled {
			availableServices = append(availableServices, key)
		}
	}

	var selectedServices []string

	selectedServicesPrompt := &survey.MultiSelect{
		Message:  "Select Services to be installed",
		Help:     fmt.Sprintf("Select one or more service from: %s to install", availableServices),
		Options:  availableServices,
		PageSize: len(availableServices),
	}

	selectedServicesErr := survey.AskOne(selectedServicesPrompt, &selectedServices, survey.WithValidator(survey.Required), survey.WithPageSize(10))
	if selectedServicesErr != nil {
		return nil, selectedServicesErr
	}

	return selectedServices, nil
}

func selectAndSetServiceVersion(applicationConfiguration *Configuration, selectedServices []string) error {

	for _, selectedSvc := range selectedServices {
		version := ""
		selectedService := applicationConfiguration.Services[selectedSvc]
		versionList := getVersionList(&selectedService.Versions)
		selectVersionPrompt := &survey.Select{
			Message:  fmt.Sprintf("Select %s version to install", selectedService.Name),
			Help:     fmt.Sprintf("[%s] versions available for: %s", len(versionList), selectedService.Name),
			Options:  versionList,
			PageSize: 10,
		}

		err := survey.AskOne(selectVersionPrompt, &version)
		if err != nil {
			return err
		}
		selectedService.SelectedVersion = addUnderScore(version)
		applicationConfiguration.Services[selectedSvc] = selectedService
	}
	return nil
}

func getVersionList(versionMap *map[string]VersionMap) []string {

	versions := make([]string, 0, len(*versionMap))
	for key, _ := range *versionMap {
		versions = append(versions, removeUnderScore(key))
	}
	return versions
}

func removeUnderScore(key string) string {
	return strings.ReplaceAll(key, "_", " ")
}

func addUnderScore(key string) string {
	return strings.ReplaceAll(key, " ", "_")
}

func downloadAndExtract(applicationConfiguration *Configuration, selectedService string) error {

	service := applicationConfiguration.Services[selectedService]

	var url = service.UrlTemplate

	if strings.Contains(url, "{{") {
		t, err := template.New("UrlTemplate").Parse(url)
		if err != nil {
			fmt.Println("Error Parsing UrlTemplate for Service", service.Name, "Please correct the url configuration")
			return err
		}

		var result bytes.Buffer

		err = t.Execute(&result, service.Versions[service.SelectedVersion])
		if err != nil {
			fmt.Println("Error Parsing UrlTemplate for Service", service.Name, "Please correct the url configuration")
			return err
		}

		url = result.String()
	}

	folderErr := os.MkdirAll(service.InstallationPath, 0755)
	if folderErr != nil {
		fmt.Println("Error Creating Installation Directory. Please select proper installation path", "Error is: ", folderErr.Error())
		return folderErr
	}

	downloadedFilePath, err := util.DownloadFile(service.InstallationPath, url)
	if err != nil {
		fmt.Println("Error Downloading Service", service.Name, "Error is: ", err.Error())
		return err
	}

	extractErr := util.ExtractTarGz(downloadedFilePath, filepath.FromSlash(service.InstallationPath+"/"+service.SelectedVersion))
	_ = os.Remove(downloadedFilePath)
	if extractErr != nil {
		return extractErr
	}

	updateConfigAfterInstallation(service, applicationConfiguration)

	return nil
}

func updateConfigAfterInstallation(service Service, applicationConfiguration *Configuration) {
	if service.ActiveVersion == "" {
		service.ActiveVersion = service.SelectedVersion
	}

	service.InstalledVersion = append(service.InstalledVersion, service.SelectedVersion)
	applicationConfiguration.Services[service.Name] = service
}
